package chaos

import (
	"context"
	"testing"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
)

func TestApplyPodDeleteChaos(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = chaosmeshv1alpha1.AddToScheme(scheme)
	_ = chaosdrv1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()

	cr := &chaosdrv1.ChaosDRTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "default",
		},
		Spec: chaosdrv1.ChaosDRTestSpec{
			AppSelector: map[string]string{"app": "redis"},
			ChaosType:   "pod-delete",
		},
	}

	err := applyPodDeleteChaos(context.Background(), cl, cr, "test-chaos")
	if err != nil {
		t.Fatalf("applyPodDeleteChaos failed: %v", err)
	}

	// Verify PodChaos was created
	chaos := &chaosmeshv1alpha1.PodChaos{}
	err = cl.Get(context.Background(), client.ObjectKey{
		Name:      "test-chaos",
		Namespace: "default",
	}, chaos)
	if err != nil {
		t.Fatalf("Failed to get created PodChaos: %v", err)
	}

	// Verify spec
	if chaos.Spec.Action != chaosmeshv1alpha1.PodKillAction {
		t.Errorf("Expected action pod-kill, got %s", chaos.Spec.Action)
	}
	if chaos.Spec.Mode != "one" {
		t.Errorf("Expected mode 'one', got %s", chaos.Spec.Mode)
	}
}

func TestApplyNetworkDelayChaos(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = chaosmeshv1alpha1.AddToScheme(scheme)
	_ = chaosdrv1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()

	cr := &chaosdrv1.ChaosDRTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "default",
		},
		Spec: chaosdrv1.ChaosDRTestSpec{
			AppSelector: map[string]string{"app": "redis"},
			ChaosType:   "network-delay",
			ChaosParameters: map[string]string{
				"delay":  "100ms",
				"jitter": "10ms",
			},
		},
	}

	err := applyNetworkDelayChaos(context.Background(), cl, cr, "test-chaos")
	if err != nil {
		t.Fatalf("applyNetworkDelayChaos failed: %v", err)
	}

	// Verify NetworkChaos was created
	chaos := &chaosmeshv1alpha1.NetworkChaos{}
	err = cl.Get(context.Background(), client.ObjectKey{
		Name:      "test-chaos",
		Namespace: "default",
	}, chaos)
	if err != nil {
		t.Fatalf("Failed to get created NetworkChaos: %v", err)
	}

	// Verify spec
	if chaos.Spec.Action != chaosmeshv1alpha1.DelayAction {
		t.Errorf("Expected action delay, got %s", chaos.Spec.Action)
	}
	if chaos.Spec.TcParameter.Delay.Latency != "100ms" {
		t.Errorf("Expected latency 100ms, got %s", chaos.Spec.TcParameter.Delay.Latency)
	}
}

func TestApplyNetworkDelayChaos_MissingDelay(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = chaosmeshv1alpha1.AddToScheme(scheme)
	_ = chaosdrv1.AddToScheme(scheme)

	cl := fake.NewClientBuilder().WithScheme(scheme).Build()

	cr := &chaosdrv1.ChaosDRTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "default",
		},
		Spec: chaosdrv1.ChaosDRTestSpec{
			AppSelector:     map[string]string{"app": "redis"},
			ChaosType:       "network-delay",
			ChaosParameters: map[string]string{}, // Missing delay parameter
		},
	}

	err := applyNetworkDelayChaos(context.Background(), cl, cr, "test-chaos")
	if err == nil {
		t.Fatal("Expected error for missing delay parameter, got nil")
	}
}

func TestCleanupChaosExperiment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = chaosmeshv1alpha1.AddToScheme(scheme)

	// Create a PodChaos to cleanup
	chaos := &chaosmeshv1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-chaos",
			Namespace: "default",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).Build()

	err := CleanupChaosExperiment(context.Background(), cl, "default", "test-chaos", "pod-delete")
	if err != nil {
		t.Fatalf("CleanupChaosExperiment failed: %v", err)
	}

	// Verify it was deleted
	err = cl.Get(context.Background(), client.ObjectKey{
		Name:      "test-chaos",
		Namespace: "default",
	}, chaos)
	if err == nil {
		t.Fatal("Expected PodChaos to be deleted")
	}
}
