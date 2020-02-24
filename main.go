package main

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	version = "n/a"
	commit  = "n/a"
)

func main() {
	log := logf.Log.WithName("manager")
	ctrl.SetLogger(zap.Logger(true))
	log.Info("starting strimzi-secret-replicator", "version", version, "commit", commit)
	// Setup a Manager
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	// Setup a new controller to reconcile ReplicaSets
	log.Info("setting up replicator")
	c, err := controller.New("secret-replicator-controller", mgr, controller.Options{
		Reconciler: &secretReplicator{
			mgr.GetClient(),
			logf.Log.WithName("secret-replicator"),
		},
	})
	if err != nil {
		log.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	// enqueue for secrets which have kafka user as owner ref
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{OwnerType: initKafkaUser()})
	if err != nil {
		log.Error(err, "unable to watch secrets")
		os.Exit(1)
	}

	// enqueue for kafka users
	if err := c.Watch(&source.Kind{Type: initKafkaUser()}, &handler.EnqueueRequestForObject{}); err != nil {
		log.Error(err, "unable to watch kafka users")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
