# krun

Run interactive one-off commands in Kubernetes pods with automatic cleanup

[![CI](https://github.com/ivoronin/krun/actions/workflows/test.yml/badge.svg)](https://github.com/ivoronin/krun/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/ivoronin/krun)](https://github.com/ivoronin/krun/releases)

[Overview](#overview) · [Features](#features) · [Installation](#installation) · [Usage](#usage) · [Configuration](#configuration) · [Requirements](#requirements) · [License](#license)

```bash
# Before: manual pod lifecycle management
kubectl run debug --image=alpine --rm -it --restart=Never -- sh
# Pod gets stuck, requires manual cleanup, no resource controls

# After: automatic cleanup with full pod configuration
krun -i alpine
```

## Overview

krun creates ephemeral pods for running interactive commands in Kubernetes clusters. It handles the complete pod lifecycle: creation, TTY attachment, and automatic deletion on exit or interrupt. The tool uses your existing kubeconfig or in-cluster service account for authentication.

## Features

- Automatic pod cleanup on exit or Ctrl-C (unless `--keeppod` specified)
- Interactive TTY attachment with stdin/stdout/stderr streaming
- Resource requests and limits for CPU and memory
- Custom labels, tolerations, and node selectors
- Service account specification for pod identity
- In-cluster and local kubeconfig authentication
- 300-second default pod startup timeout

## Installation

### GitHub Releases

Download from [Releases](https://github.com/ivoronin/krun/releases).

### Homebrew

```bash
brew install ivoronin/tap/krun
```

## Usage

### Basic Shell

```bash
krun -i alpine                          # Default command is /bin/sh
krun -i ubuntu /bin/bash                # Custom shell
krun -i busybox cat /etc/resolv.conf    # Run command and exit
```

### Resource Controls

```bash
krun -i alpine -c 100m -m 128Mi                     # CPU and memory requests
krun -i alpine -C 500m -M 512Mi                     # CPU and memory limits
krun -i alpine -c 100m -m 128Mi -C 500m -M 512Mi    # Both requests and limits
```

### Pod Configuration

```bash
krun -i alpine -n kube-system                       # Specify namespace
krun -i alpine -s my-service-account                # Use service account
krun -i alpine -l app=debug -l team=platform        # Add labels
krun -i alpine -e FOO=bar -e DEBUG=true             # Set environment variables
```

### Node Scheduling

```bash
krun -i alpine -N kubernetes.io/arch=arm64                              # Node selector
krun -i alpine -T dedicated:Equal:gpu:NoSchedule                        # Toleration
krun -i alpine -N node-type=compute -T spot:Exists::NoSchedule          # Combined
```

### Debug Options

```bash
krun -i alpine -v                   # Verbose output shows pod status
krun -i alpine -k                   # Keep pod after exit for inspection
krun -i alpine -t 600               # 10-minute startup timeout
```

### Complete Example

```bash
krun -i alpine \
    -n production \
    -s readonly-sa \
    -l app=debug -l ticket=INFRA-1234 \
    -c 100m -m 128Mi \
    -C 500m -M 512Mi \
    -N kubernetes.io/os=linux \
    -e KUBECONFIG=/dev/null \
    -t 120 \
    -v
```

### Flags Reference

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--image` | `-i` | Container image (required) | - |
| `--namespace` | `-n` | Kubernetes namespace | Current context |
| `--service-account` | `-s` | Service account name | - |
| `--timeout` | `-t` | Pod startup timeout in seconds | `300` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--labels` | `-l` | Pod labels (KEY=VALUE, repeatable) | - |
| `--toleration` | `-T` | Tolerations (KEY:OPERATOR:VALUE:EFFECT, repeatable) | - |
| `--node-selector` | `-N` | Node selectors (KEY=VALUE, repeatable) | - |
| `--requests-cpu` | `-c` | CPU request | - |
| `--requests-memory` | `-m` | Memory request | - |
| `--limits-cpu` | `-C` | CPU limit | - |
| `--limits-memory` | `-M` | Memory limit | - |
| `--keeppod` | `-k` | Keep pod after command completes | `false` |
| `--env` | `-e` | Environment variables (KEY=VALUE, repeatable) | - |

## Configuration

krun uses standard Kubernetes configuration:

- **Local**: `~/.kube/config` or `$KUBECONFIG`
- **In-cluster**: Service account token at `/var/run/secrets/kubernetes.io/serviceaccount/`

The namespace defaults to the current context namespace or the pod's namespace when running in-cluster.

## Requirements

### Kubernetes Cluster Access

- kubectl configured with cluster access, or
- Running in-cluster with appropriate service account

### RBAC Permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: krun
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["create", "get", "delete"]
  - apiGroups: [""]
    resources: ["pods/attach"]
    verbs: ["create"]
```

## License

[GPL-3.0](LICENSE)
