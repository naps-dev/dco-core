package test

import (
	v1 "k8s.io/api/core/v1"
)

func isEqual(expected map[string]bool, actual map[string]bool) bool {
	if len(expected) != len(actual) {
		return false
	}
	for k, v := range expected {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func isPodRunningOnAgent(pod v1.Pod, agent *v1.Node) bool {
	return pod.Spec.NodeName == agent.Name
}
