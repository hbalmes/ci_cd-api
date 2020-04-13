// Code generated by MockGen. DO NOT EDIT.
// Source: services/webhook.go

// Package interfaces is a generated GoMock package.
package interfaces

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/hbalmes/ci_cd-api/api/models"
	webhook "github.com/hbalmes/ci_cd-api/api/models/webhook"
	apierrors "github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	reflect "reflect"
)

// MockWebhookService is a mock of WebhookService interface
type MockWebhookService struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookServiceMockRecorder
}

// MockWebhookServiceMockRecorder is the mock recorder for MockWebhookService
type MockWebhookServiceMockRecorder struct {
	mock *MockWebhookService
}

// NewMockWebhookService creates a new mock instance
func NewMockWebhookService(ctrl *gomock.Controller) *MockWebhookService {
	mock := &MockWebhookService{ctrl: ctrl}
	mock.recorder = &MockWebhookServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWebhookService) EXPECT() *MockWebhookServiceMockRecorder {
	return m.recorder
}

// ProcessStatusWebhook mocks base method
func (m *MockWebhookService) ProcessStatusWebhook(payload *webhook.Status, conf *models.Configuration) (*webhook.Webhook, apierrors.ApiError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessStatusWebhook", payload, conf)
	ret0, _ := ret[0].(*webhook.Webhook)
	ret1, _ := ret[1].(apierrors.ApiError)
	return ret0, ret1
}

// ProcessStatusWebhook indicates an expected call of ProcessStatusWebhook
func (mr *MockWebhookServiceMockRecorder) ProcessStatusWebhook(payload, conf interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessStatusWebhook", reflect.TypeOf((*MockWebhookService)(nil).ProcessStatusWebhook), payload, conf)
}

// ProcessPullRequestWebhook mocks base method
func (m *MockWebhookService) ProcessPullRequestWebhook(payload *webhook.PullRequestWebhook) (*webhook.Webhook, apierrors.ApiError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessPullRequestWebhook", payload)
	ret0, _ := ret[0].(*webhook.Webhook)
	ret1, _ := ret[1].(apierrors.ApiError)
	return ret0, ret1
}

// ProcessPullRequestWebhook indicates an expected call of ProcessPullRequestWebhook
func (mr *MockWebhookServiceMockRecorder) ProcessPullRequestWebhook(payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessPullRequestWebhook", reflect.TypeOf((*MockWebhookService)(nil).ProcessPullRequestWebhook), payload)
}

// ProcessPullRequestReviewWebhook mocks base method
func (m *MockWebhookService) ProcessPullRequestReviewWebhook(payload *webhook.PullRequestReviewWebhook) (*webhook.Webhook, apierrors.ApiError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessPullRequestReviewWebhook", payload)
	ret0, _ := ret[0].(*webhook.Webhook)
	ret1, _ := ret[1].(apierrors.ApiError)
	return ret0, ret1
}

// ProcessPullRequestReviewWebhook indicates an expected call of ProcessPullRequestReviewWebhook
func (mr *MockWebhookServiceMockRecorder) ProcessPullRequestReviewWebhook(payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessPullRequestReviewWebhook", reflect.TypeOf((*MockWebhookService)(nil).ProcessPullRequestReviewWebhook), payload)
}

// SavePullRequestWebhook mocks base method
func (m *MockWebhookService) SavePullRequestWebhook(pullRequestWH webhook.PullRequestWebhook) apierrors.ApiError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SavePullRequestWebhook", pullRequestWH)
	ret0, _ := ret[0].(apierrors.ApiError)
	return ret0
}

// SavePullRequestWebhook indicates an expected call of SavePullRequestWebhook
func (mr *MockWebhookServiceMockRecorder) SavePullRequestWebhook(pullRequestWH interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SavePullRequestWebhook", reflect.TypeOf((*MockWebhookService)(nil).SavePullRequestWebhook), pullRequestWH)
}
