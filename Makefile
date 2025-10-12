.PHONY: all build test docker-build docker-push deploy generate

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