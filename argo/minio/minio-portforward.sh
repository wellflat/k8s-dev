#!/bin/sh

# mcコマンドでアクセスする場合
#kubectl port-forward --address 0.0.0.0 svc/minio-hl 9000 -n argo
# WebUIにアクセスする場合
kubectl port-forward --address 0.0.0.0 svc/minio-console 9443 -n argo
