// Code generated by MockGen. DO NOT EDIT.
// Source: slk/slk.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSlkClient is a mock of SlkClient interface.
type MockSlkClient struct {
	ctrl     *gomock.Controller
	recorder *MockSlkClientMockRecorder
}

// MockSlkClientMockRecorder is the mock recorder for MockSlkClient.
type MockSlkClientMockRecorder struct {
	mock *MockSlkClient
}

// NewMockSlkClient creates a new mock instance.
func NewMockSlkClient(ctrl *gomock.Controller) *MockSlkClient {
	mock := &MockSlkClient{ctrl: ctrl}
	mock.recorder = &MockSlkClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSlkClient) EXPECT() *MockSlkClientMockRecorder {
	return m.recorder
}

// PostMessage mocks base method.
func (m_2 *MockSlkClient) PostMessage(ctx context.Context, m string, mentions []string) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "PostMessage", ctx, m, mentions)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostMessage indicates an expected call of PostMessage.
func (mr *MockSlkClientMockRecorder) PostMessage(ctx, m, mentions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostMessage", reflect.TypeOf((*MockSlkClient)(nil).PostMessage), ctx, m, mentions)
}