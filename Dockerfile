FROM golang:1.24.2-bookworm AS go_builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o build/musebot_kickstart

FROM node:22-bookworm-slim AS base

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    ffmpeg \
    tini \
    openssl \
    ca-certificates \
    python3 \
    make \
    gcc \
    g++

RUN apt-get autoclean && \
    apt-get autoremove && \
    rm -rf /var/lib/apt/lists/*


FROM base AS runner

RUN addgroup container && \
	useradd -m -d /home/container -g container -s /bin/bash container

COPY --from=go_builder /app/build/musebot_kickstart /usr/local/bin/musebot_kickstart
COPY --chown=container:container muse_cfg.env /home/container/muse_cfg.env
WORKDIR /home/container

USER container
ENV USER=container HOME=/home/container

ARG BUILD_DATE=unknown
ARG COMMIT_HASH=unknown

ENV COMMIT_HASH=$COMMIT_HASH
ENV BUILD_DATE=$BUILD_DATE

ENV NODE_ENV=production

ENV ENV_FILE=/home/container/muse_cfg.env
ENV DATA_DIR=/home/container/data

ENTRYPOINT ["tini", "--", "/usr/local/bin/musebot_kickstart"]
