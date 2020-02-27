package main

import (
	"context"
	"strings"

	logr "github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type secretReplicator struct {
	client.Client
	log logr.Logger
}

var _ reconcile.Reconciler = &secretReplicator{}

var (
	toNamespaceAnnotation = "secret-replicator.k8s.puzzle.ch/to-namespace"
	ownerAnnotation       = "secret-replicator.k8s.puzzle.ch/owner"
)

func (sr *secretReplicator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := sr.log.WithValues("namespace", request.Namespace, "name", request.Name)

	log.V(4).Info("start reconcile")
	secret := &corev1.Secret{}
	if err := sr.Get(ctx, request.NamespacedName, secret); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	kafkaUser := initKafkaUser()
	if err := sr.Get(ctx, request.NamespacedName, kafkaUser); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	targetNamespaces, found := kafkaUser.GetAnnotations()[toNamespaceAnnotation]
	if !found {
		log.V(4).Info("skip because there is no annotation")
		return reconcile.Result{}, nil
	}

	namespaces := strings.Split(targetNamespaces, ",")
	log.V(4).Info("start replication", "namespaces", namespaces)
	errs := []error{}
	for _, namespace := range namespaces {
		toSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secret.Name, Namespace: namespace}}
		op, err := controllerutil.CreateOrUpdate(ctx, sr, toSecret, func() error {
			toSecret.Annotations[ownerAnnotation] = "true"
			toSecret.Data = secret.Data
			return nil
		})

		if err != nil {
			log.Error(err, "could not replicate secret", "targetNamespace", namespace)
			errs = append(errs, err)
		} else {
			if op == controllerutil.OperationResultNone {
				log.V(4).Info("secret unchanged", "targetNamespace", namespace)
			} else {
				log.Info("secret replicated", "operation", op, "targetNamespace", namespace)
			}
		}
	}

	if len(errs) == 0 {
		return reconcile.Result{}, nil
	}

	// TODO: show all errors
	log.Error(errs[0], "failed to replicate")
	return reconcile.Result{}, errs[0]
}

func initKafkaUser() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaUser",
		Version: "v1beta1",
	})
	return u
}
