package chaos

import (
	"context"
	"fmt"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
)

// ApplyChaosExperiment applies a chaos experiment to the specified application.
func ApplyChaosExperiment(ctx context.Context, cl client.Client, cr *chaosdrv1.ChaosDRTest, chaosName, chaosType string) error {
	switch chaosType {
	case "pod-delete":
		return applyPodDeleteChaos(ctx, cl, cr, chaosName)
	case "network-delay":
		return applyNetworkDelayChaos(ctx, cl, cr, chaosName)
	default:
		return fmt.Errorf("unsupported chaosType: %s", chaosType)
	}
}

func applyPodDeleteChaos(ctx context.Context, cl client.Client, cr *chaosdrv1.ChaosDRTest, chaosName string) error {
	chaos := &chaosmeshv1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chaosName,
			Namespace: cr.Namespace,
		},
		Spec: chaosmeshv1alpha1.PodChaosSpec{
			ContainerSelector: chaosmeshv1alpha1.ContainerSelector{
				PodSelector: chaosmeshv1alpha1.PodSelector{
					Selector: chaosmeshv1alpha1.PodSelectorSpec{
						GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
							Namespaces:     []string{cr.Namespace},
							LabelSelectors: cr.Spec.AppSelector,
						},
					},
					Mode: chaosmeshv1alpha1.OneMode,
				},
			},
			Action: chaosmeshv1alpha1.PodKillAction,
		},
	}

	if err := cl.Create(ctx, chaos); err != nil {
		return fmt.Errorf("failed to create pod chaos: %v", err)
	}

	return nil
}

func applyNetworkDelayChaos(ctx context.Context, cl client.Client, cr *chaosdrv1.ChaosDRTest, chaosName string) error {
	// Validate required parameters
	delay, ok := cr.Spec.ChaosParameters["delay"]
	if !ok || delay == "" {
		return fmt.Errorf("network-delay requires 'delay' parameter (e.g., '100ms')")
	}

	// Build DelaySpec
	delaySpec := &chaosmeshv1alpha1.DelaySpec{
		Latency: delay,
	}

	// Add optional jitter if provided
	if jitter, ok := cr.Spec.ChaosParameters["jitter"]; ok && jitter != "" {
		delaySpec.Jitter = jitter
	}

	// Add optional correlation if provided
	if correlation, ok := cr.Spec.ChaosParameters["correlation"]; ok && correlation != "" {
		delaySpec.Correlation = correlation
	}

	chaos := &chaosmeshv1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      chaosName,
			Namespace: cr.Namespace,
		},
		Spec: chaosmeshv1alpha1.NetworkChaosSpec{
			PodSelector: chaosmeshv1alpha1.PodSelector{
				Selector: chaosmeshv1alpha1.PodSelectorSpec{
					GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
						Namespaces:     []string{cr.Namespace},
						LabelSelectors: cr.Spec.AppSelector,
					},
				},
				Mode: chaosmeshv1alpha1.AllMode,
			},
			Action: chaosmeshv1alpha1.DelayAction,
			TcParameter: chaosmeshv1alpha1.TcParameter{
				Delay: delaySpec,
			},
		},
	}

	if err := cl.Create(ctx, chaos); err != nil {
		return fmt.Errorf("failed to create network chaos: %v", err)
	}

	return nil
}

// CleanupChaosExperiment removes a chaos experiment
func CleanupChaosExperiment(ctx context.Context, cl client.Client, namespace, chaosName, chaosType string) error {
	switch chaosType {
	case "pod-delete":
		chaos := &chaosmeshv1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chaosName,
				Namespace: namespace,
			},
		}
		return client.IgnoreNotFound(cl.Delete(ctx, chaos))
	case "network-delay":
		chaos := &chaosmeshv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chaosName,
				Namespace: namespace,
			},
		}
		return client.IgnoreNotFound(cl.Delete(ctx, chaos))
	default:
		return fmt.Errorf("unsupported chaosType: %s", chaosType)
	}
}
