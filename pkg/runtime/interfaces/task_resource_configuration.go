package interfaces

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"k8s.io/apimachinery/pkg/api/resource"
)

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
	GetAsAttribute() admin.TaskResourceAttributes
}
