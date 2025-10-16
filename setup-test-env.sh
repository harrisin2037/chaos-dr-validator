#!/bin/bash
set -e

echo "=== Setting up local test environment ==="

# Check prerequisites
command -v kind >/dev/null 2>&1 || { echo "kind is required but not installed. Install from https://kind.sigs.k8s.io/"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed."; exit 1; }

CLUSTER_NAME="chaosdr-test"

echo "Step 1: Create Kind cluster with local registry"
cat <<EOF | kind create cluster --name $CLUSTER_NAME --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
    endpoint = ["http://kind-registry:5000"]
EOF

# Create local registry if it doesn't exist
if ! docker ps | grep -q kind-registry; then
    echo "Creating local Docker registry..."
    docker run -d --restart=always -p "127.0.0.1:5000:5000" --name kind-registry registry:2
fi

# Connect registry to kind network
if ! docker network ls | grep -q kind; then
    docker network create kind
fi
docker network connect kind kind-registry 2>/dev/null || true

echo "Step 2: Install Chaos Mesh"
kubectl create ns chaos-mesh || true
curl -sSL https://mirrors.chaos-mesh.org/v2.6.3/install.sh | bash -s -- --local kind

echo "Waiting for Chaos Mesh to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller-manager -n chaos-mesh --timeout=300s

echo "Step 3: Install MinIO for backup storage"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: minio
---
apiVersion: v1
kind: Secret
metadata:
  name: minio-creds
  namespace: default
type: Opaque
stringData:
  access-key: minioadmin
  secret-key: minioadmin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: minio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
        - server
        - /data
        env:
        - name: MINIO_ROOT_USER
          value: minioadmin
        - name: MINIO_ROOT_PASSWORD
          value: minioadmin
        ports:
        - containerPort: 9000
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: minio
spec:
  ports:
  - port: 9000
    targetPort: 9000
  selector:
    app: minio
EOF

echo "Step 4: Build and load operator images"
echo "Building operator..."
docker build -t localhost:5000/chaosdr-operator:latest -f Dockerfile.operator .
docker push localhost:5000/chaosdr-operator:latest

echo "Building sidecar..."
docker build -t localhost:5000/chaosdr-sidecar:latest -f Dockerfile.sidecar .
docker push localhost:5000/chaosdr-sidecar:latest

echo "Step 5: Deploy operator"
kubectl apply -f config/crd/chaosdr.io_chaodrtests.yaml
kubectl apply -f config/rbac/
kubectl apply -f config/manager/manager.yaml

echo "Waiting for operator to be ready..."
kubectl wait --for=condition=available deployment/chaosdr-operator --timeout=300s

echo ""
echo "=== Test environment ready! ==="
echo ""
echo "Cluster: $CLUSTER_NAME"
echo "Registry: localhost:5000"
echo ""
echo "Next steps:"
echo "  1. Run tests: ./test-operator.sh"
echo "  2. Check logs: kubectl logs -l app=chaosdr-operator -f"
echo "  3. Delete cluster: kind delete cluster --name $CLUSTER_NAME"