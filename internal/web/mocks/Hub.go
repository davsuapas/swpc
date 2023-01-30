// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	sockets "github.com/swpoolcontroller/pkg/sockets"
)

// Hub is an autogenerated mock type for the Hub type
type Hub struct {
	mock.Mock
}

// Register provides a mock function with given fields: client
func (_m *Hub) Register(client sockets.Client) {
	_m.Called(client)
}

// Unregister provides a mock function with given fields: id
func (_m *Hub) Unregister(id string) {
	_m.Called(id)
}

type mockConstructorTestingTNewHub interface {
	mock.TestingT
	Cleanup(func())
}

// NewHub creates a new instance of Hub. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewHub(t mockConstructorTestingTNewHub) *Hub {
	mock := &Hub{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
