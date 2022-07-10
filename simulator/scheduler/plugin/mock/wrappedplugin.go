// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin (interfaces: Store,FilterPluginExtender,ScorePluginExtender,NormalizeScorePluginExtender)

// Package plugin is a generated GoMock package.
package plugin

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
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

// AddFilterResult mocks base method.
func (m *MockStore) AddFilterResult(arg0, arg1, arg2, arg3, arg4 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddFilterResult", arg0, arg1, arg2, arg3, arg4)
}

// AddFilterResult indicates an expected call of AddFilterResult.
func (mr *MockStoreMockRecorder) AddFilterResult(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFilterResult", reflect.TypeOf((*MockStore)(nil).AddFilterResult), arg0, arg1, arg2, arg3, arg4)
}

// AddNormalizedScoreResult mocks base method.
func (m *MockStore) AddNormalizedScoreResult(arg0, arg1, arg2, arg3 string, arg4 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddNormalizedScoreResult", arg0, arg1, arg2, arg3, arg4)
}

// AddNormalizedScoreResult indicates an expected call of AddNormalizedScoreResult.
func (mr *MockStoreMockRecorder) AddNormalizedScoreResult(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNormalizedScoreResult", reflect.TypeOf((*MockStore)(nil).AddNormalizedScoreResult), arg0, arg1, arg2, arg3, arg4)
}

// AddScoreResult mocks base method.
func (m *MockStore) AddScoreResult(arg0, arg1, arg2, arg3 string, arg4 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddScoreResult", arg0, arg1, arg2, arg3, arg4)
}

// AddScoreResult indicates an expected call of AddScoreResult.
func (mr *MockStoreMockRecorder) AddScoreResult(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddScoreResult", reflect.TypeOf((*MockStore)(nil).AddScoreResult), arg0, arg1, arg2, arg3, arg4)
}

// MockFilterPluginExtender is a mock of FilterPluginExtender interface.
type MockFilterPluginExtender struct {
	ctrl     *gomock.Controller
	recorder *MockFilterPluginExtenderMockRecorder
}

// MockFilterPluginExtenderMockRecorder is the mock recorder for MockFilterPluginExtender.
type MockFilterPluginExtenderMockRecorder struct {
	mock *MockFilterPluginExtender
}

// NewMockFilterPluginExtender creates a new mock instance.
func NewMockFilterPluginExtender(ctrl *gomock.Controller) *MockFilterPluginExtender {
	mock := &MockFilterPluginExtender{ctrl: ctrl}
	mock.recorder = &MockFilterPluginExtenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFilterPluginExtender) EXPECT() *MockFilterPluginExtenderMockRecorder {
	return m.recorder
}

// AfterFilter mocks base method.
func (m *MockFilterPluginExtender) AfterFilter(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 *framework.NodeInfo, arg4 *framework.Status) *framework.Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterFilter", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*framework.Status)
	return ret0
}

// AfterFilter indicates an expected call of AfterFilter.
func (mr *MockFilterPluginExtenderMockRecorder) AfterFilter(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterFilter", reflect.TypeOf((*MockFilterPluginExtender)(nil).AfterFilter), arg0, arg1, arg2, arg3, arg4)
}

// BeforeFilter mocks base method.
func (m *MockFilterPluginExtender) BeforeFilter(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 *framework.NodeInfo) *framework.Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeforeFilter", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*framework.Status)
	return ret0
}

// BeforeFilter indicates an expected call of BeforeFilter.
func (mr *MockFilterPluginExtenderMockRecorder) BeforeFilter(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeFilter", reflect.TypeOf((*MockFilterPluginExtender)(nil).BeforeFilter), arg0, arg1, arg2, arg3)
}

// MockScorePluginExtender is a mock of ScorePluginExtender interface.
type MockScorePluginExtender struct {
	ctrl     *gomock.Controller
	recorder *MockScorePluginExtenderMockRecorder
}

