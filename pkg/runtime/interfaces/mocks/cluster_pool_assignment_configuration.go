// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	interfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	mock "github.com/stretchr/testify/mock"
)

// ClusterPoolAssignmentConfiguration is an autogenerated mock type for the ClusterPoolAssignmentConfiguration type
type ClusterPoolAssignmentConfiguration struct {
	mock.Mock
}

type ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments struct {
	*mock.Call
}

func (_m ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments) Return(_a0 map[string]interfaces.ClusterPoolAssignment) *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments {
	return &ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments{Call: _m.Call.Return(_a0)}
}

func (_m *ClusterPoolAssignmentConfiguration) OnGetClusterPoolAssignments() *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments {
	c_call := _m.On("GetClusterPoolAssignments")
	return &ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments{Call: c_call}
}

func (_m *ClusterPoolAssignmentConfiguration) OnGetClusterPoolAssignmentsMatch(matchers ...interface{}) *ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments {
	c_call := _m.On("GetClusterPoolAssignments", matchers...)
	return &ClusterPoolAssignmentConfiguration_GetClusterPoolAssignments{Call: c_call}
}

// GetClusterPoolAssignments provides a mock function with given fields:
func (_m *ClusterPoolAssignmentConfiguration) GetClusterPoolAssignments() map[string]interfaces.ClusterPoolAssignment {
	ret := _m.Called()

	var r0 map[string]interfaces.ClusterPoolAssignment
	if rf, ok := ret.Get(0).(func() map[string]interfaces.ClusterPoolAssignment); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interfaces.ClusterPoolAssignment)
		}
	}

	return r0
}
