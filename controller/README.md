## Note: Controller is obsolete

### Use this [kubernetes-sigs/external-dns fork for GameServers](https://github.com/saulmaldonado/external-dns)

---

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

<br />
<p align="center">
  <h2 align="center">Agones Minecraft DNS Controller</h2>

  <p align="center">
  Custom Kubernetes controller for automating external DNS on Minecraft GameServers
    <br />
    <a href="https://github.com/saulmaldonado/agones-minecraft/tree/main/controller"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="#usage">View Example</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-minecraft/issues">Report Bug</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-minecraft/issues">Request Feature</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
    <li><a href="#author">Author</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

Custom Kubernetes controller for automating external DNS records for Agones Minecraft GameServers using third-party DNS providers

### Built With

- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
- [Agones](agones.dev/agones)

<!-- GETTING STARTED -->

## Getting Started

### Prerequisites

You need a running GKE cluster running with Agones resources and controllers installed

- GKE

```sh
gcloud container clusters create minecraft --cluster-version=1.18 \
  --tags=mc \
  --scopes=gke-default,"https://www.googleapis.com/auth/ndev.clouddns.readwrite" \ # GKE scope needed for Cloud DNS
  --node-labels=agones-mc/<DOMAIN_NAME> \ # Replace with the domain for the zone that the controller will manage
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

- Agones

```sh
kubectl create namespace agones-system
kubectl apply -f https://raw.githubusercontent.com/googleforgames/agones/release-1.13.0/install/yaml/install.yaml
```

### Installation

`kubectl apply`

```yml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: agones-mc-dns
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: agones-mc-dns
rules:
  - apiGroups: ['agones.dev']
    resources: ['gameservers']
    verbs: ['get', 'watch', 'list', 'update']
  - apiGroups: ['']
    resources: ['nodes']
    verbs: ['get', 'watch', 'list', 'update']
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: agones-mc-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: agones-mc-dns
subjects:
  - kind: ServiceAccount
    name: agones-mc-dns
    namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agones-mc-dns-controller
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: agones-mc-dns-controller
  template:
    metadata:
      labels:
        app: agones-mc-dns-controller
    spec:
      serviceAccountName: agones-mc-dns
      containers:
        - name: agones-mc-dns-controller
          image: saulmaldonado/agones-mc-dns-controller
          args:
            - --zone=<MANAGED_ZONE> # Replace with name of DNS managed zone
          imagePullPolicy: Always
```

#### Or run locally out of cluster

```sh
 docker run -it --rm --name agones-mc-dns-controller \
 -v $HOME/.kube/config:/root/.kube/config \ # Passes kubeconfig to container
 -v $HOME/.config/gcloud/:/root/.config/gcloud/ \ # Passes gcloud credentials to container
 -e HOME=/root \ # Needed for google oauth authentication
 saulmaldonado/agones-mc-dns-controller \
 --gcp-project=$(GCP_PROJECT) \ # Replace with GCP project ID
 --zone=$(MANAGED_ZONE) # Replace with Managed DNS zone
```

<!-- USAGE EXAMPLES -->

## Usage

This controller takes advantage of dynamic host port allocation that Agones provisions for GameServers. Instead of creating external DNS records for Kubernetes services and ingresses, DNS `A` and `SRV` are created to point to GameServer host nodes and the their GameServer ports.

### Nodes

To provision an `A` record for Nodes, they need to have `agones-mc/domain` label that contains the domain of the `zone` that the controller is managing. This will indicate to the controller that the Node needs an `A` record.

Labeling existing Nodes can be done using `kubectl`:

```sh
kubectl label node/<NODE_NAME> agones-mc/domain=<DOMAIN>
```

_All Nodes that you intend to host GameServers one should have this label._

A new annotation with `agones-mc/externalDNS` will contain the new `A` record that points to the Node IP.

Example:

| Node Name                                  | domain label  | Resulting `A` Record                                   |
| ------------------------------------------ | ------------- | ------------------------------------------------------ |
| `gke-minecraft-default-pool-79cd0803-42d7` | `example.com` | `gke-minecraft-default-pool-79cd0803-42d7.example.com` |

### GameServer

To provision an `SRV` record for GameServers, they need to contain an `agones-mc/domain` annotation that contains the domain for the controller's Managed Zone. This will indicate to the controller that the GameServer needs a `SRV` record.

#### GameServer Pod template example

```yml
template:
  metadata:
    annotations:
      agones-mc/domain: <DOMAIN_NAME> # Domain name of the managed zone
      # agones-mc/externalDNS: <GAMESERVER_NAME>.<DOMAIN_NAME> # Will be added by the controller
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft server image
        imagePullPolicy: Always
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: 'TRUE'

      - name: mc-monitor
        image: saulmaldonado/agones-mc-monitor # Agones monitor sidecar
        imagePullPolicy: Always
