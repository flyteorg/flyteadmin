package interfaces

import "k8s.io/apimachinery/pkg/api/resource"

type DeprecatedTaskResourceSet struct {
	CPU              string `json:"cpu"`
	GPU              string `json:"gpu"`
	Memory           string `json:"memory"`
	Storage          string `json:"storage"`
	EphemeralStorage string `json:"ephemeralStorage"`
}

type TaskResourceSet struct {
	CPU              resource.Quantity `json:"cpu"`
	GPU              resource.Quantity `json:"gpu"`
	Memory           resource.Quantity `json:"memory"`
	Storage          resource.Quantity `json:"storage"`
	EphemeralStorage resource.Quantity `json:"ephemeralStorage"`
}

// Provides default values for task resource limits and defaults.
type TaskResourceConfiguration interface {
	GetDefaults() TaskResourceSet
	GetLimits() TaskResourceSet
}
