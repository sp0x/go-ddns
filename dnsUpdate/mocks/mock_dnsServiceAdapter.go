// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/sp0x/go-ddns/dnsUpdate (interfaces: DnsServiceAdapter)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	dns "google.golang.org/api/dns/v1"
	reflect "reflect"
)

// MockDnsServiceAdapter is a mock of DnsServiceAdapter interface
type MockDnsServiceAdapter struct {
	ctrl     *gomock.Controller
	recorder *MockDnsServiceAdapterMockRecorder
}

// MockDnsServiceAdapterMockRecorder is the mock recorder for MockDnsServiceAdapter
type MockDnsServiceAdapterMockRecorder struct {
	mock *MockDnsServiceAdapter
}

// NewMockDnsServiceAdapter creates a new mock instance
func NewMockDnsServiceAdapter(ctrl *gomock.Controller) *MockDnsServiceAdapter {
	mock := &MockDnsServiceAdapter{ctrl: ctrl}
	mock.recorder = &MockDnsServiceAdapterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDnsServiceAdapter) EXPECT() *MockDnsServiceAdapterMockRecorder {
	return m.recorder
}

// Change mocks base method
func (m *MockDnsServiceAdapter) Change(arg0 string, arg1 *dns.Change) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Change", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Change indicates an expected call of Change
func (mr *MockDnsServiceAdapterMockRecorder) Change(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Change", reflect.TypeOf((*MockDnsServiceAdapter)(nil).Change), arg0, arg1)
}

// List mocks base method
func (m *MockDnsServiceAdapter) List(arg0, arg1, arg2 string) ([]*dns.ResourceRecordSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*dns.ResourceRecordSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockDnsServiceAdapterMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockDnsServiceAdapter)(nil).List), arg0, arg1, arg2)
}