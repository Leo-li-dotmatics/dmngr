package dmngr

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type WorkloadType string

const (
	DeploymentString   WorkloadType = "deployment"
	StatefulSetsString WorkloadType = "statefulset"
)
const containerName = "backend"

func loadKubernetesConfig(kContext string) *kubernetes.Clientset {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kContext},
	)

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to load the config: %v", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read the config: %v", err))
	}
	return clientset
}

func GetPodRestartTime(kcontext, namespace, podName string) (time.Time, error) {

	clientset := loadKubernetesConfig(kcontext)

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting pod description: %s\n", err.Error())
		return time.Time{}, err
	}
	if pod.Status.StartTime != nil {
		return pod.Status.StartTime.Time, nil
	}
	return pod.CreationTimestamp.Time, nil
}

func GetLastLogTime(kcontext, namespace, podName string) (time.Time, error) {
	var lastLogTimeString string
	var lastLogTime time.Time

	clientset := loadKubernetesConfig(kcontext)
	podLogs, err := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: containerName, Timestamps: true}).Stream(context.Background())
	if err != nil {
		fmt.Printf("Error getting pod logs: %s\n", err)
		return lastLogTime, nil
	}

	defer podLogs.Close()

	buf := make([]byte, 4096)
	for {
		bytesRead, err := podLogs.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Printf("Error reading pod logs: %s\n", err)
			break
		}
		logStream := string(buf[:bytesRead])

		lines := strings.Split(logStream, "\n")
		for _, line := range lines {
			if strings.Contains(line, "\"UserID\"") {
				lastLogTimeString = line
			}
		}
		fmt.Print()
	}

	lastLogTime, err = time.Parse(time.RFC3339, strings.Split(lastLogTimeString, " ")[0])
	if err != nil {
		fmt.Println(err)
		return lastLogTime, nil
	}

	return lastLogTime, nil
}

func GetLastImageUpdateTime(kcontext, namespace, resourceName string, resourceType WorkloadType) (time.Time, string, error) {
	clientset := loadKubernetesConfig(kcontext)
	var lastImageUpdateTime time.Time
	var currentImageVersion string

	switch resourceType {
	case DeploymentString:
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return time.Time{}, "", err
		}
		lastImageUpdateTime = deployment.CreationTimestamp.Time
		currentImageVersion = deployment.Spec.Template.Spec.Containers[0].Image
	case StatefulSetsString:
		statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return time.Time{}, "", err
		}
		lastImageUpdateTime = statefulSet.CreationTimestamp.Time
		currentImageVersion = statefulSet.Spec.Template.Spec.Containers[1].Image
	default:
		return time.Time{}, "", fmt.Errorf("invalid resource type: %s", resourceType)
	}

	return lastImageUpdateTime, currentImageVersion, nil
}

func UpdateImage(kcontext, resourceName, namespace, image string, resourceType WorkloadType, dryrun bool) error {
	var err error

	clientset := loadKubernetesConfig(kcontext)

	switch resourceType {
	case DeploymentString:
		err = updateDeploymentImage(clientset, resourceName, namespace, image, dryrun)
	case StatefulSetsString:
		err = updateStatefulsetImage(clientset, resourceName, namespace, image, dryrun)
	}
	if err != nil {
		return err
	}
	return nil
}

func updateDeploymentImage(clientset *kubernetes.Clientset, resourceName, namespace, image string, dryrun bool) error {
	updateOptions := metav1.UpdateOptions{}
	if dryrun {
		updateOptions.DryRun = []string{metav1.DryRunAll}
	}
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	newDeployment, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, updateOptions)
	if err != nil {
		return err
	}

	if dryrun {
		jsonData, err := json.MarshalIndent(newDeployment, "", "  ")
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		fmt.Println("New Deployment:", string(jsonData))
	}
	return nil
}

func updateStatefulsetImage(clientset *kubernetes.Clientset, resourceName, namespace, image string, dryrun bool) error {
	updateOptions := metav1.UpdateOptions{}
	if dryrun {
		updateOptions.DryRun = []string{metav1.DryRunAll}
	}

	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	statefulSet.Spec.Template.Spec.Containers[1].Image = image

	newStatefulSet, err := clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, updateOptions)
	if err != nil {
		return err
	}
	if dryrun {
		jsonData, err := json.MarshalIndent(newStatefulSet, "", "  ")
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}

		fmt.Println("New Deployment:", string(jsonData))
	}

	return nil
}

func GetAllKcontext() []string {
	clusters := make([]string, 0)

	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		fmt.Printf("Error building config: %s\n", err)
		return clusters
	}
	for _, context := range config.Contexts {
		clusters = append(clusters, context.Cluster)
	}
	return clusters
}
