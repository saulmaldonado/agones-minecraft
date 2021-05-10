# Agones Minecraft

Minecraft dedicated game server cluster hosting solution using GKE and Agones.

Suitable for both ephemeral and eternal game servers

## Installation

### Prerequisites

- [gcloud](https://cloud.google.com/sdk/docs/install)
- [kubectl](https://kubernetes.io/docs/tasks/tools/included/install-kubectl-gcloud/)

### 1. Create public DNS Zone

This public DNS zone will be used to assign `A` and `SRV` DNS record to Minecraft GameServers

```sh
gcloud dns managed-zones create agones-minecraft \
    --description="agones-mc-dns-controller managed DNS zone" \
    --dns-name=<DOMAIN> \
    --visibility=public
```

Point your domain to Google DNS nameservers

```sh
gcloud dns managed-zones describe agones-minecraft
```

### 2. Create backup storage bucket

This bucket will contain zip world backups of Minecraft server worlds

```
gsutil mc -l us-central1 gs://agones-minecraft-mc-worlds
```

### 3. Create Kubernetes Cluster

Creates GKE cluster with 2 _n2-standard-4_ (4 x vCPU, 16GB), tags for firewall, and necessary scopes for Cloud DNS and Cloud Storage

```sh
gcloud container clusters create minecraft --cluster-version=1.18 \
  --tags=mc \
  --scopes=gke-default,storage-rw,"https://www.googleapis.com/auth/ndev.clouddns.readwrite" \
  --num-nodes=2 \
  --no-enable-autoupgrade \
  --machine-type=n2-standard-4
```

Set cluster as default and get credentials for `kubectl`

```sh
gcloud config set container/cluster minecraft
gcloud container clusters get-credentials minecraft
```

### 4. Add Firewall rules

Players will connect to GameServer Pods using controller allocated hostPorts. Assign firewall rules to open all Agones allocatable ports.

```sh
gcloud compute firewall-rules create mc-server-firewall \
  --allow tcp:7000-8000 \
  --target-tags mc \
  --description "Firewall rule to allow mc server tcp traffic"
```

### 5. Install Agones

```sh
kubectl create namespace agones-system
kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.14.0/install/yaml/install.yaml
```

or

```sh
helm repo add agones https://agones.dev/chart/stable
helm repo update
helm install agones --namespace agones-system --create-namespace agones/agones
```

### 6. Verify Agones Installation

```sh
kubectl get pods -n agones-system
```

### 7. Install ExternalDNS

#### [Controller Documentation](https://github.com/saulmaldonado/external-dns)

This controller is a fork of [kubernetes-sigs/external-dns](https://github.com/kubernetes-sigs/external-dns) custom support for Agones GameServer sources. It will manage `A` and `SRV` records for Agones GameServers and Fleets allowing players to connect using a unique subdomain

```sh
kubectl apply -f https://raw.githubusercontent.com/saulmaldonado/agones-minecraft/main/k8s/external-dns.yml # agones-minecraft matches the name of zone created earlier
```

## Deploy Java Servers

### 1. Deploy Minecraft GameServer Fleet

Fleets will deploy and manage a set of `Ready` GameServers that can immediately be allocated for players to connect to.

```sh
sed 's/<DOMAIN>/example.com/' k8s/mc-server-fleet.yml | kubectl apply -f - # replace 'example.com' with the domain of the managed zone
```

### 2. Allocate A Ready Server

```sh
kubectl create -f k8s/allocation.yml
```

### 3. Connect

Players can connect using unique subdomain

```sh
<GAMESERVER_NAME>.<MANAGED_ZONE_DOMAIN>
```

example:

```
NAME             STATE       ADDRESS       PORT   NODE
mc-dj4jq-52tsd   Allocated   35.232.46.5   7701   gke-minecraft-default-pool-47e5fdf8-5h9q

external-dns.alpha.kubernetes.io/hostname: saulmaldonado.me.

mc-dj4jq-52tsd.saulmaldonado.me
```

## Deploy Bedrock Servers

### 1. Deploy Bedrock GameServer Fleet

Fleets will deploy and manage a set of `Ready` GameServers that can immediately be allocated for players to connect to.

```sh
sed 's/<DOMAIN>/example.com/' k8s/mc-bedrock-fleet.yml | kubectl apply -f - # replace 'example.com' with the domain of the managed zone
```

### 2. Allocate Ready Bedrock Servers

```sh
kubectl create -f k8s/bedrock-allocation.yml
```

### 3. Connect

Players can connect using unique subdomain and allocated port (bedrock does not support SRV records)

```sh
<GAMESERVER_NAME>.<MANAGED_ZONE_DOMAIN>:<GAMESERVER_PORT>
```

example:

```sh
NAME             STATE       ADDRESS       PORT   NODE
mc-dj4jq-52tsd   Allocated   35.232.46.5   7701   gke-minecraft-default-pool-47e5fdf8-5h9q

external-dns.alpha.kubernetes.io/hostname: saulmaldonado.me.

address: mc-dj4jq-52tsd.saulmaldonado.me
port: 7701
```
