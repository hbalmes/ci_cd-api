// Code generated by MockGen. DO NOT EDIT.
// Source: services/configuration.go

// Package interfaces is a generated GoMock package.
package interfaces

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/hbalmes/ci_cd-api/api/models"
	reflect "reflect"
)

// MockConfigurationService is a mock of ConfigurationService interface
type MockConfigurationService struct {
	ctrl     *gomock.Controller
	recorder *MockConfigurationServiceMockRecorder
}

// MockConfigurationServiceMockRecorder is the mock recorder for MockConfigurationService
type MockConfigurationServiceMockRecorder struct {
	mock *MockConfigurationService
}

// NewMockConfigurationService creates a new mock instance
func NewMockConfigurationService(ctrl *gomock.Controller) *MockConfigurationService {
	mock := &MockConfigurationService{ctrl: ctrl}
	mock.recorder = &MockConfigurationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigurationService) EXPECT() *MockConfigurationServiceMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockConfigurationService) Create(arg0 *models.PostRequestPayload) (*models.Configuration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(*models.Configuration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockConfigurationServiceMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockConfigurationService)(nil).Create), arg0)
}

// Get mocks base method
func (m *MockConfigurationService) Get(arg0 string) (*models.Configuration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*models.Configuration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockConfigurationServiceMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockConfigurationService)(nil).Get), arg0)
}

// Update mocks base method
func (m *MockConfigurationService) Update(r *models.PutRequestPayload) (*models.Configuration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", r)
	ret0, _ := ret[0].(*models.Configuration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update
func (mr *MockConfigurationServiceMockRecorder) Update(r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockConfigurationService)(nil).Update), r)
}

// Delete mocks base method
func (m *MockConfigurationService) Delete(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockConfigurationServiceMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockConfigurationService)(nil).Delete), id)
}
