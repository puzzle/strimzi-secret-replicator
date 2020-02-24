FROM alpine:latest
COPY strimzi-secret-replicator /bin/strimzi-secret-replicator
CMD ["/bin/strimzi-secret-replicator"]
