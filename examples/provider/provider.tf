terraform {
  required_providers {
    greynoise = {
      source  = "hashicorp/greynoise"
      version = "0.1.0"
    }
  }
}

provider "greynoise" {
  // GN_API_KEY env var is used to provide key
}
