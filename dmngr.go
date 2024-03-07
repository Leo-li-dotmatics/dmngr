package dmngr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
const timeoutDuration = 60 //seconds

type Error struct {
	Message string
}

type PodRestartTimeResp struct {
	T time.Time
}

// func GetPodRestartTime(ctx context.Context, kcontext, namespace, podName string) tea.Cmd {
// 	return func() tea.Msg {
// 		t, err := getPodRestartTime(ctx, kcontext, namespace, podName)
// 		if err != nil {
// 			return Error{Message: err.Error()}
// 		}
// 		return PodRestartTimeResp{T: t}
// 	}
// }

func getPodRestartTime(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string) (time.Time, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return time.Time{}, err
	}
	if pod.Status.StartTime != nil {
		return pod.Status.StartTime.Time, nil
	}
	return pod.CreationTimestamp.Time, nil
}

type LastLogTimeResp struct {
	T time.Time
}

// func GetLastLogTime(ctx context.Context, kcontext, namespace, podName string) tea.Cmd {
// 	return func() tea.Msg {
// 		timestamp, err := getLastLogTime(ctx, kcontext, namespace, podName)
// 		if err != nil {
// 			return Error{Message: err.Error()}
// 		}

// 		return LastLogTimeResp{T: timestamp}
// 	}
// }

func getLastLogTime(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string) (time.Time, error) {
	var lastLogTimeString string
	var lastLogTime time.Time

	podLogs, err := clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: containerName, Timestamps: true}).Stream(ctx)
	if err != nil {
		return lastLogTime, err
	}

	defer podLogs.Close()

	buf := make([]byte, 4096)
	for {
		bytesRead, err := podLogs.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			break
		}
		logStream := string(buf[:bytesRead])

		lines := strings.Split(logStream, "\n")
		for _, line := range lines {
			if strings.Contains(line, "\"UserID\"") {
				lastLogTimeString = line
			}
		}
	}

	lastLogTime, err = time.Parse(time.RFC3339, strings.Split(lastLogTimeString, " ")[0])
	if err != nil {
		return lastLogTime, err
	}

	return lastLogTime, nil
}

type LastImageUpdateResp struct {
	T       time.Time
	I       string
	Message string
}

// func GetLastImageUpdateTime(ctx context.Context, kcontext, namespace, resourceName string, resourceType WorkloadType) tea.Cmd {
// 	return func() tea.Msg {
// 		timestamp, image, err := getLastImageUpdateTime(ctx, kcontext, namespace, resourceName, resourceType)
// 		if err != nil {
// 			return Error{Message: err.Error()}
// 		}

//			return LastImageUpdateResp{T: timestamp, I: image}
//		}
//	}

func getLastImageUpdateTime(ctx context.Context, clientset *kubernetes.Clientset, namespace, resourceName string, resourceType WorkloadType) (time.Time, string, error) {
	var lastImageUpdateTime time.Time
	var currentImageVersion string

	switch resourceType {
	case DeploymentString:
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return time.Time{}, "", err
		}
		lastImageUpdateTime = deployment.CreationTimestamp.Time
		currentImageVersion = deployment.Spec.Template.Spec.Containers[0].Image
	case StatefulSetsString:
		statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
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

type UpdateImageResp struct {
	Message string
}

func UpdateImage(ctx context.Context, kcontext, resourceName, namespace, image string, resourceType WorkloadType, dryrun bool) tea.Cmd {
	return func() tea.Msg {
		ret, err := updateImage(kcontext, resourceName, namespace, image, resourceType, dryrun)
		if err != nil {
			return Error{Message: err.Error()}
		}

		return UpdateImageResp{Message: ret}
	}
}

func updateImage(kcontext, resourceName, namespace, image string, resourceType WorkloadType, dryrun bool) (string, error) {
	var err error

	clientset, err := loadKubernetesConfig(kcontext)
	ret := ""
	switch resourceType {
	case DeploymentString:
		ret, err = updateDeploymentImage(clientset, resourceName, namespace, image, dryrun)
	case StatefulSetsString:
		ret, err = updateStatefulsetImage(clientset, resourceName, namespace, image, dryrun)
	}
	if err != nil {
		return "", err
	}
	return ret, nil
}

func updateDeploymentImage(clientset *kubernetes.Clientset, resourceName, namespace, image string, dryrun bool) (string, error) {
	startTime := time.Now()
	timeout := timeoutDuration * time.Second

	updateOptions := metav1.UpdateOptions{}
	if dryrun {
		updateOptions.DryRun = []string{metav1.DryRunAll}
	}
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	deployment.Spec.Template.Spec.Containers[0].Image = image

	newDeployment, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, updateOptions)
	if err != nil {
		return "", err
	}

	if dryrun {
		jsonData, err := json.MarshalIndent(newDeployment, "", "  ")
		if err != nil {
			return "", nil
		}
		return string(jsonData), nil
	}

	for {
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Check if the update has been rolled out to all replicas
		if deployment.Status.UpdatedReplicas == deployment.Status.Replicas {
			// Check if all pods are ready
			if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
				break
			}
		}

		if time.Since(startTime) > timeout {
			return "", errors.New("timeout")
		}

		time.Sleep(5 * time.Second)
	}
	return "", nil
}

