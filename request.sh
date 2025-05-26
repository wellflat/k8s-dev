#!/bin/sh

# service経由
curl $(minikube service hello-service --url)

# ingress経由
curl http://$(minikube ip) -H 'Host: mini.trigkey.local'


portforward() {
    kubectl port-forward --address 0.0.0.0 service/hello-service 8888:80
}