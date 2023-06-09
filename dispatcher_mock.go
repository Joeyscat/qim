// Code generated by MockGen. DO NOT EDIT.
// Source: dispatcher.go

// Package qim is a generated GoMock package.
package qim

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pkt "github.com/joeyscat/qim/wire/pkt"
)

// MockDispatcher is a mock of Dispatcher interface.
type MockDispatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDispatcherMockRecorder
}

// MockDispatcherMockRecorder is the mock recorder for MockDispatcher.
type MockDispatcherMockRecorder struct {
	mock *MockDispatcher
}

// NewMockDispatcher creates a new mock instance.
func NewMockDispatcher(ctrl *gomock.Controller) *MockDispatcher {
	mock := &MockDispatcher{ctrl: ctrl}
	mock.recorder = &MockDispatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDispatcher) EXPECT() *MockDispatcherMockRecorder {
	return m.recorder
}

// Push mocks base method.
func (m *MockDispatcher) Push(gateway string, channels []string, p *pkt.LogicPkt) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", gateway, channels, p)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push.
func (mr *MockDispatcherMockRecorder) Push(gateway, channels, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockDispatcher)(nil).Push), gateway, channels, p)
}
