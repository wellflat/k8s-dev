package main

import (
	"context"
	"fmt"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Kubernetes設定ファイルのパスを取得
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	// Kubernetesクライアントを設定
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	// すべてのノードを取得
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing nodes: %v", err)
	}

	for _, node := range nodes.Items {
		fmt.Printf("Node: %s\n", node.Name)

		// NodeのラベルからGPU製品名を取得
		if gpuProduct, ok := node.Labels["nvidia.com/gpu.product"]; ok {
			fmt.Printf("  GPU Product Name: %s\n", gpuProduct)
		} else {
			fmt.Println("  No GPU product label found.")
		}
	}
}