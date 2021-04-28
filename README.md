# Installation

## Prerequisites

- [gcloud](https://cloud.google.com/sdk/docs/install)
- [kubectl](https://kubernetes.io/docs/tasks/tools/included/install-kubectl-gcloud/)

## 1. Create public DNS Zone

```sh
gcloud dns managed-zones create agones-minecraft \ # Any name
    --description="agones-mc-dns-controller managed DNS zone"\
    --dns-name=<DOMAIN> \ # Domain that you own
    --visibility=public
```

Point domain to Google nameservers

```sh
gcloud dns managed-zones describe agones-minecraft
```

## 2. Create Kubernetes Cluster

```sh
gcloud container clusters create minecraft --cluster-version=1.18 \
  --tags=mc \
  --scopes=gke-default,"https://www.googleapis.com/auth/ndev.clouddns.readwrite" \ # GKE scope needed for Cloud DNS
  --node-labels=agones-mc/<DOMAIN_NAME> # Replace with the domain for the zone that the controller will manage
  --num-nodes=2 \
  --no-enable-autoupgrade \
  --machine-type=n2-standard-4
```

```sh
gcloud config set container/cluster minecraft
gcloud container clusters get-credentials minecraft
```

```sh
gcloud compute firewall-rules create mc-server-firewall \
  --allow tcp:7000-8000 \
  --target-tags mc \
  --description "Firewall rule to allow mc server tcp traffic"
```

## 3. Install Agones

```sh
kubectl create namespace agones-system
kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.13.0/install/yaml/install.yaml
```

## 4. Verify Agones

```sh
kubectl describe --namespace agones-system pods
```

## 5. Install Custom Minecraft DNS Controller

### [Controller Documentation](./controller)

```sh
kubectl apply -f https://raw.githubusercontent.com/saulmaldonado/agones-minecraft/main/k8s/agones-mc-dns-controller.yaml
```

## 6. Deploy Minecraft GameServer Fleet

```sh
sed 's/<DOMAIN_NAME>/example.com/' k8s/mc-server-fleet.yml | kubectl apply -f - # replace 'example.com' with the domain you will be using
```

## 7. List Minecraft GameServer Addresses

```sh
kubectl get gs -o jsonpath='{.items[*].metadata.annotations.agones-mc/externalDNS}'
```
