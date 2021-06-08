variable "host" {}

variable "token" {}

variable "cluster_ca_certificate" {}


provider "kubernetes" {
  host                   = var.host
  token                  = var.token
  cluster_ca_certificate = var.cluster_ca_certificate
}

resource "kubernetes_deployment" "externaldns" {
  metadata {
    name = "external-dns"
  }

  spec {
    selector {
      match_labels = {
        app = "external-dns"
      }
    }

    template {
      metadata {
        labels = {
          app = "external-dns"
        }
      }

      spec {
        service_account_name = "external-dns"
        container {
          image             = "saulmaldonado/external-dns"
          name              = "external-dns"
          args              = ["--source=gameserver", "--provider=google", "--registry=txt", "--txt-owner-id=external-dns-controller"]
          image_pull_policy = "Always"
        }
      }
    }
  }
}

resource "kubernetes_service_account" "externaldns" {
  metadata {
    name = "external-dns"
  }
}

resource "kubernetes_cluster_role" "externaldns" {
  metadata {
    name = "external-dns"
  }

  rule {
    api_groups = [""]
    resources  = ["services", "endpoints", "pods"]
    verbs      = ["get", "watch", "list"]
  }

  rule {
    api_groups = ["extensions", "networking.k8s.io"]
    resources  = ["ingresses"]
    verbs      = ["get", "watch", "list"]
  }

  rule {
    api_groups = [""]
    resources  = ["nodes"]
    verbs      = ["get", "watch", "list"]
  }

  rule {
    api_groups = ["agones.dev"]
    resources  = ["gameservers"]
    verbs      = ["get", "watch", "list"]
  }
}


resource "kubernetes_cluster_role_binding" "externaldns" {
  metadata {
    name = "external-dns-viewer"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "external-dns"
  }
  subject {
    kind      = "ServiceAccount"
    name      = "external-dns"
    namespace = "default"
  }
}
