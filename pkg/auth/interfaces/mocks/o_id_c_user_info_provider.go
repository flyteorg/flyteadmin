// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	service "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

// OIdCUserInfoProvider is an autogenerated mock type for the OIdCUserInfoProvider type
type OIdCUserInfoProvider struct {
	mock.Mock
}

type OIdCUserInfoProvider_UserInfo struct {
	*mock.Call
}

func (_m OIdCUserInfoProvider_UserInfo) Return(_a0 *service.UserInfoResponse, _a1 error) *OIdCUserInfoProvider_UserInfo {
	return &OIdCUserInfoProvider_UserInfo{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *OIdCUserInfoProvider) OnUserInfo(_a0 context.Context, _a1 *service.UserInfoRequest) *OIdCUserInfoProvider_UserInfo {
	c := _m.On("UserInfo", _a0, _a1)
	return &OIdCUserInfoProvider_UserInfo{Call: c}
}

func (_m *OIdCUserInfoProvider) OnUserInfoMatch(matchers ...interface{}) *OIdCUserInfoProvider_UserInfo {
	c := _m.On("UserInfo", matchers...)
	return &OIdCUserInfoProvider_UserInfo{Call: c}
}

// UserInfo provides a mock function with given fields: _a0, _a1
func (_m *OIdCUserInfoProvider) UserInfo(_a0 context.Context, _a1 *service.UserInfoRequest) (*service.UserInfoResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *service.UserInfoResponse
	if rf, ok := ret.Get(0).(func(context.Context, *service.UserInfoRequest) *service.UserInfoResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.UserInfoResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *service.UserInfoRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
