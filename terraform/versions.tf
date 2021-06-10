terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.70.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.3"
    }

    helm = {
      version = "~> 2.1"
      source  = "hashicorp/helm"
    }

  }
  required_version = ">= 1.0.0"
}
