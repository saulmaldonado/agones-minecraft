variable "project_id" {
  description = "gcp project id"
}

variable "dns_name" {
  description = "managed dns zone name"
}

variable "region" {
  description = "gcp region"
}

variable "zone" {
  description = "gcp zone"
}

variable "cluster_version" {
  description = "gke cluster version"
}

variable "agones_version" {
  description = "agones version"
}

module "gke_cluster" {
  source = "./gke"

  project_id      = var.project_id
  region          = var.region
  zone            = var.zone
  dns_name        = var.dns_name
  cluster_version = var.cluster_version
}

module "helm_agones" {
  source                 = "./agones-helm"

  agones_version         = var.agones_version

  host                   = module.gke_cluster.host
  token                  = module.gke_cluster.token
  cluster_ca_certificate = module.gke_cluster.cluster_ca_certificate
}
