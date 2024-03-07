package dmngr

import (
	"context"
	"fmt"
	"testing"
)

// func TestGetPodRestartTime(t *testing.T) {
// 	const cluster = "gke_omiq-dev_us-central1-a_dev-cluster-us"
// 	const namespace = "default"
// 	const podName = "omiq-api-0"
// 	restartTime, err := GetPodRestartTime(cluster, namespace, podName)
// 	if err != nil {
// 		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
// 	}

// 	fmt.Printf("Pod %s in the namespace of %s, is restarted at: %v\n", "opmiq-api-0", "default", restartTime)
// }

// func TestGetLastImageUpdateTime(t *testing.T) {
// 	const cluster = "gke_omiq-dev_us-central1-a_dev-cluster-us"
// 	const namespace = "default"
// 	const resourceName = "omiq-api"

// 	lastImageUpdateTime, currentImageVersion, err := GetLastImageUpdateTime(cluster, namespace, resourceName, StatefulSetsString)
// 	if err != nil {
// 		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
// 	}

// 	fmt.Printf("The image %s of the resource type %s in the namespace of %s is %s, updated at %s\n", currentImageVersion, StatefulSetsString, namespace, resourceName, lastImageUpdateTime)
// }

// func TestGetLastLogTime(t *testing.T) {
// 	const cluster = "gke_omiq-dev_us-central1-a_dev-cluster-us"
// 	const namespace = "default"
// 	const resourceName = "omiq-api-0"

// 	lastImageUpdateTime, err := GetLastLogTime(cluster, namespace, resourceName)
// 	if err != nil {
// 		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
// 	}

// 	fmt.Println(lastImageUpdateTime)
// }

// func TestGetAllKcontext(t *testing.T) {
// 	clusters := GetAllKcontext()
// 	fmt.Println(clusters)
// }

// func TestUpdateImageStatefulSet(t *testing.T) {
// 	const cluster = "gke_omiq-dev_us-central1-a_dev-cluster-us"
// 	const namespace = "default"
// 	const resourceName = "omiq-api"
// 	const image = "gcr.io/omiq-dev/api:pr-1505" // 1507
// 	const dryrun = false
// 	err := updateImage(context.TODO(), cluster, resourceName, namespace, image, StatefulSetsString, dryrun)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

// func TestUpdateImageDeployment(t *testing.T) {
// 	const cluster = "gke_omiq-dev_us-central1-a_dev-cluster-us"
// 	const namespace = "default"
// 	const resourceName = "webapp"
// 	const image = "gcr.io/omiq-dev/webapp:latest"
// 	const dryrun = true
// 	err := updateImage(context.TODO(), cluster, resourceName, namespace, image, DeploymentString, dryrun)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

func TestGetAllClustersInfo(t *testing.T) {
	fmt.Println(getAllClustersInfo(context.TODO()))
}
