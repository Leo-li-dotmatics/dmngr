package operationutils

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodRestartTime(clientset *kubernetes.Clientset, namespace, podName string) (time.Time, error) {
	// events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
	// 	FieldSelector: fmt.Sprintf("involvedObject.name=%s", "podName"),
	// })
	// if err != nil {
	// 	fmt.Println(fmt.Errorf("failed to get the events: %v", err))
	// 	return time.Time{}, err
	// }

	// var lastRestartTime time.Time
	// for _, event := range events.Items {
	// 	fmt.Println(event)
	// 	if event.Reason == "Started" && event.InvolvedObject.Kind == "Pod" {
	// 		lastRestartTime = event.LastTimestamp.Time
	// 	}
	// }

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting pod description: %s\n", err.Error())
	}
	fmt.Println(pod)
	return pod.CreationTimestamp.Time, nil
}

func GetAllPods(clientset *kubernetes.Clientset, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read the config: %v", err))
	}

	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}
