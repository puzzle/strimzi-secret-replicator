# strimzi-secret-replicator ![Docker](https://github.com/puzzle/strimzi-secret-replicator/workflows/Docker/badge.svg?branch=master)
The `strimzi-secret-replicator` replicates secrets which belong to a Strimzi `KafkaUser` to other namespaces.

## Usage
* Run `strimzi-secret-replicator`
```
./strimzi-secret-replicator
```

## Development Setup
```
kind create cluster

helm repo add strimzi https://strimzi.io/charts/
kubectl create ns kafka
helm install strimzi --namespace kafka strimzi/strimzi-kafka-operator

kubectl -n kafka apply -f resources/cluster.yaml # then be patient
kubectl -n kafka apply -f resources/topic.yaml
kubectl -n kafka apply -f resources/user.yaml
```

Then annotate the user and observe the actions of the operator
```
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
