ChaosDRValidator Design

Architecture

Go Operator: Watches ChaosDRTest CRs, orchestrates Velero backup, LitmusChaos pod-delete, sandbox restore, validation. Calls Rust sidecar via gRPC.

Rust Sidecar: Processes restored data (dedup, encrypt, MinIO upload).

Flow: CR apply → Velero backup → Chaos injection → Restore → Validate (curl healthz) → MinIO proof → Prometheus metrics.

Components: Velero (backup/restore), LitmusChaos (chaos), MinIO (storage), Prometheus (metrics).

Threat Model (STRIDE)

Spoofing: Mitigated by RBAC (least-privilege operator access).
Tampering: Data integrity via Rust checksums.
Repudiation: Structured logs in operator.
Information Disclosure: MinIO creds in K8s Secrets.
Denial of Service: Chaos tests ensure resilience.
Elevation of Privilege: Restricted RBAC, no root containers.