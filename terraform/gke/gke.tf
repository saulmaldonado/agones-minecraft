# Cluster name
variable "cluster_name" {
  default     = "minecraft"
  description = "cluster name"
}

# Total number of cluster nodes
variable "gke_num_nodes" {
  default     = 2
  description = "number of gke nodes"
}

# Zone for zonal cluster
variable "zone" {
  default     = ""
  description = "cluster zone"
}

variable "cluster_version" {
  default     = "1.18"
  description = "gke cluster version"
}

variable "auto_scaling" {
  type = bool
  default = false
  description = "enable cluster node auto scaling"
}

variable "min_node_count" {
  type = number
  description = "minimum node count for auto scaling"
}

variable "max_node_count" {
  type = number
  description = "maximum node count for auto scaling"
}

data "google_client_config" "default" {}

provider "google" {
  project = var.project_id
  region  = var.region
}

# GKE cluster
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.zone

  min_master_version = var.cluster_version

  node_pool {
    name       = "default"
    node_count = var.gke_num_nodes
    version    = var.cluster_version

    management {
      auto_upgrade = false
    }

    dynamic "autoscaling" {
      for_each = var.auto_scaling ? [1] : []
      content {
        max_node_count = var.max_node_count
        min_node_count = var.min_node_count
      }
    }

    node_config {
      oauth_scopes = [
        "https://www.googleapis.com/auth/devstorage.read_only",
        "https://www.googleapis.com/auth/logging.write",
        "https://www.googleapis.com/auth/monitoring",
        "https://www.googleapis.com/auth/service.management.readonly",
        "https://www.googleapis.com/auth/servicecontrol",
        "https://www.googleapis.com/auth/trace.append",
        "https://www.googleapis.com/auth/devstorage.read_write",
        "https://www.googleapis.com/auth/ndev.clouddns.readwrite"
      ]

      machine_type = "n2-standard-4"
      tags         = ["mc"]

      metadata = {
        disable-legacy-endpoints = "true"
      }
    }
  }

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
}
