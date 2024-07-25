default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

build:
	mkdir -p .terraform.d/plugins/registry.terraform.io/greynoise-io/greynoise/1.0.0/linux_arm64
	go build -o terraform-provider-greynoise_v1.0.0 .
	mv terraform-provider-greynoise_v1.0.0 .terraform.d/plugins/registry.terraform.io/greynoise-io/greynoise/1.0.0/linux_arm64
