# strimzi-secret-replicator
[![Go Report Card](https://goreportcard.com/badge/github.com/puzzle/strimzi-secret-replicator)](https://goreportcard.com/report/github.com/puzzle/strimzi-secret-replicator)
[![Build](https://github.com/puzzle/strimzi-secret-replicator/workflows/Build/badge.svg?branch=master)](https://github.com/puzzle/strimzi-secret-replicator/actions)
[![Dockerhub](https://img.shields.io/docker/pulls/puzzle/strimzi-secret-replicator.svg)](https://hub.docker.com/repository/docker/puzzle/strimzi-secret-replicator)

The `strimzi-secret-replicator` allows to replicate secrets of `KafkaUser`s to other namespaces. This is intended in environments where you want to give applications in certain namespaces access to a KafkaCluster without giving them permission to read secrets in the namespace where the KafkaUsers are created.

To enable the replication for a KafkaUser the annotation `secret-replicator.k8s.puzzle.ch/to-namespace=target-namespace` has to be set on that KafkaUser.

## Development
### Quickstart
* Build
```
make build
```
* Run
```
./strimzi-secret-replicator
```

### Full Setup
To test the full setup together with Strimzi you need a test cluster where Strimzi is installed. The follwing example shows a setup with [kind](https://kind.sigs.k8s.io/) but the same can be achieved with minikube/minishift or any other cluster.
```shell
kind create cluster # or use minikube/minishift or other test cluster

helm repo add strimzi https://strimzi.io/charts/
kubectl create ns kafka
helm install strimzi --namespace kafka strimzi/strimzi-kafka-operator

kubectl -n kafka apply -f resources/cluster.yaml # then be patient
kubectl -n kafka apply -f resources/topic.yaml
kubectl -n kafka apply -f resources/user.yaml
```

Make sure that your cluster configuration (`KUBECONFIG` or `~/.kube/config`) points to your desired cluster and start the `strimzi-secret-replicator`.
```shell
./strimzi-secret-replicator
```

Then annotate the user and observe the actions of the operator
```shell
# create target namespaces
kubectl create ns foo
kubectl create ns bla

# first only to one
kubectl -n kafka annotate kafkausers.kafka.strimzi.io my-user secret-replicator.k8s.puzzle.ch/to-namespace=foo
# add second namespace
kubectl -n kafka annotate --overwrite kafkausers.kafka.strimzi.io my-user secret-replicator.k8s.puzzle.ch/to-namespace=foo,bla

# remove annotation
kubectl -n kafka annotate kafkausers.kafka.strimzi.io my-user secret-replicator.k8s.puzzle.ch/to-namespace-
```

## Open points / thoughts
* Remove namespace from annotation or remove annotation completley
  * We do not necessarly have to remove the secrets from namespaces because we already leaked that secret and even if we remove it we can not be sure that is already copied somewhere else. So we have to ensure in a other way that the secret is no longer valid (probably deleting the entire KafkaUser would be always the best option if this does revoke the certificate)
  * Probably it would be the best to replicate one secret always only to one namespace. Then we can delete a KafkaUser and only affect one namespace.
* Secret has different name than KafkaUser -> currently not handeld, is this even possible?
* Describe/prepare Installation resources
  * kubectl apply
  * kustomize
  * helm
