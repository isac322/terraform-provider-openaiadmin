// Code generated by MockGen. DO NOT EDIT.
// Source: project_service.go
//
// Generated by this command:
//
//	mockgen -package openai -destination mock_project_service.go -source project_service.go -typed
//

// Package openai is a generated GoMock package.
package openai

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockProjectService is a mock of ProjectService interface.
type MockProjectService struct {
	ctrl     *gomock.Controller
	recorder *MockProjectServiceMockRecorder
	isgomock struct{}
}

// MockProjectServiceMockRecorder is the mock recorder for MockProjectService.
type MockProjectServiceMockRecorder struct {
	mock *MockProjectService
}

// NewMockProjectService creates a new mock instance.
func NewMockProjectService(ctrl *gomock.Controller) *MockProjectService {
	mock := &MockProjectService{ctrl: ctrl}
	mock.recorder = &MockProjectServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProjectService) EXPECT() *MockProjectServiceMockRecorder {
	return m.recorder
}

// Archive mocks base method.
func (m *MockProjectService) Archive(ctx context.Context, projectID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Archive", ctx, projectID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Archive indicates an expected call of Archive.
func (mr *MockProjectServiceMockRecorder) Archive(ctx, projectID any) *MockProjectServiceArchiveCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Archive", reflect.TypeOf((*MockProjectService)(nil).Archive), ctx, projectID)
	return &MockProjectServiceArchiveCall{Call: call}
}

// MockProjectServiceArchiveCall wrap *gomock.Call
type MockProjectServiceArchiveCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectServiceArchiveCall) Return(arg0 error) *MockProjectServiceArchiveCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectServiceArchiveCall) Do(f func(context.Context, string) error) *MockProjectServiceArchiveCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectServiceArchiveCall) DoAndReturn(f func(context.Context, string) error) *MockProjectServiceArchiveCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Create mocks base method.
func (m *MockProjectService) Create(ctx context.Context, name string) (*Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, name)
	ret0, _ := ret[0].(*Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockProjectServiceMockRecorder) Create(ctx, name any) *MockProjectServiceCreateCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockProjectService)(nil).Create), ctx, name)
	return &MockProjectServiceCreateCall{Call: call}
}

// MockProjectServiceCreateCall wrap *gomock.Call
type MockProjectServiceCreateCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectServiceCreateCall) Return(arg0 *Project, arg1 error) *MockProjectServiceCreateCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectServiceCreateCall) Do(f func(context.Context, string) (*Project, error)) *MockProjectServiceCreateCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectServiceCreateCall) DoAndReturn(f func(context.Context, string) (*Project, error)) *MockProjectServiceCreateCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// List mocks base method.
func (m *MockProjectService) List(ctx context.Context) ([]Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockProjectServiceMockRecorder) List(ctx any) *MockProjectServiceListCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockProjectService)(nil).List), ctx)
	return &MockProjectServiceListCall{Call: call}
}

// MockProjectServiceListCall wrap *gomock.Call
type MockProjectServiceListCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectServiceListCall) Return(arg0 []Project, arg1 error) *MockProjectServiceListCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectServiceListCall) Do(f func(context.Context) ([]Project, error)) *MockProjectServiceListCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectServiceListCall) DoAndReturn(f func(context.Context) ([]Project, error)) *MockProjectServiceListCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Modify mocks base method.
func (m *MockProjectService) Modify(ctx context.Context, projectID, name string) (*Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Modify", ctx, projectID, name)
	ret0, _ := ret[0].(*Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Modify indicates an expected call of Modify.
func (mr *MockProjectServiceMockRecorder) Modify(ctx, projectID, name any) *MockProjectServiceModifyCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Modify", reflect.TypeOf((*MockProjectService)(nil).Modify), ctx, projectID, name)
	return &MockProjectServiceModifyCall{Call: call}
}

// MockProjectServiceModifyCall wrap *gomock.Call
type MockProjectServiceModifyCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectServiceModifyCall) Return(arg0 *Project, arg1 error) *MockProjectServiceModifyCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectServiceModifyCall) Do(f func(context.Context, string, string) (*Project, error)) *MockProjectServiceModifyCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectServiceModifyCall) DoAndReturn(f func(context.Context, string, string) (*Project, error)) *MockProjectServiceModifyCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Retrieve mocks base method.
func (m *MockProjectService) Retrieve(ctx context.Context, projectID string) (*Project, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Retrieve", ctx, projectID)
	ret0, _ := ret[0].(*Project)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Retrieve indicates an expected call of Retrieve.
func (mr *MockProjectServiceMockRecorder) Retrieve(ctx, projectID any) *MockProjectServiceRetrieveCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Retrieve", reflect.TypeOf((*MockProjectService)(nil).Retrieve), ctx, projectID)
	return &MockProjectServiceRetrieveCall{Call: call}
}

// MockProjectServiceRetrieveCall wrap *gomock.Call
type MockProjectServiceRetrieveCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectServiceRetrieveCall) Return(arg0 *Project, arg1 error) *MockProjectServiceRetrieveCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectServiceRetrieveCall) Do(f func(context.Context, string) (*Project, error)) *MockProjectServiceRetrieveCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectServiceRetrieveCall) DoAndReturn(f func(context.Context, string) (*Project, error)) *MockProjectServiceRetrieveCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
