variable "dns_name" {
  description = "dns name"
}

resource "google_dns_managed_zone" "dns" {
  name        = "agones-minecraft"
  dns_name    = var.dns_name
  description = "externalDNS controller managed DNS zone"
}
