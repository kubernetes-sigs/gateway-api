# Gateway API CRD Installer

This directory contains a Kubernetes Job designed to safely install or upgrade [Gateway API](https://gateway-api.sigs.k8s.io/) Custom Resource Definitions (CRDs) in a cluster.

## Overview

Bundling Gateway API CRDs directly with add-ons (e.g., via Helm charts) can lead to conflicts when multiple controllers try to manage the same CRDs, potentially downgrading them or switching release channels unexpectedly.

This Job provides a safe mechanism to:
1.  **Install** missing Gateway API CRDs.
2.  **Upgrade** existing CRDs to a requested version *only if* the requested version is newer.
3.  **Avoid Conflicts** by refusing to overwrite CRDs from a different release channel (e.g., Standard vs. Experimental).

## Components

*   **`job-list.yaml`**: Job to list installed Gateway API CRDs and their versions.
*   **`job-all.yaml`**: Job to install/upgrade **all** Gateway API CRDs.
*   **`job-subset.yaml`**: Job to install/upgrade a **specific subset** of CRDs.
*   **`rbac.yaml`**: ServiceAccount, ClusterRole, and ClusterRoleBinding allowing the Job to manage CRDs.
*   **`Dockerfile`**: Builds the container image containing the `install.sh` script.

## Configuration

The Jobs are configured via environment variables set in the YAML files:
...
## Usage

### 1. Build and Apply Prerequisites
1.  **Build the Docker Image:**
    ```bash
    docker build -t gateway-api-crd-installer:local .
    ```
    *Note: If running in a Kind/Minikube cluster, ensure you load this image into the cluster nodes (e.g., `kind load docker-image gateway-api-crd-installer:local`).*

2.  **Apply RBAC:**
    ```bash
    kubectl apply -f rbac.yaml
    ```

### 2. Run a Job
Choose the appropriate job manifest for your needs.

#### Option A: List Installed CRDs
To see which Gateway API CRDs are currently installed and their versions:

```bash
kubectl apply -f job-list.yaml
kubectl logs -l job-name=gateway-api-crd-list
```

#### Option B: Upgrade Existing CRDs
To upgrade **all** currently installed Gateway API CRDs:
1.  Edit `job-all.yaml` to set your desired `GATEWAY_API_VERSION` and `RELEASE_CHANNEL`.
2.  Run:
    ```bash
    kubectl apply -f job-all.yaml
    kubectl logs -l job-name=gateway-api-crd-install-all
    ```
    *Note: This option only updates CRDs that are already present in the cluster. It will not install new, missing CRDs.*

#### Option C: Upgrade/Install Subset of CRDs
To upgrade or install only specific CRDs (e.g., if you only use `Gateway` and `HTTPRoute`):
1.  Edit `job-subset.yaml` to set `GATEWAY_API_VERSION`, `RELEASE_CHANNEL`, and `CRD_SUBSET`.
2.  Run:
    ```bash
    kubectl apply -f job-subset.yaml
    kubectl logs -l job-name=gateway-api-crd-install-subset
    ```

## Logic Details

For every CRD processed, the installer script performs the following checks:

1.  **Existence Check:** If the CRD does not exist in the cluster, it is installed immediately from the official upstream source.
2.  **Channel Safety Check:** If the CRD exists, it checks the `gateway.networking.k8s.io/channel` label. If the installed channel does not match the requested `RELEASE_CHANNEL`, the script skips this CRD to avoid accidentally switching a cluster from Experimental to Standard (or vice-versa) potentially breaking fields.
3.  **Upgrade Safety Check:** Automated upgrades are **only** supported for the `standard` channel. If `RELEASE_CHANNEL` is not `standard`, the job will exit with an error.
4.  **Version Check:** It compares the requested `GATEWAY_API_VERSION` with the installed `gateway.networking.k8s.io/bundle-version` annotation.
    *   If installed version < requested version: **UPGRADE**.
    *   If installed version >= requested version: **SKIP** (already up-to-date).

