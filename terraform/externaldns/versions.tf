terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.3"
    }
  }
  required_version = ">= 1.0.0"
}
