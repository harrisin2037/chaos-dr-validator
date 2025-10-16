#!/bin/bash
set -e

echo "=== ChaosDR Operator Integration Test ==="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE=${TEST_NAMESPACE:-default}
TEST_NAME="redis-dr-test-$(date +%s)"

cleanup() {
    echo -e "${YELLOW}Cleaning up test resources...${NC}"
    kubectl delete chaodrtest $TEST_NAME -n $NAMESPACE --ignore-not-found
    kubectl delete pod redis-test -n $NAMESPACE --ignore-not-found
}

trap cleanup EXIT

echo -e "${YELLOW}Step 1: Check prerequisites${NC}"
# Check if cluster is ready
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Kubernetes cluster not accessible${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Cluster accessible${NC}"

# Check if Chaos Mesh is installed
if ! kubectl get crd podchaos.chaos-mesh.org &> /dev/null; then
    echo -e "${RED}❌ Chaos Mesh not installed${NC}"
    echo "Install with: kubectl apply -f https://mirrors.chaos-mesh.org/latest/crd.yaml"
    exit 1
fi
echo -e "${GREEN}✓ Chaos Mesh installed${NC}"

# Check if operator is running
if ! kubectl get deployment chaosdr-operator -n $NAMESPACE &> /dev/null; then
    echo -e "${RED}❌ ChaosDR operator not deployed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Operator deployed${NC}"

echo -e "\n${YELLOW}Step 2: Deploy test application (Redis)${NC}"
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: redis-test
  namespace: $NAMESPACE
  labels:
    app: redis
spec:
  containers:
  - name: redis
    image: redis:7-alpine
    ports:
    - containerPort: 6379
EOF

echo "Waiting for Redis pod to be ready..."
kubectl wait --for=condition=ready pod/redis-test -n $NAMESPACE --timeout=60s
echo -e "${GREEN}✓ Redis pod ready${NC}"

echo -e "\n${YELLOW}Step 3: Create ChaosDRTest CR${NC}"
kubectl apply -f - <<EOF
apiVersion: chaosdr.io/v1
kind: ChaosDRTest
metadata:
  name: $TEST_NAME
  namespace: $NAMESPACE
spec:
  appSelector:
    app: redis
  chaosType: pod-delete
  validationScript: "echo 'Validation passed'"
  validationConfig:
    script: "echo 'Testing Redis' && sleep 2"
  chaosParameters: {}
EOF

echo -e "${GREEN}✓ ChaosDRTest created${NC}"

echo -e "\n${YELLOW}Step 4: Monitor test execution${NC}"
echo "Waiting for ChaosDRTest to complete (this may take a few minutes)..."

# Wait up to 5 minutes for test completion
for i in {1..30}; do
    STATUS=$(kubectl get chaodrtest $TEST_NAME -n $NAMESPACE -o jsonpath='{.status.success}' 2>/dev/null || echo "")
    ERROR=$(kubectl get chaodrtest $TEST_NAME -n $NAMESPACE -o jsonpath='{.status.errorMessage}' 2>/dev/null || echo "")
    
    if [ "$STATUS" == "true" ]; then
        echo -e "${GREEN}✓ Test completed successfully!${NC}"
        break
    elif [ "$STATUS" == "false" ]; then
        echo -e "${RED}❌ Test failed: $ERROR${NC}"
        exit 1
    fi
    
    echo "  [$(date +%H:%M:%S)] Waiting... (attempt $i/30)"
    sleep 10
done

if [ "$STATUS" != "true" ]; then
    echo -e "${RED}❌ Test timeout${NC}"
    kubectl describe chaodrtest $TEST_NAME -n $NAMESPACE
    exit 1
fi

echo -e "\n${YELLOW}Step 5: Verify test results${NC}"
echo "ChaosDRTest Status:"
kubectl get chaodrtest $TEST_NAME -n $NAMESPACE -o yaml

# Check if PodChaos was created
CHAOS_NAME="chaos-$TEST_NAME"
if kubectl get podchaos $CHAOS_NAME -n $NAMESPACE &> /dev/null; then
    echo -e "${GREEN}✓ PodChaos created${NC}"
else
    echo -e "${YELLOW}⚠ PodChaos not found (may have been cleaned up)${NC}"
fi

echo -e "\n${GREEN}=== All tests passed! ===${NC}"