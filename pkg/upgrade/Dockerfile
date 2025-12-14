FROM alpine:3.19

# Install dependencies:
# - ca-certificates: required for HTTPS calls to GitHub/Kubernetes API
# - wget: used to download the CRD manifests
# - openssl: often required for robust HTTPS support with wget
RUN apk add --no-cache ca-certificates curl openssl wget

# Install kubectl matching the target architecture
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  chmod +x kubectl && \
  mv kubectl /usr/local/bin/kubectl


# Create a directory for our scripts
WORKDIR /scripts

# Copy the installation script into the image
COPY install.sh /scripts/install.sh
RUN chmod +x /scripts/install.sh

# Set the script as the entrypoint
ENTRYPOINT ["/bin/sh", "/scripts/install.sh"]
