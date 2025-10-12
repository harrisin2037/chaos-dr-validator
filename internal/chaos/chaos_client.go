package chaos

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdrv1 "github.com/harrisin2037/chaos-dr-validator/api/v1"
)

// ApplyChaosExperiment applies a chaos experiment (pod-delete) to the specified application.
func ApplyChaosExperiment(ctx context.Context, cl client.Client, cr *chaosdrv1.ChaosDRTest, chaosName, chaosType string) error {
	if chaosType != "pod-delete" {
		return fmt.Errorf("unsupported chaosType: %s, only pod-delete is supported", chaosType)
	}

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
}