```

Once the pod has been created, a new `SRV` will be generated with the format `_minecraft._tcp.<GAMESERVER_NAME>.<DOMAIN>.` that points to the `A` record of the host Node `0 0 <PORT> <HOST_A_RECORD>`.

A new annotation `agones-mc/externalDNS` will then be added to the GameServer containing the URL from which players can connect to.

| GameServer Name   | Port | domain annotation               | Node `A` Record                                         | Resulting `SRV` Record                                                                                       | Minecraft Server URL        |
| ----------------- | ---- | ------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | --------------------------- |
| `mc-server-cfwd7` | 7908 | `agones-mc/domain: example.com` | `gke-minecraft-default-pool-79cd0803-42d7.example.com.` | `_minecraft._tcp.mc-server-cfwd7.example.com 0 0 7908 gke-minecraft-default-pool-79cd0803-42d7.example.com.` | mc-server-cfwd7.example.com |

#### [Full GameServer specification example](../k8s/mc-server.yml)

#### [Full Fleet specification example](../k8s/mc-server-fleet.yml)

### Run Locally with Docker

```sh
 docker run -it --rm --name agones-mc-dns-controller \
 -v $HOME/.kube/config:/root/.kube/config \ # Passes kubeconfig to container
 -v $HOME/.config/gcloud/:/root/.config/gcloud/ \ # Passes gcloud credentials to container
 -e HOME=/root \ # Needed for google oauth authentication
 saulmaldonado/agones-mc-dns-controller \
 --gcp-project=$(GCP_PROJECT) \ # Replace with GCP project ID
 --zone=$(MANAGED_ZONE) # Replace with Managed DNS zone
```

Flags:

```
  --gcp-project string
        GCP project id
  --kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  --zone string
        DNS zone that the controller will manage
```

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/saulmaldonado/agones-minecraft/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Clone the Project
3. Create your Feature or Fix Branch (`git checkout -b (feat|fix)/AmazingFeatureOrFix`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeatureOrFix'`)
5. Push to the Branch (`git push origin (feat|fix)/AmazingFeature`)
6. Open a Pull Request

### Build from source

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-minecraft.git
   ```

2. Build

   ```sh
   make go-build
   ```

### Build from Dockerfile

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-minecraft.git
   ```

2. Build

   ```sh
   docker build -t <hub-user>/agones-mc-dns-controller:latest ./controller
   ```

3. Push to Docker repo

   ```sh
   docker push <hub-user>/agones-mc-dns-controller:latest
   ```

<!-- LICENSE -->

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.

<!-- ACKNOWLEDGEMENTS -->

## Acknowledgements

- [itzg/docker-minecraft-server](https://github.com/itzg/docker-minecraft-server)

## Author

### Saul Maldonado

- 🐱 Github: [@saulmaldonado](https://github.com/saulmaldonado)
- 🤝 LinkedIn: [@saulmaldonado4](https://www.linkedin.com/in/saulmaldonado4/)
- 🐦 Twitter: [@saul_mal](https://twitter.com/saul_mal)
- 💻 Website: [saulmaldonado.com](https://saulmaldonado.com/)

## Show your support

Give a ⭐️ if this project helped you!

[contributors-shield]: https://img.shields.io/github/contributors/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[contributors-url]: https://github.com/saulmaldonado/agones-minecraft/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[forks-url]: https://github.com/saulmaldonado/agones-minecraft/network/members
[stars-shield]: https://img.shields.io/github/stars/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[stars-url]: https://github.com/saulmaldonado/agones-minecraft/stargazers
[issues-shield]: https://img.shields.io/github/issues/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[issues-url]: https://github.com/saulmaldonado/agones-minecraft/issues
[license-shield]: https://img.shields.io/github/license/saulmaldonado/agones-minecraft.svg?style=for-the-badge
[license-url]: https://github.com/saulmaldonado/agones-minecraft/blob/master/LICENSE.txt
