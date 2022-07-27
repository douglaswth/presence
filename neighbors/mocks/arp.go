// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	neighbors "douglasthrift.net/presence/neighbors"
	mock "github.com/stretchr/testify/mock"
)

// ARP is an autogenerated mock type for the ARP type
type ARP struct {
	mock.Mock
}

// Count provides a mock function with given fields: count
func (_m *ARP) Count(count uint) {
	_m.Called(count)
}

// Present provides a mock function with given fields: ctx, ifs, state, addrStates
func (_m *ARP) Present(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
	ret := _m.Called(ctx, ifs, state, addrStates)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, neighbors.Interfaces, neighbors.State, neighbors.HardwareAddrStates) error); ok {
		r0 = rf(ctx, ifs, state, addrStates)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewARP interface {
	mock.TestingT
	Cleanup(func())
}

// NewARP creates a new instance of ARP. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewARP(t mockConstructorTestingTNewARP) *ARP {
	mock := &ARP{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}