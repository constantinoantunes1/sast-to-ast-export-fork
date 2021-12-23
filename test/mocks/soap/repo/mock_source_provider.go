// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/checkmarxDev/ast-sast-export/internal/soap/repo (interfaces: SourceProvider)

// Package mock_soap_repo is a generated GoMock package.
package mock_soap_repo

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSourceProvider is a mock of SourceProvider interface.
type MockSourceProvider struct {
	ctrl     *gomock.Controller
	recorder *MockSourceProviderMockRecorder
}

// MockSourceProviderMockRecorder is the mock recorder for MockSourceProvider.
type MockSourceProviderMockRecorder struct {
	mock *MockSourceProvider
}

// NewMockSourceProvider creates a new mock instance.
func NewMockSourceProvider(ctrl *gomock.Controller) *MockSourceProvider {
	mock := &MockSourceProvider{ctrl: ctrl}
	mock.recorder = &MockSourceProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSourceProvider) EXPECT() *MockSourceProviderMockRecorder {
	return m.recorder
}

// DownloadSourceFiles mocks base method.
func (m *MockSourceProvider) DownloadSourceFiles(arg0 string, arg1 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadSourceFiles", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadSourceFiles indicates an expected call of DownloadSourceFiles.
func (mr *MockSourceProviderMockRecorder) DownloadSourceFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadSourceFiles", reflect.TypeOf((*MockSourceProvider)(nil).DownloadSourceFiles), arg0, arg1)
}
