# ChaosDRValidator

Kubernetes operator for chaos-infused DR testing with MinIO storage.

## Prerequisites
- Kubernetes cluster (bare-metal)
- MinIO deployed (StatefulSet, Service at `minio:9000`)
- Velero installed with MinIO backend
- Go 1.24+, Rust 1.80+
- Local Docker registry (`docker run -d -p 5000:5000 registry`)
- controller-gen (`go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.0`)
- protoc (`brew install protobuf` or equivalent)
- protoc-gen-go and protoc-gen-go-grpc (`go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0` and `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0`)

## Setup
1. Clone repo: `git clone https://github.com/harrisin2037/chaos-dr-validator`
2. Install dependencies: `go mod download && cd cmd/sidecar && cargo fetch`
3. Generate CRD and deep-copy: `make generate`
4. Build: `make build`
5. Build Docker images: `make docker-build`
6. Push images: `make docker-push`
7. Deploy operator: `make deploy`
8. Apply sample CR: `kubectl apply -f config/samples/chaosdr_v1_chaodrtest.yaml`

## Demo
- Apply CR to test Redis app.
- Operator triggers Velero backup, custom pod-delete chaos, sandbox restore, validation.
- Check MinIO (`backups` bucket) for proof.
- Monitor metrics at operator's `:8080/metrics`.

# ChaosDR Validator

A Kubernetes operator for disaster recovery testing using chaos engineering.

## Installation
```bash
make docker-build docker-push
make deploy
```

## Usage
Create a `ChaosDRTest` CR:
```yaml
apiVersion: chaosdr.io/v1
kind: ChaosDRTest
metadata:
name: redis-dr-test
spec:
appSelector:
    app: redis
chaosType: pod-delete
validationScript: "curl http://redis-sandbox/healthz"
```

## Development
- Build: `make build`
- Test: `make test`
- Deploy: `make deploy`