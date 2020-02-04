package main

import (
	"context"
	"log"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type secretReplicator struct {
	client.Client
}

var _ reconcile.Reconciler = &secretReplicator{}

var (
	toNamespaceAnnotation = "secret-replicator.k8s.puzzle.ch/to-namespace"
	ownerAnnotation       = "secret-replicator.k8s.puzzle.ch/owner"
)

func (sr *secretReplicator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	log.Println("start reconcile for secret", request)
	secret := &corev1.Secret{}
	if err := sr.Get(ctx, request.NamespacedName, secret); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// maybe also have to take other names into account
	kafkaUser := createKafkaUser()

	if err := sr.Get(ctx, request.NamespacedName, kafkaUser); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	targetNamespaces, found := kafkaUser.GetAnnotations()[toNamespaceAnnotation]
	if !found {
		// TODO: check created by annotation to recreate replicated secrets which got deleted
		log.Println("do not replicate user", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	namespaces := strings.Split(targetNamespaces, ",")
	log.Println("replicate to:", namespaces)

	errs := []error{}
	for _, namespace := range namespaces {
		toSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secret.Name, Namespace: namespace}}
		op, err := controllerutil.CreateOrUpdate(ctx, sr, toSecret, func() error {
			toSecret.Data = secret.Data
			return nil
		})

		if err != nil {
			errs = append(errs, err)
		} else {
			log.Printf("%s: %s/%s\n", op, namespace, secret.Name)
		}
	}

	if len(errs) == 0 {
		return reconcile.Result{}, nil
	}

	// TODO: show all errors
	log.Println(errs[0])
	return reconcile.Result{}, errs[0]
}

func createKafkaUser() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaUser",
		Version: "v1beta1",
	})
	return u
}

func main() {
	// Setup a Manager
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Println("unable to set up controller manager", err)
		os.Exit(1)
	}

	// Setup a new controller to reconcile ReplicaSets
	log.Println("Setting up controller")
	c, err := controller.New("secret-replicator-controller", mgr, controller.Options{
		Reconciler: &secretReplicator{mgr.GetClient()},
	})
	if err != nil {
		log.Println("unable to set up individual controller", err)
		os.Exit(1)
	}

	// enqueue for secrets which have kafka user as owner ref
	if err := c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{OwnerType: createKafkaUser()}); err != nil {
		log.Println("unable to watch secrets", err)
		os.Exit(1)
	}

	// enqueue for kafka users
	if err := c.Watch(&source.Kind{Type: createKafkaUser()}, &handler.EnqueueRequestForObject{}); err != nil {
		log.Println("unable to watch kafka users", err)
		os.Exit(1)
	}

	log.Println("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Println("unable to run manager", err)
		os.Exit(1)
	}
}
