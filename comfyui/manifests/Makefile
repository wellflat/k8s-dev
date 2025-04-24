MODEL_PATH ?= /home/ryo/workspace/dev/k8s-dev/comfyui/ComfyUI/models

# This Makefile is used to build and run ComfyUI in a Kubernetes cluster.
# Cluster
cluster:
	@[ -d "$(MODEL_PATH)" ] || { echo "Please set MODEL_PATH"; exit 1; }
	minikube start --driver docker --container-runtime docker \
		--memory 8192MB --cpus 4 --gpus all --mount \
		--mount-string $(MODEL_PATH):/minikube-host/models
	# we use custom nvidia-device-plugin helm chart to enable GPU sharing.
	#minikube addons disable nvidia-device-plugin

cluster-delete:
	minikube delete