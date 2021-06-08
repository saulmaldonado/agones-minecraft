# Cloud stroage bucket for minecraft world saves
resource "google_storage_bucket" "storage" {
  name          = "agones-minecraft-mc-worlds-${random_id.bucket.hex}"
  location      = var.region
  force_destroy = true
}

# Unique suffix for globally unique bucket name
resource "random_id" "bucket" {
  byte_length = 8
}
