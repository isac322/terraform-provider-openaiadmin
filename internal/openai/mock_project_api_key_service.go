// Code generated by MockGen. DO NOT EDIT.
// Source: project_api_key_service.go
//
// Generated by this command:
//
//	mockgen -package openai -destination mock_project_api_key_service.go -source project_api_key_service.go -typed
//

// Package openai is a generated GoMock package.
package openai

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockProjectAPIKeyService is a mock of ProjectAPIKeyService interface.
type MockProjectAPIKeyService struct {
	ctrl     *gomock.Controller
	recorder *MockProjectAPIKeyServiceMockRecorder
	isgomock struct{}
}

// MockProjectAPIKeyServiceMockRecorder is the mock recorder for MockProjectAPIKeyService.
type MockProjectAPIKeyServiceMockRecorder struct {
	mock *MockProjectAPIKeyService
}

// NewMockProjectAPIKeyService creates a new mock instance.
func NewMockProjectAPIKeyService(ctrl *gomock.Controller) *MockProjectAPIKeyService {
	mock := &MockProjectAPIKeyService{ctrl: ctrl}
	mock.recorder = &MockProjectAPIKeyServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProjectAPIKeyService) EXPECT() *MockProjectAPIKeyServiceMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockProjectAPIKeyService) Delete(ctx context.Context, projectID, apiKeyID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, projectID, apiKeyID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockProjectAPIKeyServiceMockRecorder) Delete(ctx, projectID, apiKeyID any) *MockProjectAPIKeyServiceDeleteCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockProjectAPIKeyService)(nil).Delete), ctx, projectID, apiKeyID)
	return &MockProjectAPIKeyServiceDeleteCall{Call: call}
}

// MockProjectAPIKeyServiceDeleteCall wrap *gomock.Call
type MockProjectAPIKeyServiceDeleteCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectAPIKeyServiceDeleteCall) Return(arg0 error) *MockProjectAPIKeyServiceDeleteCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectAPIKeyServiceDeleteCall) Do(f func(context.Context, string, string) error) *MockProjectAPIKeyServiceDeleteCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectAPIKeyServiceDeleteCall) DoAndReturn(f func(context.Context, string, string) error) *MockProjectAPIKeyServiceDeleteCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// List mocks base method.
func (m *MockProjectAPIKeyService) List(ctx context.Context, projectID string) ([]ProjectAPIKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, projectID)
	ret0, _ := ret[0].([]ProjectAPIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockProjectAPIKeyServiceMockRecorder) List(ctx, projectID any) *MockProjectAPIKeyServiceListCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockProjectAPIKeyService)(nil).List), ctx, projectID)
	return &MockProjectAPIKeyServiceListCall{Call: call}
}

// MockProjectAPIKeyServiceListCall wrap *gomock.Call
type MockProjectAPIKeyServiceListCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectAPIKeyServiceListCall) Return(arg0 []ProjectAPIKey, arg1 error) *MockProjectAPIKeyServiceListCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectAPIKeyServiceListCall) Do(f func(context.Context, string) ([]ProjectAPIKey, error)) *MockProjectAPIKeyServiceListCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectAPIKeyServiceListCall) DoAndReturn(f func(context.Context, string) ([]ProjectAPIKey, error)) *MockProjectAPIKeyServiceListCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Retrieve mocks base method.
func (m *MockProjectAPIKeyService) Retrieve(ctx context.Context, projectID, apiKeyID string) (*ProjectAPIKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Retrieve", ctx, projectID, apiKeyID)
	ret0, _ := ret[0].(*ProjectAPIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Retrieve indicates an expected call of Retrieve.
func (mr *MockProjectAPIKeyServiceMockRecorder) Retrieve(ctx, projectID, apiKeyID any) *MockProjectAPIKeyServiceRetrieveCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Retrieve", reflect.TypeOf((*MockProjectAPIKeyService)(nil).Retrieve), ctx, projectID, apiKeyID)
	return &MockProjectAPIKeyServiceRetrieveCall{Call: call}
}

// MockProjectAPIKeyServiceRetrieveCall wrap *gomock.Call
type MockProjectAPIKeyServiceRetrieveCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockProjectAPIKeyServiceRetrieveCall) Return(arg0 *ProjectAPIKey, arg1 error) *MockProjectAPIKeyServiceRetrieveCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockProjectAPIKeyServiceRetrieveCall) Do(f func(context.Context, string, string) (*ProjectAPIKey, error)) *MockProjectAPIKeyServiceRetrieveCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockProjectAPIKeyServiceRetrieveCall) DoAndReturn(f func(context.Context, string, string) (*ProjectAPIKey, error)) *MockProjectAPIKeyServiceRetrieveCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
