package election

import (
	"context"
	"os"
	"time"

	"github.com/forbearing/k8s/pod"
	"github.com/forbearing/ratel-webterminal/pkg/args"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	leaseLockName      = "ratel-webterminal"
	leaseLockNamespace = os.Getenv("NAMESPACE")
	id                 = os.Getenv("NAME")
)

var lock *resourcelock.LeaseLock

func Init() {
	handler, err := pod.New(context.TODO(), args.GetKubeConfigFile(), "")
	if err != nil {
		log.Fatalf("create handler failed in election: %s", err.Error())
	}
	if len(leaseLockNamespace) == 0 {
		log.Fatal(`ratel-webterminal require a "NAMESPACE" environment variable`)
	}
	if len(id) == 0 {
		id = uuid.New().String()
	}

	// we use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	lock = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseLockName,
			Namespace: leaseLockNamespace,
		},
		Client: handler.Clientset().CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	// start the leader election code loop

}

func Run(ctx context.Context, run func(ctx context.Context)) {
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// run ratel-webterminal
				run(ctx)
			},
			OnStoppedLeading: func() {
				// start cleanup
				log.Infof("leader lost: %s", id)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == id {
					// I just got the lock
					return
				}
				log.Infof("new leader elected: %s", identity)
			},
		},
	})
}
