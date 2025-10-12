
1. **Update Rust Dependencies**:
   - Run:
     ```bash
     cd cmd/sidecar && cargo fetch
     ```
     This downloads `minio-rsc v0.2.6` and other dependencies.

2. **Update Go Dependencies**:
   - For Go (if not already done):
     ```bash
     go mod tidy
     go mod download
     ```

3. **Generate Go gRPC Code** (if not already done):
   - Run:
     ```bash
     protoc --go_out=internal/proto --go-grpc_out=internal/proto cmd/sidecar/src/proto/drtest.proto
     ```
     This generates `internal/proto/drtest.pb.go` and `internal/proto/drtest_grpc.pb.go` for the Go operator.

4. **Build the Project**:
   - Run:
     ```bash
     make build
     ```
     This compiles the Go operator (`bin/manager`) and the Rust sidecar (`cmd/sidecar/target/release/sidecar`). The `build.rs` script generates the Rust gRPC code during `cargo build`.

5. **Build and Push Docker Images**:
   - Run:
     ```bash
     make docker-build
     make docker-push
     ```
     This builds and pushes `localhost:5000/chaosdr-operator:latest` and `localhost:5000/chaosdr-sidecar:latest`.

6. **Deploy the Operator**:
   - Run:
     ```bash
     make deploy
     ```
     This applies the CRD, RBAC, and deployment manifests.

7. **Prerequisites**:
   - Ensure MinIO is deployed with a `backups` bucket and a `minio-creds` Secret:
     ```yaml
     apiVersion: v1
     kind: Secret
     metadata:
       name: minio-creds
       namespace: default
     type: Opaque
     stringData:
       access-key: your-access-key
       secret-key: your-secret-key
     ```
   - Install Velero:
     ```bash
     velero install --provider aws --bucket backups --plugins velero/velero-plugin-for-aws:v1.10.0 --secret-file ./minio-creds
     ```
   - Deploy a Redis app (example):
     ```yaml
     apiVersion: apps/v1
     kind: Deployment
     metadata:
       name: redis
       namespace: default
     spec:
       selector:
         matchLabels:
           app: redis
       template:
         metadata:
           labels:
             app: redis
         spec:
           containers:
           - name: redis
             image: redis:6
             ports:
             - containerPort: 6379
     ---
     apiVersion: v1
     kind: Service
     metadata:
       name: redis
     spec:
       selector:
         app: redis
       ports:
       - port: 6379
         targetPort: 6379
     ```
   - Update the sample CR (`config/samples/chaosdr_v1_chaodrtest.yaml`) to use a valid validation script for Redis:
     ```yaml
     apiVersion: chaosdr.io/v1
     kind: ChaosDRTest
     metadata:
       name: redis-dr-test
       namespace: default
     spec:
       appSelector:
         app: redis
       chaosType: pod-delete
       validationScript: "redis-cli -h redis-sandbox ping"
     ```

8. **Test the Prototype**:
   - Apply the sample CR:
     ```bash
     kubectl apply -f config/samples/chaosdr_v1_chaodrtest.yaml
     ```
   - Check CR status:
     ```bash
     kubectl get chaodrtest redis-dr-test -o yaml
     ```
   - Verify MinIO (`backups/validation-*.bin`).
   - Check metrics:
     ```bash
     kubectl port-forward deploy/chaosdr-operator 8080:8080
     curl localhost:8080/metrics
     ```

---