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

variable "storage_location" {
  description = "gcs bucket location"
}

variable "cluster_version" {
  description = "gke cluster version"
}

variable "agones_version" {
  description = "agones version"
}

variable "auto_scaling" {
  type        = bool
  default     = false
  description = "enable cluster node auto scaling"
}

variable "min_node_count" {
  type        = number
  description = "minimum node count for auto scaling"
}

variable "max_node_count" {
  type        = number
  description = "maximum node count for auto scaling"
}
