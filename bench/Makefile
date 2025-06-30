MODEL_PATH := /home/gpuadmin/.cache/huggingface/hub

cluster:
	@[ -d $(MODEL_PATH) ] || { echo "Please set MODEL_PATH"; exit 1; }
	minikube start --driver docker --container-runtime docker \
		--memory 515900mb --cpus 16 --gpus all --mount \
		--mount-string ${MODEL_PATH}:/minikube-host/.cache/huggingface/hub
	# we use custom nvidia-device-plugin helm chart to enable GPU sharing.
	#minikube addons disable nvidia-device-plugin

cluster-delete:
	minikube delete
