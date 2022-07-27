package controller

import (
	"fmt"

	"context"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	listerscore "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// podController is a global pod controller.
// GetPod() will get pod resources from this pod controller.
var podController = &controller{}

// controller is the controller implementation for Pod resources.
type controller struct {
	podLister listerscore.PodLister
	// podSynced is a flag to determine if pod informer had been synced.
	podSynced cache.InformerSynced
}

// newController returns a new pod controller
func newController(
	podInformer cache.SharedIndexInformer,
	podLister listerscore.PodLister) *controller {

	controller := &controller{
		podLister: podLister,
		podSynced: podInformer.HasSynced,
	}

	return controller
}

// run will set up the event handlers for types we are interested in, as well
// as syncing informer coaches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finished processing their current work items.
func (c *controller) run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	// Start the informer factories to begin populating the informer caches
	log.Infof("Starting ratel-webterminal controller")
	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	log.Info("Pod synced successfully")
	return nil
}

// Init will create a global pod controller by calling newController()
func Init() {
	podHandler, err := pod.New(context.TODO(), args.GetKubeConfigFile(), "")
	if err != nil {
		log.Panicf("Create a pod handler error: %s", err.Error())
	}

	podController = newController(podHandler.Informer(), podHandler.Lister())
	stopCh := make(chan struct{})
	podHandler.InformerFactory().Start(stopCh)
	if err := podController.run(stopCh); err != nil {
		log.Panicf("Error running pod controller: %s", err.Error())
	}
}

// GetPod try get a pod resource with given namespace and pod name from
// global pod controller.
// If the pod resource no longer exist in pod lister, it will make this function
// caller to get pod by calling apiserver API directly.
func GetPod(namespace, name string) (*corev1.Pod, error) {
	podObj, err := podController.podLister.Pods(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("pod '%s/%s' in pod lister no longer exists, it will get pod resource by calling apiserver API directly", namespace, name)
		}
		return nil, err
	}
	return podObj, nil
}
