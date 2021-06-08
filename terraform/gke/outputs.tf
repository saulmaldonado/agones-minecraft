output "cluster_ca_certificate" {
  value = base64decode(google_container_cluster.primary.master_auth.0.cluster_ca_certificate)
}

output "host" {
  value = "https://${google_container_cluster.primary.endpoint}"
}

output "token" {
  value     = data.google_client_config.default.access_token
  sensitive = true

}

output "name_servers" {
  value = google_dns_managed_zone.dns.name_servers
}
