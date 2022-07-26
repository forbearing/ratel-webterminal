package controller

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerscore "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	// current controller name.
	controllerAgentName = "rate-webterminal"

	// SuccessSynced is used as part of the Event 'reason' when a pod is synced.
	SuccessSynced = "Synced"
	// MessagePodSynced is the message used for an Event fired when a pod
	// is synced Successfully.
	MessagePodSynced = "Pod synced successfully"
)

type Controller struct {
	/// clientset is a standard kubernetes clientset.
	clientset kubernetes.Interface
	podLister listerscore.PodLister
	// podSynced is a flag to determine if pod informer had been synced.
	podSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of  performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a time,
	// and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is a event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(clientset kubernetes.Interface,
	podInformer cache.SharedIndexInformer,
	podLister listerscore.PodLister) *Controller {
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		clientset: clientset,
		podLister: podLister,
		podSynced: podInformer.HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pod"),
		recorder:  recorder,
	}

	return controller
}
