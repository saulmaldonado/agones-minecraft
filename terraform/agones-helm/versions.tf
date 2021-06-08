terraform {
  required_version = "~> 0.15"
  required_providers {
    helm  = {
      version = "~> 2.1"
      source  = "hashicorp/helm"
    }
  }
}
