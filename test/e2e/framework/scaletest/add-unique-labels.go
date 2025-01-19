package scaletest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type AddUniqueLabelsToAllPods struct {
	KubeConfigFilePath    string
	NumUniqueLabelsPerPod int
	Namespace             string
}

// Useful when wanting to do parameter checking, for example
// if a parameter length is known to be required less than 80 characters,
// do this here so we don't find out later on when we run the step
// when possible, try to avoid making external calls, this should be fast and simple
func (a *AddUniqueLabelsToAllPods) Prevalidate() error {
	return nil
}

// Primary step where test logic is executed
// Returning an error will cause the test to fail
func (a *AddUniqueLabelsToAllPods) Run() error {

	if a.NumUniqueLabelsPerPod < 1 {
		return nil
	}

	config, err := clientcmd.BuildConfigFromFlags("", a.KubeConfigFilePath)
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutSeconds*time.Second)
	defer cancel()

	resources, err := clientset.CoreV1().Pods(a.Namespace).List(ctx, metav1.ListOptions{})

	count := 0

	for _, resource := range resources.Items {
		patch := []patchStringValue{}

		for i := 0; i < a.NumUniqueLabelsPerPod; i++ {
			patch = append(patch, patchStringValue{
				Op:    "add",
				Path:  "/metadata/labels/uni-lab-" + fmt.Sprintf("%05d", count),
				Value: "val",
			})
			count++
		}

		patchBytes, err := json.Marshal(patch)
		if err != nil {
			return fmt.Errorf("error marshalling patch: %w", err)
		}

		clientset.CoreV1().Pods(a.Namespace).Patch(ctx, resource.Name,
			types.JSONPatchType,
			patchBytes,
			metav1.PatchOptions{},
		)
	}

	return nil
}

// Require for background steps
func (a *AddUniqueLabelsToAllPods) Stop() error {
	return nil
}
