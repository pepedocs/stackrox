// Code generated by MockGen. DO NOT EDIT.
// Source: reprocessor.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLoop is a mock of Loop interface
type MockLoop struct {
	ctrl     *gomock.Controller
	recorder *MockLoopMockRecorder
}

// MockLoopMockRecorder is the mock recorder for MockLoop
type MockLoopMockRecorder struct {
	mock *MockLoop
}

// NewMockLoop creates a new mock instance
func NewMockLoop(ctrl *gomock.Controller) *MockLoop {
	mock := &MockLoop{ctrl: ctrl}
	mock.recorder = &MockLoopMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLoop) EXPECT() *MockLoopMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockLoop) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockLoopMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockLoop)(nil).Start))
}

// ShortCircuit mocks base method
func (m *MockLoop) ShortCircuit() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ShortCircuit")
}

// ShortCircuit indicates an expected call of ShortCircuit
func (mr *MockLoopMockRecorder) ShortCircuit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortCircuit", reflect.TypeOf((*MockLoop)(nil).ShortCircuit))
}

// Stop mocks base method
func (m *MockLoop) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockLoopMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockLoop)(nil).Stop))
}
