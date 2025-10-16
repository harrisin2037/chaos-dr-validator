package controllers

import (
	"context"
	"testing"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
	ctrr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestChaosDRTestReconcile(t *testing.T) {
	// Setup fake client
	cl := fake.NewClientBuilder().Build()
	r := &ChaosDRTestReconciler{Client: cl}

	// Create test CR
	_ = &chaosdrv1.ChaosDRTest{
		Spec: chaosdrv1.ChaosDRTestSpec{
			AppSelector:      map[string]string{"app": "redis"},
			ChaosType:        "pod-delete",
			ValidationScript: "curl http://redis/healthz",
		},
	}

	// Test reconcile
	_, err := r.Reconcile(context.Background(), ctrr.Request{})
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}
}
