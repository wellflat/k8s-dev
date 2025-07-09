#!/bin/sh

kubectl get secret argo-server-token -n argo -o jsonpath='{.data.token}' | base64 --decode