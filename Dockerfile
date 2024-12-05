FROM alpine:3.17.2

COPY kubernetes-toolkit /usr/bin/

ENTRYPOINT ["/usr/bin/kubernetes-toolkit"]
