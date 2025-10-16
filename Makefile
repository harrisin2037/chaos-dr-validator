.PHONY: all build test docker-build docker-push deploy generate test-unit test-integration test-local test-all

all: build

build:
	go build -o bin/manager cmd/operator/main.go
	cd cmd/sidecar && cargo build --release

test:
	go test ./...
	cd cmd/sidecar && cargo test

docker-build:
	docker build -t localhost:5000/chaosdr-operator:latest -f Dockerfile.operator .
	docker build -t localhost:5000/chaosdr-sidecar:latest -f Dockerfile.sidecar .

docker-push:
	docker push localhost:5000/chaosdr-operator:latest
	docker push localhost:5000/chaosdr-sidecar:latest

deploy:
	kubectl apply -f config/crd/chaosdr.io_chaodrtests.yaml
	kubectl apply -f config/rbac/
	kubectl apply -f config/manager/

generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
	controller-gen crd:trivialVersions=true paths="./api/..." output:crd:artifacts:config=config/crd

helm-install:
	helm install chaosdr-operator charts/ --values charts/values.yaml

test:
	go test ./...
	cd cmd/sidecar && cargo test

test-unit:
	go test ./internal/chaos/... -v
	go test ./controllers/... -v
	cd cmd/sidecar && cargo test

test-integration:
	./test-operator.sh

test-local:
	./setup-test-env.sh

test-all: test-unit test-local test-integration

test-pod-delete:
	kubectl apply -f config/samples/chaosdr_v1_chaodrtest.yaml
	kubectl wait --for=condition=complete chaodrtest/redis-dr-test --timeout=300s
	kubectl get chaodrtest redis-dr-test -o yaml

logs:
	kubectl logs -l app=chaosdr-operator -f --all-containers

debug:
	@echo "=== ChaosDRTest Resources ==="
	kubectl get chaodrtest -A
	@echo "\n=== PodChaos Resources ==="
	kubectl get podchaos -A
	@echo "\n=== NetworkChaos Resources ==="
	kubectl get networkchaos -A
	@echo "\n=== Operator Logs ==="
	kubectl logs -l app=chaosdr-operator --tail=50
