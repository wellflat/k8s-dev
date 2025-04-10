#!/bin/sh

curl $(minikube service hello-service --url)

portforward() {
    kubectl port-forward --address 0.0.0.0 service/hello-service 8888:80
}