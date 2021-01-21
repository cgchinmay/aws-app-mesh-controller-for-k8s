package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sWrapper represents an interface with all the common operations on pod objects
type K8sWrapper interface {
	GetPod(namespace string, name string) (*v1.Pod, error)
	ListPodsWithMatchingLabels(opts client.ListOptions) (*v1.PodList, error)
}

// k8sWrapper is the wrapper object with the client
type k8sWrapper struct {
	podController *PodController
	cacheClient   client.Client
}

// NewK8sWrapper returns a new K8sWrapper
func NewK8sWrapper(client client.Client, podController *PodController) K8sWrapper {
	return &k8sWrapper{
		cacheClient:   client,
		podController: podController,
	}
}

// GetPod returns the pod object using NamespacedName
func (k *k8sWrapper) GetPod(namespace string, name string) (*v1.Pod, error) {
	nsName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}.String()
	obj, exists, err := k.podController.GetDataStore().GetByKey(nsName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("failed to find pod %s", nsName)
	}
	return obj.(*v1.Pod), nil
}

// ListPods return list of pods within a Namespace having Matching Labels
// ListOptions.LabelSelector must be specified to return pods with matching labels
// ListOptions.Namespace will scope result list to a given namespace
func (k *k8sWrapper) ListPodsWithMatchingLabels(opts client.ListOptions) (*v1.PodList, error) {
	var items []interface{}
	var err error

	if opts.Namespace != "" {
		items, err = k.podController.GetDataStore().ByIndex(Namespace, opts.Namespace)
	} else {
		items = k.podController.GetDataStore().List()
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	podList := &v1.PodList{}

	var labelSel labels.Selector
	if opts.LabelSelector != nil {
		labelSel = opts.LabelSelector
	}

	for _, item := range items {
		pod, ok := item.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("cache contained %T, which is not a Pod", item)
		}

		meta, err := apimeta.Accessor(pod)
		if err != nil {
			return nil, err
		}
		if labelSel != nil {
			lbls := labels.Set(meta.GetLabels())
			if !labelSel.Matches(lbls) {
				continue
			}
		}
		podList.Items = append(podList.Items, *pod)
	}
	return podList, nil
}
