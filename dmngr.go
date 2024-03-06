package dmngr

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type WorkloadType string

const (
	DeploymentString   WorkloadType = "deployment"
	StatefulSetsString WorkloadType = "statefulset"
)

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

func UpdateImage(kcontext, resourceName, namespace, image string, resourceType WorkloadType) error {
	var err error

	clientset := loadKubernetesConfig(kcontext)

	switch resourceType {
	case DeploymentString:
		err = updateDeploymentImage(clientset, resourceName, namespace, image)
	case StatefulSetsString:
		err = updateStatefulsetImage(clientset, resourceName, namespace, image)
	}
	if err != nil {
		return err
	}
	return nil
}

func updateDeploymentImage(clientset *kubernetes.Clientset, resourceName, namespace, image string) error {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	if _, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func updateStatefulsetImage(clientset *kubernetes.Clientset, resourceName, namespace, image string) error {
	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	statefulSet.Spec.Template.Spec.Containers[1].Image = image

	if _, err := clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
