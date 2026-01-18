# Vault Operator

A Kubernetes operator for managing HashiCorp Vault infrastructure and resources declaratively using Custom Resource Definitions (CRDs).

## Overview

Vault Operator simplifies the deployment and management of Vault servers and their associated resources within Kubernetes environments. Built with [Kubebuilder](https://kubebuilder.io/), it provides a native Kubernetes experience for working with Vault, enabling you to define authentication methods, policies, secrets, and secret engines as Kubernetes manifests.

### Key Features

- **Automated Vault Lifecycle Management**: Auto-initialization and auto-unsealing of Vault servers
- **Declarative Configuration**: Manage Vault resources using Kubernetes CRDs
- **Secret Management**: Create and manage secrets with optional random generation for non-sensitive configuration data
- **Authentication Methods**: Native support for AppRole and UserPass authentication
- **Policy as Code**: Define Vault policies as Kubernetes resources
- **Secret Engines**: Declaratively configure and manage secret engines
- **Export Capabilities**: Share secrets across namespaces securely

## Custom Resources

The operator manages the following Vault resources:

| Resource | Description |
|----------|-------------|
| `VaultServer` | Vault server deployment with automatic initialization and unsealing |
| `AppRole` | AppRole authentication method configuration with credential export |
| `UserPass` | Username/password authentication method |
| `Policy` | Vault policy definitions |
| `Secret` | Secret storage with optional random generation |
| `SecretEngine` | Secret engine configuration and management |

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured to communicate with your cluster
- HashiCorp Vault

### Installation

```bash
# Install the operator CRDs
kubectl apply -f config/crd/bases/

# Deploy the operator
kubectl apply -f config/manager/manager.yaml
```

### Integrate with a Vault Server

Create a `VaultServer` resource to deploy and configure Vault:

```yaml
apiVersion: vault.example.com/v1alpha1
kind: VaultServer
metadata:
  name: vault-primary
spec:
  server:
    serviceName: vault
    port: 8200
    namespace: vault-system
  init: true
  autoUnlock: true
```

Apply the configuration:

```bash
kubectl apply -f vault-server.yaml
```

The operator will automatically initialize and unseal the Vault server.

### Create an AppRole

Define an AppRole for application authentication:

```yaml
apiVersion: vault.example.com/v1alpha1
kind: AppRole
metadata:
  name: my-app-role
spec:
  vaultOperator:
    name: vault-primary
    namespace: vault-system
  name: my-application
  mount_path: approle
  policies:
    - my-app-policy
  secret_id_ttl: 3600
  export:
    namespace: applications
```

The operator will create the AppRole in Vault and export the credentials as a Kubernetes Secret to the specified namespace.

## Configuration Management

Vault Operator supports using Vault for configuration management alongside sensitive secrets. You can create configuration entries with:

- **Static values**: For known configuration parameters
- **Random generation**: Let the operator generate random values for non-sensitive configs like instance IDs or correlation tokens

Example secret with random generation:

```yaml
apiVersion: vault.example.com/v1alpha1
kind: Secret
metadata:
  name: app-config
spec:
  vaultOperator:
    name: vault-primary
  path: secret/data/app/config
  data:
    database_url: "postgres://db.example.com:5432/mydb"
    instance_id: 
      random: true
      length: 16
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│           Kubernetes Cluster                    │
│                                                 │
│  ┌──────────────┐      ┌──────────────────┐   │
│  │   Vault      │◄─────│  Vault Operator  │   │
│  │   Server     │      │                  │   │
│  └──────────────┘      └──────────────────┘   │
│         ▲                       │              │
│         │                       │              │
│         │              ┌────────▼────────┐     │
│         │              │  Custom         │     │
│         └──────────────│  Resources      │     │
│                        │  (CRDs)         │     │
│                        └─────────────────┘     │
└─────────────────────────────────────────────────┘
```

## Status Tracking

All managed resources provide detailed status information:

```bash
kubectl get vaultservers
kubectl get approles
kubectl get policies
```

Each resource tracks:
- Current phase/synchronization status
- Detailed status messages
- Last update timestamps
- Kubernetes conditions (Available, Progressing, Degraded)

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/danielnegreirosb/vault-operator
cd vault-operator

# Install dependencies
go mod download

# Build the operator
make build

# Run tests
make test
```

### Running Locally

```bash
# Install CRDs
make install

# Run the operator locally
make run
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/danielnegreirosb/vault-operator/issues)
- **Discussions**: Join the conversation in [GitHub Discussions](https://github.com/danielnegreirosb/vault-operator/discussions)
- **Documentation**: Full documentation available at [docs/](docs/)

## Acknowledgments

Built with [Kubebuilder](https://kubebuilder.io/) and designed to work seamlessly with [HashiCorp Vault](https://www.vaultproject.io/).