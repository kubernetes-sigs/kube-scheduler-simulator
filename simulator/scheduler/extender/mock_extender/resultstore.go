// Code generated by MockGen. DO NOT EDIT.
// Source: ./resultstore/resultstore.go
//
// Generated by this command:
//
//	mockgen -package=mock_extender -source=./resultstore/resultstore.go -destination=./mock_extender/resultstore.go
//

// Package mock_extender is a generated GoMock package.
package mock_extender

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	v10 "k8s.io/kube-scheduler/extender/v1"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
	isgomock struct{}
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// AddBindResult mocks base method.
func (m *MockStore) AddBindResult(args v10.ExtenderBindingArgs, result v10.ExtenderBindingResult, hostName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddBindResult", args, result, hostName)
}

// AddBindResult indicates an expected call of AddBindResult.
func (mr *MockStoreMockRecorder) AddBindResult(args, result, hostName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBindResult", reflect.TypeOf((*MockStore)(nil).AddBindResult), args, result, hostName)
}

// AddFilterResult mocks base method.
func (m *MockStore) AddFilterResult(args v10.ExtenderArgs, result v10.ExtenderFilterResult, hostName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddFilterResult", args, result, hostName)
}

// AddFilterResult indicates an expected call of AddFilterResult.
func (mr *MockStoreMockRecorder) AddFilterResult(args, result, hostName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFilterResult", reflect.TypeOf((*MockStore)(nil).AddFilterResult), args, result, hostName)
}

// AddPreemptResult mocks base method.
func (m *MockStore) AddPreemptResult(args v10.ExtenderPreemptionArgs, result v10.ExtenderPreemptionResult, hostName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddPreemptResult", args, result, hostName)
}

// AddPreemptResult indicates an expected call of AddPreemptResult.
func (mr *MockStoreMockRecorder) AddPreemptResult(args, result, hostName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPreemptResult", reflect.TypeOf((*MockStore)(nil).AddPreemptResult), args, result, hostName)
}

// AddPrioritizeResult mocks base method.
func (m *MockStore) AddPrioritizeResult(args v10.ExtenderArgs, result v10.HostPriorityList, hostName string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddPrioritizeResult", args, result, hostName)
}

// AddPrioritizeResult indicates an expected call of AddPrioritizeResult.
func (mr *MockStoreMockRecorder) AddPrioritizeResult(args, result, hostName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPrioritizeResult", reflect.TypeOf((*MockStore)(nil).AddPrioritizeResult), args, result, hostName)
}

// DeleteData mocks base method.
func (m *MockStore) DeleteData(pod v1.Pod) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteData", pod)
}

// DeleteData indicates an expected call of DeleteData.
func (mr *MockStoreMockRecorder) DeleteData(pod any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteData", reflect.TypeOf((*MockStore)(nil).DeleteData), pod)
}

// GetStoredResult mocks base method.
func (m *MockStore) GetStoredResult(pod *v1.Pod) map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStoredResult", pod)
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// GetStoredResult indicates an expected call of GetStoredResult.
func (mr *MockStoreMockRecorder) GetStoredResult(pod any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStoredResult", reflect.TypeOf((*MockStore)(nil).GetStoredResult), pod)
}
