package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/ghodss/yaml"
)

func main() {
	// client, err := k8s.NewInClusterClient()

	client, err := loadClient("kubeconfig.yaml")

	if err != nil {
		log.Fatal(err)
	}

	var nodes corev1.NodeList
	if err := client.List(context.Background(), k8s.AllNamespaces, &nodes); err != nil {
		log.Fatal(err)
	}

	var allPods corev1.PodList

	if err := client.List(context.Background(), k8s.AllNamespaces, &allPods); err != nil {
		log.Fatal(err)
	}

	// var foundPod *corev1.Pod
	// for _, pod := range allPods.Items {
	// 	if *pod.Metadata.Name == "shoppingcart-589868587-xbgpv" {
	// 		foundPod = pod
	// 	}
	// }

	// // client.Create(context.Background(), foundPod)
	// client.Delete(context.Background(), foundPod)

	// return

	for _, node := range nodes.Items {
		var pods []*corev1.Pod
		for _, pod := range allPods.Items {
			if *pod.Spec.NodeName == *node.Metadata.Name {
				pods = append(pods, pod)
			}
		}

		fmt.Printf("Node %v\n", *node.Metadata.Name)
		allocatableCPU := cpuReqStrToCPU(*node.Status.Allocatable["cpu"].String_)
		allocatableMemory := memoryReqStrToMemoryMB(*node.Status.Allocatable["memory"].String_)
		fmt.Printf("Allocatable CPU: %vm, memory: %vMi\n", allocatableCPU, allocatableMemory)
		fmt.Printf("Node status: %v\n", node.Metadata)

		podsTotalCPUReq := 0
		podsTotalMemoryReq := 0

		for _, pod := range pods {
			cpuReq := 0
			memoryReq := 0
			for _, container := range pod.Spec.Containers {
				if container.Resources.Requests["cpu"] != nil {
					cpuReqStr := *container.Resources.Requests["cpu"].String_
					cpuReq += cpuReqStrToCPU(cpuReqStr)
				}

				if container.Resources.Requests["memory"] != nil {
					memoryReqStr := *container.Resources.Requests["memory"].String_
					memoryReq += memoryReqStrToMemoryMB(memoryReqStr)
				}
			}

			podsTotalCPUReq += cpuReq
			podsTotalMemoryReq += memoryReq
		}

		fmt.Printf("Pods on node total requests, CPU: %vm, memory: %vMi\n", podsTotalCPUReq, podsTotalMemoryReq)
		fmt.Printf("CPU utilization: %v%%, memory utilization: %v%%\n", float64(podsTotalCPUReq)/float64(allocatableCPU)*100, float64(podsTotalMemoryReq)/float64(allocatableMemory)*100)
		fmt.Printf("\n")
	}
}

func memoryReqStrToMemoryMB(str string) int {
	unit := str[len(str)-2:]
	str = str[:len(str)-2] // For example: 2000Mi
	memory, _ := strconv.Atoi(str)
	switch unit {
	case "Ki":
		return memory / 1024
	case "Mi":
		return memory
	default:
		return 0
	}
}

func cpuReqStrToCPU(str string) int {
	if str[len(str)-1:] == "m" {
		str = str[:len(str)-1] // For example: 1500m
		cpu, _ := strconv.Atoi(str)
		return cpu
	} else {
		coreCount, _ := strconv.Atoi(str) // For example: 3

		return coreCount * 1000
	}
}

// loadClient parses a kubeconfig from a file and returns a Kubernetes client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %v", err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
	}
	return k8s.NewClient(&config)
}

// for _, node := range nodes.Items {
// node := nodes.Items[0]
// fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
// fmt.Printf("capacity=%v\n", node.Status.Capacity)
// fmt.Printf("allocatable=%v\n", node.Status.Allocatable)
// fmt.Printf("phase=%v\n", *node.Status.Phase)
// fmt.Printf("conditions=%v\n", node.Status.Conditions)
// for _, container := range node.Status.Images {
// 	fmt.Printf("%v\n", container)
// }
// }

// pod := pods.Items[0]
// fmt.Printf("name: %v\n", *pod.Metadata.Name)
// fmt.Printf("namespace: %v\n", *pod.Metadata.Namespace)
// fmt.Printf("node: %v\n", *pod.Spec.NodeName)
// fmt.Printf("resource limits: %v, resource requests: %v\n", pod.Spec.Containers[0].Resources.Limits, pod.Spec.Containers[0].Resources.Requests)
// fmt.Printf("resource limits: %v, resource requests: %v\n", pod.Spec.Containers[1].Resources.Limits, pod.Spec.Containers[1].Resources.Requests)
// fmt.Printf("resource limits: %v, resource requests: %v\n", pod.Spec.Containers[2].Resources.Limits, pod.Spec.Containers[2].Resources.Requests)
