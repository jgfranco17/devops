# ========== CLI BUILD STAGE ==========

ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod/ \
    CGO_ENABLED=0 go build -o /bin/devops .

# ========== CLI RUN STAGE ==========

FROM ubuntu:22.04 AS app

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates xz-utils && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /bin/devops /bin/

# Tests should use non-root to simulate a real user.
ARG USERNAME=testuser
ARG USER_ID=1000
ARG GROUP_ID=1000
RUN addgroup --gid $GROUP_ID ${USERNAME} && \
    adduser --disabled-password --gecos '' --uid ${USER_ID} --gid ${GROUP_ID} ${USERNAME}

USER $USERNAME
ENV HOME=/home/${USERNAME}
WORKDIR ${HOME}/projects
COPY --chown="${USERNAME}:${USERNAME}" integration/sample.yaml "${HOME}/projects/devops-definition.yaml"

ENTRYPOINT [ "/bin/devops" ]
CMD ["--help"]
