package main

import (
	"fmt"

	"github.com/paul-freeman/deployerai/pkg/operationutils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: "gke_omiq-dev_us-central1-a_dev-cluster-us"},
	)

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to load the config: %v", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read the config: %v", err))
	}

	restartTime, err := operationutils.GetPodRestartTime(clientset, "default", "omiq-api-0")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
	}
	operationutils.GetAllPods(clientset, "default")

	fmt.Printf("Pod %s in the namespace of %s, is restarted at: %v\n", "opmiq-api-0", "default", restartTime)
}
