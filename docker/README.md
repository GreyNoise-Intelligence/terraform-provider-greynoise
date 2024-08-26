# Terraform Docker Shell

This folder provides a simple Docker environment for doing local development with Terragrunt and Terraform. Run the
following to shell into the Docker environment:

```
make shell 
```

This command will build an image with these versions installed:

```
ENV TG_VERSION=0.53.0
ENV TF_VERSION=1.6.6
```

See [compatibility matrix](https://terragrunt.gruntwork.io/docs/getting-started/supported-versions/).

It will mount the following directories :

* `~/.aws directory` for AWS credentials; in addition it will propagate `AWS*` environment variables.
* Github repo root to `/workspace` in the container

