terraform {
  required_providers {
    helm = {
      version = "~> 2.1"
      source  = "hashicorp/helm"
    }
  }
  required_version = ">= 1.0.0"
}
