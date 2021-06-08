variable "agones_version" {
  description = "agones version"
}

variable "host" {}

variable "token" {}

variable "cluster_ca_certificate" {}

provider "helm" {
  kubernetes {
    host                   = var.host
    token                  = var.token
    cluster_ca_certificate = var.cluster_ca_certificate
  }
}

resource "helm_release" "agones" {
  name             = "agones"
  repository       = "https://agones.dev/chart/stable"
  chart            = "agones"
  version          = var.agones_version
  namespace        = "agones-system"
  create_namespace = true


  set {
    name  = "agones.crds.CleanupOnDelete"
    value = true
  }

  set {
    name  = "agones.ping.udp.expose"
    value = false
  }

  set {
    name  = "agones.allocator.install"
    value = false
  }

}
