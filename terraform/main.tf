module "gke_cluster" {
  source = "./gke"

  project_id       = var.project_id
  region           = var.region
  zone             = var.zone
  dns_name         = var.dns_name
  cluster_version  = var.cluster_version
  storage_location = var.storage_location
  auto_scaling     = var.auto_scaling
  min_node_count   = var.min_node_count
  max_node_count   = var.max_node_count
}

module "helm_agones" {
  source = "./agones-helm"

  agones_version = var.agones_version

  host                   = module.gke_cluster.host
  token                  = module.gke_cluster.token
  cluster_ca_certificate = module.gke_cluster.cluster_ca_certificate
}

module "external_dns" {
  source = "./externaldns"

  host                   = module.gke_cluster.host
  token                  = module.gke_cluster.token
  cluster_ca_certificate = module.gke_cluster.cluster_ca_certificate
}

output "name_servers" {
  description = "name servers for DNS zone"
  value       = module.gke_cluster.name_servers
}
