resource "google_storage_bucket" "storage" {
  name          = "agones-minecraft-mc-worlds"
  force_destroy = true
}
