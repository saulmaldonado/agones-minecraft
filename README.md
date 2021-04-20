# Installation

## 1. Create Kubernetes Cluster

```sh
gcloud container clusters create minecraft --cluster-version=1.18 \
  --tags=mc \
  --scopes=gke-default,"https://www.googleapis.com/auth/ndev.clouddns.readwrite" \
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

## 2. Install Agones

```sh
kubectl create namespace agones-system
kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.13.0/install/yaml/install.yaml
```

## 3. Verify Agones

```sh
kubectl describe --namespace agones-system pods
```

## 4. Create Minecraft Server Fleet

```sh
kubectl apply -f k8s/mc-server-fleet.yml
```

## 5. Get A GameServer IP

```sh
kube get fleet mc-fleet
```
