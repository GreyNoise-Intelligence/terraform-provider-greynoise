# Args for FROM stage
ARG TF_VERSION=1.9.1

# Available images: https://hub.docker.com/r/alpine/terragrunt/tags
FROM alpine/terragrunt:$TF_VERSION

# Args for post-FROM stage
ARG TG_VERSION=0.60.0

ENV AWS_SDK_LOAD_CONFIG=true
ENV AWS_DEFAULT_OUTPUT=json
ENV AWS_DEFAULT_REGION=us-east-1
# enable non-interative mode
ENV TF_INPUT=0

RUN wget -O /usr/local/bin/terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v${TG_VERSION}/terragrunt_linux_amd64 && \
    chmod +x /usr/local/bin/terragrunt

RUN apk add --no-cache --virtual \
    build-base \
    zip \
    make \
    bash \
    gcc \
    musl-dev \
    openssl \
    python3 \
    python3-dev \
    py3-pip \
    curl \
    libffi-dev \
    openssl-dev \
    jq \
    aws-cli

RUN mkdir -p /workspace /root/.aws /root/.terraform.d/plugins
ADD .docker-prompt /root/.profile

WORKDIR /workspace
