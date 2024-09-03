default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

docker-shell:
	make -C docker shell

docker-install:
	mkdir -p .terraform.d/plugins/registry.terraform.io/hashicorp/greynoise/1.0.0/linux_arm64
	GOOS=linux GOARCH=amd64 go build -o terraform-provider-greynoise_v1.0.0 .
	mv terraform-provider-greynoise_v1.0.0 .terraform.d/plugins/registry.terraform.io/hashicorp/greynoise/1.0.0/linux_arm64

generate:
	go generate
