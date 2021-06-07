variable "cluster_name" {
  default = "miencraft"
  description = "cluster name"
}

variable "gke_num_nodes" {
  default     = 2
  description = "number of gke nodes"
}

# GKE cluster
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.region

  min_master_version = "1.18"

  node_pool {
    name = "default"
    node_count = var.gke_num_nodes
    version = "1.18"
    management {
      auto_upgrade = false
    }

    node_config {
      oauth_scopes = [
        "gke-default",
        "storage-rw",
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