func updateStatefulsetImage(clientset *kubernetes.Clientset, resourceName, namespace, image string, dryrun bool) (string, error) {
	startTime := time.Now()
	timeout := timeoutDuration * time.Second
	updateOptions := metav1.UpdateOptions{}
	if dryrun {
		updateOptions.DryRun = []string{metav1.DryRunAll}
	}

	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	statefulSet.Spec.Template.Spec.Containers[1].Image = image

	newStatefulSet, err := clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, updateOptions)
	if err != nil {
		return "", err
	}

	if dryrun {
		jsonData, err := json.MarshalIndent(newStatefulSet, "", "  ")
		if err != nil {
			return "", nil
		}
		return string(jsonData), nil
	}

	for {
		newStatefulSet, err = clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if newStatefulSet.Status.UpdatedReplicas == newStatefulSet.Status.Replicas {
			if newStatefulSet.Status.ReadyReplicas == newStatefulSet.Status.Replicas {
				break
			}
		}

		if time.Since(startTime) > timeout {
			return "", errors.New("timeout")
		}

		time.Sleep(5 * time.Second)
	}

	return "", nil
}

type AllKcontextResp struct {
	Clusters []string
}

func GetAllKcontext() tea.Cmd {
	return func() tea.Msg {
		clusters, err := getAllKcontext()
		if err != nil {
			return Error{Message: err.Error()}
		}

		return AllKcontextResp{Clusters: clusters}
	}
}

func getAllKcontext() ([]string, error) {
	clusters := make([]string, 0)

	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return clusters, err
	}
	for _, context := range config.Contexts {
		clusters = append(clusters, context.Cluster)
	}
	return clusters, nil
}

func loadKubernetesConfig(kContext string) (*kubernetes.Clientset, error) {
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kContext},
	)

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

type AllPodsResp struct {
	Pods []string
}

func GetAllPods(ctx context.Context, kcontext, namespace string) tea.Cmd {
	return func() tea.Msg {
		pods, err := getAllPods(ctx, kcontext, namespace)
		if err != nil {
			return Error{Message: err.Error()}
		}
		return AllPodsResp{Pods: pods}
	}
}

func getAllPods(ctx context.Context, kcontext, namespace string) ([]string, error) {
	clientset, err := loadKubernetesConfig(kcontext)
	if err != nil {
		return nil, err
	}

	var pods []string
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return pods, err
	}
	for _, pod := range podList.Items {
		pods = append(pods, pod.Name)
	}
	return pods, nil
}

type Target struct {
	Name            string
	CurrentImage    string
	LastRestart     time.Time
	LastLogTime     time.Time
	LastImageUpdate time.Time
}

type AllClustersInfoResp struct {
	Targets []Target
}

func GetAllClustersInfo(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		ret, err := getAllClustersInfo(ctx)
		if err != nil {
			return Error{Message: err.Error()}
		}
		return AllClustersInfoResp{Targets: ret}
	}
}

func getAllClustersInfo(ctx context.Context) ([]Target, error) {
	const dev = "dev"
	const namespace = "default"
	const webapp = "webapp"
	const api = "omiq-api"
	contextList := make([]string, 0)
	targetList := make([]Target, 0)

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	config, err := kubeconfig.RawConfig()
	if err != nil {
		return []Target{}, err
	}

	// get all contexts
	contexts := config.Contexts
	for contextName := range contexts {
		if strings.Contains(contextName, dev) {
			contextList = append(contextList, contextName)
		}
	}

	// iterate all contexts to get deployment info
	for _, c := range contextList {
		clientset, _ := loadKubernetesConfig(c)
		lastImageUpdateTime, currentImageVersion, err := getLastImageUpdateTime(ctx, clientset, namespace, webapp, DeploymentString)
		if err != nil {
			continue
		}
		podsName, err := getWebPodsName(clientset, namespace, webapp)
		if err != nil {
			continue
		}
		lastPodRestart, err := getPodRestartTime(ctx, clientset, namespace, podsName[0])
		if err != nil {
			lastPodRestart = time.Time{}
		}

		lastLogTime, err := getLastLogTime(ctx, clientset, namespace, podsName[0])
		if err != nil {
			lastLogTime = lastPodRestart
		}

		targetList = append(targetList, Target{
			Name:            webapp,
			CurrentImage:    currentImageVersion,
			LastImageUpdate: lastImageUpdateTime,
			LastLogTime:     lastLogTime,
			LastRestart:     lastPodRestart,
		})
	}

	for _, c := range contextList {
		clientset, _ := loadKubernetesConfig(c)
		lastImageUpdateTime, currentImageVersion, err := getLastImageUpdateTime(ctx, clientset, namespace, api, StatefulSetsString)
		if err != nil {
			continue
		}
		podsName, err := getApiPodsName(clientset, namespace, api)
		if err != nil {
			continue
		}
		lastPodRestart, err := getPodRestartTime(ctx, clientset, namespace, podsName[0])
		if err != nil {
			continue
		}
		lastLogTime, err := getLastLogTime(ctx, clientset, namespace, podsName[0])
		if err != nil {
			lastLogTime = lastPodRestart
		}
		targetList = append(targetList, Target{
			Name:            api,
			CurrentImage:    currentImageVersion,
			LastImageUpdate: lastImageUpdateTime,
			LastRestart:     lastPodRestart,
			LastLogTime:     lastLogTime,
		})
	}

	return targetList, nil

}

func getApiPodsName(clientset *kubernetes.Clientset, namespace, deploymentName string) ([]string, error) {
	podNames := make([]string, 0)
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deploymentName),
	})
	if err != nil {
		return podNames, err
	}
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.GetName())
	}
	return podNames, nil
}

func getWebPodsName(clientset *kubernetes.Clientset, namespace, deploymentName string) ([]string, error) {
	podNames := make([]string, 0)
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("run=%s", deploymentName),
	})
	if err != nil {
		return podNames, err
	}
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.GetName())
	}
	return podNames, nil
}
