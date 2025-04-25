# This Makefile is used to build and run ComfyUI in a Kubernetes cluster.
# Cluster
cluster:
	minikube start --driver docker --container-runtime docker \
		--memory 8192MB --cpus 4 --gpus all

cluster-delete:
	minikube delete