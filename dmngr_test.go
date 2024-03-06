package dmngr

import (
	"fmt"
	"testing"
)

func TestGetPodRestartTime(t *testing.T) {
	restartTime, err := GetPodRestartTime("gke_omiq-dev_us-central1-a_dev-cluster-us", "default", "omiq-api-0")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
	}

	fmt.Printf("Pod %s in the namespace of %s, is restarted at: %v\n", "opmiq-api-0", "default", restartTime)
}

func TestGetLastImageUpdateTime(t *testing.T) {
	const namespace = "default"
	const resourceName = "omiq-api"
	const resourceType = "statefulset" // "deployment"

	lastImageUpdateTime, currentImageVersion, err := GetLastImageUpdateTime("gke_omiq-dev_us-central1-a_dev-cluster-us", namespace, resourceName, resourceType)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to get the timestamp from GetPodRestartTime: %v", err))
	}

	fmt.Printf("The image %s of the resource type %s in the namespace of %s is %s, updated at %s\n", currentImageVersion, resourceType, namespace, resourceName, lastImageUpdateTime)
}
