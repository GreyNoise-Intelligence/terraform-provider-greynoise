# Terraform Greynoise Provider

The GreyNoise Provider enables Terraform to manage GreyNoise sensors and personas.

* [Contributing guide]()
* [FAQ]()
* [Tutorials]()

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

### Local Development/Testing

This repo provides a Docker-based environment so you don't have to install Terraform on your laptop.
This environment can be used to run the [examples/aws](examples/aws). To get started perform the following
actions:

1. Build the provider for the docker environment:

```shell
make docker-build
```

2. Set your environment by setting the GreyNoise API key as an environment variable:

```shell 
export GN_API_KEY=<API_KEY>
```

3. Run the interactive Docker environment:

```shell 
make docker-shell 
```

4. Update the Terraform variables file `terraform.tfvars`. See the `variables.tf` file for explanation of variables.
   You need access to the GreyNoise sensors feature to run this example.

```shell
terraform init 
terraform plan 
```
