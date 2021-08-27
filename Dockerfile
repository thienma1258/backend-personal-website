ARG DOCKER_IMAGE_TAG="alpine"


# ===== Builder =====
FROM golang:${DOCKER_IMAGE_TAG} as builder

WORKDIR /app

RUN \
  apk add \
  --no-cache \
  --update \
  --virtual "shared-dependencies" \
  "bash" \
  "bc" \
  "curl" \
  "jq"

# Core
COPY "./.bin" "./.bin"
COPY "./lib/bash/core.sh" "./lib/bash/core.sh"
COPY "./lib/bash/go.sh" "./lib/bash/go.sh"

RUN \
  apk add \
  --virtual "build-dependencies" \
  "git"



# Install
COPY "./pipeline/install" "./pipeline/install"
COPY "./go.mod" "./go.mod"
COPY "./go.sum" "./go.sum"
COPY "./submodules/mrlibs" "./submodules/mrlibs"
COPY "./src" "./src"
RUN ./pipeline/install

# Build
COPY "./pipeline/build" "./pipeline/build"
COPY "./project.config.json" "./project.config.json"
#RUN go get -u github.com/swaggo/swag/cmd/swag
#RUN ./pipeline/build


# ===== Production =====
FROM golang:${DOCKER_IMAGE_TAG}

WORKDIR /app

RUN \
  apk add \
  --no-cache \
  --update \
  --virtual "shared-dependencies" \
  "bash" \
  "bc" \
  "curl" \
  "jq"

# Core
COPY "./.bin" "./.bin"
COPY "./lib/bash/core.sh" "./lib/bash/core.sh"
COPY "./lib/bash/go.sh" "./lib/bash/go.sh"

# EnvKey
RUN bash -c "source ./lib/bash/core.sh; dependency envkey-source"


# Run
COPY "./pipeline/run" "./pipeline/run"
COPY "./project.config.json" "./project.config.json"

# Built
COPY --from=builder "/app/build" "./build"
COPY --from=builder "/app/submodules" "./submodules"

EXPOSE 8088

ENTRYPOINT [ \
  "./pipeline/run" \
  ]
