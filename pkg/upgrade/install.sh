#!/bin/sh
set -e

# Define all known Gateway API CRDs for --all and --list
ALL_KNOWN_CRDS="gatewayclasses.gateway.networking.k8s.io gateways.gateway.networking.k8s.io httproutes.gateway.networking.k8s.io grpcroutes.gateway.networking.k8s.io referencegrants.gateway.networking.k8s.io tlsroutes.gateway.networking.k8s.io tcproutes.gateway.networking.k8s.io udproutes.gateway.networking.k8s.io backendtlspolicies.gateway.networking.k8s.io backendlbpolicies.gateway.networking.k8s.io"

LIST_MODE=false
ALL_MODE=false

# Argument parsing
while [ "$#" -gt 0 ]; do
  case "$1" in
    -l|--list)
      LIST_MODE=true
      shift
      ;;
    -a|--all)
      ALL_MODE=true
      shift
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

if [ "$LIST_MODE" = "true" ]; then
  echo "Installed Gateway API CRDs:"
  printf "%-50s %-20s %-20s\n" "CRD NAME" "VERSION" "CHANNEL"
  for CRD in $ALL_KNOWN_CRDS; do
    if kubectl get crd "$CRD" >/dev/null 2>&1; then
      VER=$(kubectl get crd "$CRD" -o jsonpath='{.metadata.annotations.gateway\.networking\.k8s\.io/bundle-version}' 2>/dev/null || echo "Unknown")
      CHAN=$(kubectl get crd "$CRD" -o jsonpath='{.metadata.annotations.gateway\.networking\.k8s\.io/channel}' 2>/dev/null || echo "Unknown")
      printf "%-50s %-20s %-20s\n" "$CRD" "$VER" "$CHAN"
    fi
  done
  exit 0
fi

if [ "$ALL_MODE" = "true" ]; then
  echo "Fetching all installed Gateway API CRDs..."
  # Fetch all CRDs that have the Gateway API suffix
  INSTALLED_CRDS=$(kubectl get crds -o name | grep "gateway.networking.k8s.io" | cut -d/ -f2)
  
  if [ -z "$INSTALLED_CRDS" ]; then
    echo "No Gateway API CRDs found in the cluster."
    exit 0
  fi
  # Replace newlines with spaces for the loop
  CRDS=$(echo "$INSTALLED_CRDS" | tr '\n' ' ')
else
  if [ -z "$CRD_SUBSET" ]; then
    echo "Error: CRD_SUBSET environment variable is not set (and --all not used)."
    exit 1
  fi
  # Split comma-separated CRD list.
  CRDS=$(echo $CRD_SUBSET | tr "," " ")
fi

if [ -z "$GATEWAY_API_VERSION" ]; then
  echo "Error: GATEWAY_API_VERSION environment variable is not set."
  exit 1
fi
if [ -z "$RELEASE_CHANNEL" ]; then
  echo "Error: RELEASE_CHANNEL environment variable is not set."
  exit 1
fi

if [ "$RELEASE_CHANNEL" != "standard" ]; then
  echo "Error: This installer only supports the 'standard' release channel. Target channel is '$RELEASE_CHANNEL'."
  exit 1
fi

echo "Starting Gateway API CRD Installer"
echo "Target Version: $GATEWAY_API_VERSION"
echo "Target Channel: $RELEASE_CHANNEL"

for CRD in $CRDS; do
  echo "----------------------------------------------------------------"
  echo "Processing CRD: $CRD"
  
  if kubectl get crd "$CRD" >/dev/null 2>&1; then
    # Fetch annotations.
    EXISTING_CHANNEL=$(kubectl get crd "$CRD" -o jsonpath='{.metadata.annotations.gateway\.networking\.k8s\.io/channel}' 2>/dev/null || true)
    EXISTING_VERSION=$(kubectl get crd "$CRD" -o jsonpath='{.metadata.annotations.gateway\.networking\.k8s\.io/bundle-version}' 2>/dev/null || true)
    
    if [ -z "$EXISTING_CHANNEL" ] || [ -z "$EXISTING_VERSION" ]; then
       echo "Warning: Existing CRD $CRD is missing standard Gateway API version/channel annotations. Proceeding with caution."
    else
       echo "Found installed CRD: $CRD"
       echo "  Current Version: $EXISTING_VERSION"
       echo "  Current Channel: $EXISTING_CHANNEL"
       
       if [ "$EXISTING_CHANNEL" != "$RELEASE_CHANNEL" ]; then
          echo "Mismatch: Existing channel ($EXISTING_CHANNEL) != Target channel ($RELEASE_CHANNEL). Skipping."
          continue
       fi

       # Compare versions
       NEWER_VERSION=$(printf "%s\n%s" "$EXISTING_VERSION" "$GATEWAY_API_VERSION" | sort -V | tail -n1)
       if [ "$NEWER_VERSION" = "$EXISTING_VERSION" ] && [ "$EXISTING_VERSION" != "$GATEWAY_API_VERSION" ]; then
          echo "Skipping: Existing version $EXISTING_VERSION is newer than target $GATEWAY_API_VERSION."
          continue
       elif [ "$EXISTING_VERSION" = "$GATEWAY_API_VERSION" ]; then
          echo "Skipping: Existing version $EXISTING_VERSION is already installed."
          continue
       fi
       echo "Upgrading $CRD from $EXISTING_VERSION to $GATEWAY_API_VERSION..."
    fi
  else
    echo "CRD $CRD not found. Proceeding to install."
  fi

  # Construct Raw GitHub URL
  CRD_GROUP=$(echo "$CRD" | cut -d. -f2-)
  CRD_PLURAL=$(echo "$CRD" | cut -d. -f1)
  REPO_FILENAME="${CRD_GROUP}_${CRD_PLURAL}.yaml"
  
  URL="https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${GATEWAY_API_VERSION}/config/crd/${RELEASE_CHANNEL}/${REPO_FILENAME}"
  
  echo "Downloading manifest from: $URL"
  if ! wget -qO /tmp/${CRD}.yaml "$URL"; then
      echo "Error: Failed to download manifest for $CRD from $URL"
      exit 1
  fi

  echo "Applying manifest for $CRD..."
  if kubectl apply --server-side -f /tmp/${CRD}.yaml; then
     INSTALLED_VERSION=$(kubectl get crd "$CRD" -o jsonpath='{.metadata.annotations.gateway\.networking\.k8s\.io/bundle-version}' 2>/dev/null || true)
     echo "Success: $CRD successfully applied. New version: $INSTALLED_VERSION"
  else
     echo "Error: Failed to apply manifest for $CRD."
     exit 1
  fi
done

echo "----------------------------------------------------------------"
echo "Gateway API CRD installation/upgrade process complete."⏎   
