// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/checkmarxDev/ast-sast-export/internal/sast/report (interfaces: SimilarityCalculator)

// Package mock_report is a generated GoMock package.
package mock_report

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSimilarityCalculator is a mock of SimilarityCalculator interface.
type MockSimilarityCalculator struct {
	ctrl     *gomock.Controller
	recorder *MockSimilarityCalculatorMockRecorder
}

// MockSimilarityCalculatorMockRecorder is the mock recorder for MockSimilarityCalculator.
type MockSimilarityCalculatorMockRecorder struct {
	mock *MockSimilarityCalculator
}

// NewMockSimilarityCalculator creates a new mock instance.
func NewMockSimilarityCalculator(ctrl *gomock.Controller) *MockSimilarityCalculator {
	mock := &MockSimilarityCalculator{ctrl: ctrl}
	mock.recorder = &MockSimilarityCalculatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSimilarityCalculator) EXPECT() *MockSimilarityCalculatorMockRecorder {
	return m.recorder
}

// Calculate mocks base method.
func (m *MockSimilarityCalculator) Calculate(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Calculate", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Calculate indicates an expected call of Calculate.
func (mr *MockSimilarityCalculatorMockRecorder) Calculate(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Calculate", reflect.TypeOf((*MockSimilarityCalculator)(nil).Calculate), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10)
}