// MockScorePluginExtenderMockRecorder is the mock recorder for MockScorePluginExtender.
type MockScorePluginExtenderMockRecorder struct {
	mock *MockScorePluginExtender
}

// NewMockScorePluginExtender creates a new mock instance.
func NewMockScorePluginExtender(ctrl *gomock.Controller) *MockScorePluginExtender {
	mock := &MockScorePluginExtender{ctrl: ctrl}
	mock.recorder = &MockScorePluginExtenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScorePluginExtender) EXPECT() *MockScorePluginExtenderMockRecorder {
	return m.recorder
}

// AfterScore mocks base method.
func (m *MockScorePluginExtender) AfterScore(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 string, arg4 int64, arg5 *framework.Status) (int64, *framework.Status) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterScore", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(*framework.Status)
	return ret0, ret1
}

// AfterScore indicates an expected call of AfterScore.
func (mr *MockScorePluginExtenderMockRecorder) AfterScore(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterScore", reflect.TypeOf((*MockScorePluginExtender)(nil).AfterScore), arg0, arg1, arg2, arg3, arg4, arg5)
}

// BeforeScore mocks base method.
func (m *MockScorePluginExtender) BeforeScore(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 string) (int64, *framework.Status) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeforeScore", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(*framework.Status)
	return ret0, ret1
}

// BeforeScore indicates an expected call of BeforeScore.
func (mr *MockScorePluginExtenderMockRecorder) BeforeScore(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeScore", reflect.TypeOf((*MockScorePluginExtender)(nil).BeforeScore), arg0, arg1, arg2, arg3)
}

// MockNormalizeScorePluginExtender is a mock of NormalizeScorePluginExtender interface.
type MockNormalizeScorePluginExtender struct {
	ctrl     *gomock.Controller
	recorder *MockNormalizeScorePluginExtenderMockRecorder
}

// MockNormalizeScorePluginExtenderMockRecorder is the mock recorder for MockNormalizeScorePluginExtender.
type MockNormalizeScorePluginExtenderMockRecorder struct {
	mock *MockNormalizeScorePluginExtender
}

// NewMockNormalizeScorePluginExtender creates a new mock instance.
func NewMockNormalizeScorePluginExtender(ctrl *gomock.Controller) *MockNormalizeScorePluginExtender {
	mock := &MockNormalizeScorePluginExtender{ctrl: ctrl}
	mock.recorder = &MockNormalizeScorePluginExtenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNormalizeScorePluginExtender) EXPECT() *MockNormalizeScorePluginExtenderMockRecorder {
	return m.recorder
}

// AfterNormalizeScore mocks base method.
func (m *MockNormalizeScorePluginExtender) AfterNormalizeScore(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 framework.NodeScoreList, arg4 *framework.Status) *framework.Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AfterNormalizeScore", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*framework.Status)
	return ret0
}

// AfterNormalizeScore indicates an expected call of AfterNormalizeScore.
func (mr *MockNormalizeScorePluginExtenderMockRecorder) AfterNormalizeScore(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AfterNormalizeScore", reflect.TypeOf((*MockNormalizeScorePluginExtender)(nil).AfterNormalizeScore), arg0, arg1, arg2, arg3, arg4)
}

// BeforeNormalizeScore mocks base method.
func (m *MockNormalizeScorePluginExtender) BeforeNormalizeScore(arg0 context.Context, arg1 *framework.CycleState, arg2 *v1.Pod, arg3 framework.NodeScoreList) *framework.Status {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeforeNormalizeScore", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*framework.Status)
	return ret0
}

// BeforeNormalizeScore indicates an expected call of BeforeNormalizeScore.
func (mr *MockNormalizeScorePluginExtenderMockRecorder) BeforeNormalizeScore(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeforeNormalizeScore", reflect.TypeOf((*MockNormalizeScorePluginExtender)(nil).BeforeNormalizeScore), arg0, arg1, arg2, arg3)
}
