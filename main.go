package main

import (
	"context"
	"egtyl.xyz/omnibill/linux"
	"encoding/json"
	"errors"
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v71/github"
	"github.com/joho/godotenv"
	"github.com/kr/pretty"
	"io/fs"
	"os"
	"path/filepath"
	"shibidev.xyz/pterodactyl/musebot/logger"
	"strings"
)

const MUSEBOT_GIT = "https://github.com/museofficial/muse.git"

var WORK_DIR string
var HOME_DIR = os.Getenv("HOME")
var REMOTE_NAME = "origin"
var REF_NAME plumbing.ReferenceName

var firstInit = false

var appLog = logger.New(logger.Options{
	Prefix: "MCST",
})

var envMap = make(map[string]string)

var (
	yarnInstallCmd *linux.LinuxCommand
	prismaGenCmd   *linux.LinuxCommand
	buildCmd       *linux.LinuxCommand
)

func init() {

	var err error
	WORK_DIR, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	if WORK_DIR != HOME_DIR {
		HOME_DIR = filepath.Join(WORK_DIR, HOME_DIR)
	}

	for _, envVar := range os.Environ() {
		splitVar := strings.SplitN(envVar, "=", 2)
		envMap[splitVar[0]] = splitVar[1]
	}

	envFile, err := os.OpenFile(filepath.Join(HOME_DIR, "muse_cfg.env"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	envFileMap, err := godotenv.Parse(envFile)
	if err != nil {
		panic(err)
	}

	for k, v := range envFileMap {
		if varValue, varExists := envMap[k]; varExists {
			pretty.Println(varValue)
			if len(varValue) == 0 {
				envMap[k] = v
			}
		}
		envMap[k] = v
	}

	commandOptions := linux.CommandOptions{
		Env:     envMap,
		Cwd:     filepath.Join(HOME_DIR, "muse"),
		Command: "yarn",
	}

	yarnInstallCmdOpts := &commandOptions
	yarnInstallCmdOpts.Args = []string{"install", "--production=false"}
	yarnInstallCmdOpts.PrintOutput = true
	yarnInstallCmd, err = linux.NewCommand(*yarnInstallCmdOpts)
	if err != nil {
		panic(err)
	}

	prismaGenCmdOpts := &commandOptions
	prismaGenCmdOpts.Args = []string{"prisma", "generate"}
	prismaGenCmd, err = linux.NewCommand(*prismaGenCmdOpts)
	if err != nil {
		panic(err)
	}

	buildCmdOpts := &commandOptions
	buildCmdOpts.Args = []string{"build"}
	buildCmd, err = linux.NewCommand(*buildCmdOpts)
	if err != nil {
		panic(err)
	}

}

func main() {
	githubAPI := github.NewClient(nil)

	release, _, err := githubAPI.Repositories.GetLatestRelease(context.Background(), "museofficial", "muse")
	if err != nil {
		return
	}

	gitReleaseVer, err := semver.NewVersion(*release.TagName)
	if err != nil {
		panic(err)
	}

	REF_NAME = plumbing.NewTagReferenceName("v" + gitReleaseVer.String())

	if err := os.Mkdir(filepath.Join(HOME_DIR, "muse"), 0755); err != nil && !errors.Is(err, fs.ErrExist) {
		panic(err)
	}

	var gitRepo *git.Repository

	gitRepo, err = git.PlainOpen(filepath.Join(HOME_DIR, "muse"))
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {

			appLog.Info("Grabbing latest Muse release from GitHub.")

			firstInit = true
			gitRepo, err = git.PlainClone(filepath.Join(HOME_DIR, "muse"), false, &git.CloneOptions{
				URL:           MUSEBOT_GIT,
				Depth:         1,
				RemoteName:    REMOTE_NAME,
				ReferenceName: REF_NAME,
				Tags:          git.AllTags,
			})
			if err != nil {

				switch {
				case errors.Is(err, git.ErrRepositoryAlreadyExists):
					appLog.Error("Installation Failed: An untracked repository exists, please delete the \033[1m\"muse\"\033[0m directory.")
					break
				default:
					appLog.Error("Installation Failed: An error occurred while trying to install Muse. Please report this to MCST Support Staff.")
					break
				}
				os.Exit(1)
			}

			if err := os.Symlink(filepath.Join(HOME_DIR, "muse", "README.md"), filepath.Join(HOME_DIR, "README.md")); err != nil && !errors.Is(err, fs.ErrExist) {
				panic(err)
			}

		} else {
			panic(err)
		}
	}

	packageJsonFile, err := os.ReadFile(filepath.Join(HOME_DIR, "muse", "package.json"))
	if err != nil {
		panic(err)
	}

	var packageJson map[string]interface{}
	if err := json.Unmarshal(packageJsonFile, &packageJson); err != nil {
		panic(err)
	}

	packageJsonVer, err := semver.NewVersion(packageJson["version"].(string))
	if err != nil {
		panic(err)
	}

	worktree, err := gitRepo.Worktree()
	if err != nil {
		panic(err)
	}

	if packageJsonVer.LessThan(gitReleaseVer) {
		appLog.Warn("New Update Found: " + gitReleaseVer.String())
		appLog.Info("Downloading Update...")

		if err := gitRepo.Fetch(&git.FetchOptions{
			RemoteName: REMOTE_NAME,
			Tags:       git.AllTags,
		}); err != nil {
			panic(err)
		}

		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName("v" + gitReleaseVer.String()),
		}); err != nil {
			panic(err)
		}

		appLog.Info("Upgraded Muse: " + "v" + packageJsonVer.String() + " -> " + "v" + gitReleaseVer.String())

	} else if packageJsonVer.Equal(gitReleaseVer) && !firstInit {
		appLog.Info("No updates found.")
	} else {
		if err := gitRepo.Fetch(&git.FetchOptions{
			RemoteName: REMOTE_NAME,
			Tags:       git.AllTags,
		}); err != nil {
			panic(err)
		}

		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName("v" + gitReleaseVer.String()),
		}); err != nil {
			panic(err)
		}
	}

	appLog.Info("Running 'yarn install'...")

	if err := yarnInstallCmd.Run(); err != nil {
		panic(err)
	}

	appLog.Info("Generating PrismaORM Client...")

	if err := prismaGenCmd.Run(); err != nil {
		panic(err)
	}

	appLog.Info("Building Application...")
	//buildCmd.Options.Env["PATH"] = buildCmd.Options.Env["PATH"] + ":" + filepath.Join(HOME_DIR, "muse", "node_modules", ".bin")

	if err := buildCmd.Run(); err != nil {
		panic(err)
	}

	runCmd, err := linux.NewCommand(linux.CommandOptions{
		Env:         envMap,
		Cwd:         filepath.Join(HOME_DIR, "muse"),
		Command:     "node",
		Args:        []string{"--enable-source-maps", "dist/scripts/migrate-and-start.js"},
		PrintOutput: true,
	})
	if err != nil {
		panic(err)
	}

	_ = runCmd.Run() // If an error happens, it's probably a user error.

}
