resource "google_compute_firewall" "filewall" {
  name    = "mc-server-firewall"
  network = google_compute_network.vpc.name

  allow {
    protocol = "tcp"
    ports    = ["7000-8000"]
  }

  allow {
    protocol = "udp"
    ports    = ["7000-8000"]
  }

  target_tags = ["mc"]
}
