#!/bin/sh

CILIUM_POD=$(kubectl get pods -n kube-system -l k8s-app=cilium -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n kube-system $CILIUM_POD -- cilium fqdn cache list