#
# Common Makefile for building and publishing ECR images.
#
# Expects the following to be defined upstream
# - DOCKER_IMAGE: name of the docker image - also name of the AWS ECR
# - DOCKER_FILE_PATH: Path for Dockerfile used to build the image
# - DOCKER_CONTEXT_PATH: Path of Docker context
# - DOCKER_ARGS: Args for running container test
#
# In addition to the variables above, also expected AWS credentials to be configured
#
GIT_COMMIT_HASH  ?= $(shell git rev-parse HEAD)
DOCKER_IMAGE_TAG := $(DOCKER_IMAGE):$(GIT_COMMIT_HASH)

DOCKER_DEV_REGISTRY   := 683356862652.dkr.ecr.us-east-1.amazonaws.com/$(DOCKER_IMAGE)
DOCKER_PROD_REGISTRY  := 676188196945.dkr.ecr.us-east-1.amazonaws.com/$(DOCKER_IMAGE)

DOCKER_ARGS       ?=
DOCKER_BUILD_ARGS ?=
DOCKER_REVERT_TAG ?= $(GIT_COMMIT_HASH)
DOCKER_REGISTRY   := $(DOCKER_DEV_REGISTRY)

TF_ENVIRONMENT ?= development
ifeq ($(TF_ENVIRONMENT), production)
	DOCKER_REGISTRY = $(DOCKER_PROD_REGISTRY)
endif

.PHONY: build-image
build-image:
	@docker build $(DOCKER_BUILD_ARGS) -t $(DOCKER_IMAGE_TAG) -f $(DOCKER_FILE_PATH)/Dockerfile --pull $(DOCKER_CONTEXT_PATH)

run-image:
	@docker run $(DOCKER_ARGS) $(DOCKER_IMAGE_TAG)

shell-image:
	@docker run -it $(DOCKER_ARGS) --entrypoint /bin/bash $(DOCKER_IMAGE_TAG) --login

.PHONY: publish-image-commit
publish-image-commit:
	# expects AWS credentials to be configured
	@aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(DOCKER_REGISTRY)
	@docker tag $(DOCKER_IMAGE_TAG) $(DOCKER_REGISTRY):$(GIT_COMMIT_HASH)
	@docker push $(DOCKER_REGISTRY):$(GIT_COMMIT_HASH)

.PHONY: publish-image-latest
publish-image-latest:
	@docker tag $(DOCKER_IMAGE_TAG) $(DOCKER_REGISTRY):latest
	@docker push $(DOCKER_REGISTRY):latest

.PHONY: publish-image
ifeq ($(SKIP_LATEST_TAG), true)
publish-image: build-image publish-image-commit
	@echo 'Skipping latest image tag'
else
publish-image: build-image publish-image-commit publish-image-latest
endif

# target is used to update the `latest` tag to a known ECR tag
.PHONY: revert-image
revert-image:
	# expects AWS credentials to be configured
	@aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(DOCKER_REGISTRY)
	@docker pull $(DOCKER_REGISTRY):$(DOCKER_REVERT_TAG)
	@docker tag $(DOCKER_REGISTRY):$(DOCKER_REVERT_TAG) $(DOCKER_REGISTRY):latest
	@docker push $(DOCKER_REGISTRY):latest

.PHONY: pull-image
pull-image:
	# expects AWS credentials to be configured
	@aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(DOCKER_REGISTRY)
	@docker pull $(DOCKER_REGISTRY):latest
	@docker tag $(DOCKER_REGISTRY):latest $(DOCKER_IMAGE_TAG)

.PHONY: clean-images
clean-images:
	@docker image rm $(DOCKER_IMAGE_TAG) &> /dev/null|| true
	@docker image rm $(DOCKER_REGISTRY):$(GIT_COMMIT_HASH) &> /dev/null|| true
	@docker image rm $(DOCKER_REGISTRY):latest &> /dev/null|| true
