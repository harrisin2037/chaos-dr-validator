package chaos

import (
	"context"
	"fmt"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
)

// ApplyChaosExperiment applies a chaos experiments to the specified application.
func ApplyChaosExperiment(ctx context.Context, cl client.Client, cr *chaosdrv1.ChaosDRTest, chaosName, chaosType string) error {
	switch chaosType {
	case "pod-delete":
		// Convert appSelector map to labels.Selector
		selector := labels.SelectorFromSet(cr.Spec.AppSelector)
		listOptions := &client.ListOptions{
			LabelSelector: selector,
			Namespace:     cr.Namespace,
		}

		// List pods matching the selector
		var podList corev1.PodList
		if err := cl.List(ctx, &podList, listOptions); err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}

		if len(podList.Items) == 0 {
			return fmt.Errorf("no pods found matching selector: %v", cr.Spec.AppSelector)
		}

		// Delete the first pod found (simplified for prototype)
		pod := podList.Items[0]
		if err := cl.Delete(ctx, &pod); err != nil {
			return fmt.Errorf("failed to delete pod %s: %v", pod.Name, err)
		}

		return nil
	case "network-delay":
		// Apply network delay using Chaos Mesh
		chaos := &chaosmeshv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      chaosName,
				Namespace: cr.Namespace,
			},
			//TODO:
			// Spec: chaosmeshv1alpha1.NetworkChaosSpec{
			// 	Selector: chaosmeshv1alpha1.SelectorSpec{
			// 		LabelSelectors: cr.Spec.AppSelector,
			// 	},
			// 	Action: "delay",
			// 	Delay:  &chaosmeshv1alpha1.DelaySpec{Latency: cr.Spec.ChaosParameters["delay"]},
			// },
		}
		return cl.Create(ctx, chaos)
	default:
		return fmt.Errorf("unsupported chaosType: %s", chaosType)
	}
}
