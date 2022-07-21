// Copyright 2021 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/kubefay/kubefay/pkg/agent/interfacestore (interfaces: InterfaceStore)

// Package testing is a generated GoMock package.
package testing

import (
	gomock "github.com/golang/mock/gomock"
	interfacestore "github.com/kubefay/kubefay/pkg/agent/interfacestore"
	reflect "reflect"
)

// MockInterfaceStore is a mock of InterfaceStore interface
type MockInterfaceStore struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceStoreMockRecorder
}

// MockInterfaceStoreMockRecorder is the mock recorder for MockInterfaceStore
type MockInterfaceStoreMockRecorder struct {
	mock *MockInterfaceStore
}

// NewMockInterfaceStore creates a new mock instance
func NewMockInterfaceStore(ctrl *gomock.Controller) *MockInterfaceStore {
	mock := &MockInterfaceStore{ctrl: ctrl}
	mock.recorder = &MockInterfaceStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterfaceStore) EXPECT() *MockInterfaceStoreMockRecorder {
	return m.recorder
}

// AddInterface mocks base method
func (m *MockInterfaceStore) AddInterface(arg0 *interfacestore.InterfaceConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddInterface", arg0)
}

// AddInterface indicates an expected call of AddInterface
func (mr *MockInterfaceStoreMockRecorder) AddInterface(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddInterface", reflect.TypeOf((*MockInterfaceStore)(nil).AddInterface), arg0)
}

// DeleteInterface mocks base method
func (m *MockInterfaceStore) DeleteInterface(arg0 *interfacestore.InterfaceConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteInterface", arg0)
}

// DeleteInterface indicates an expected call of DeleteInterface
func (mr *MockInterfaceStoreMockRecorder) DeleteInterface(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInterface", reflect.TypeOf((*MockInterfaceStore)(nil).DeleteInterface), arg0)
}

// GetContainerInterface mocks base method
func (m *MockInterfaceStore) GetContainerInterface(arg0 string) (*interfacestore.InterfaceConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainerInterface", arg0)
	ret0, _ := ret[0].(*interfacestore.InterfaceConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetContainerInterface indicates an expected call of GetContainerInterface
func (mr *MockInterfaceStoreMockRecorder) GetContainerInterface(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainerInterface", reflect.TypeOf((*MockInterfaceStore)(nil).GetContainerInterface), arg0)
}

// GetContainerInterfaceNum mocks base method
func (m *MockInterfaceStore) GetContainerInterfaceNum() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainerInterfaceNum")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetContainerInterfaceNum indicates an expected call of GetContainerInterfaceNum
func (mr *MockInterfaceStoreMockRecorder) GetContainerInterfaceNum() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainerInterfaceNum", reflect.TypeOf((*MockInterfaceStore)(nil).GetContainerInterfaceNum))
}

// GetContainerInterfacesByPod mocks base method
func (m *MockInterfaceStore) GetContainerInterfacesByPod(arg0, arg1 string) []*interfacestore.InterfaceConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContainerInterfacesByPod", arg0, arg1)
	ret0, _ := ret[0].([]*interfacestore.InterfaceConfig)
	return ret0
}

// GetContainerInterfacesByPod indicates an expected call of GetContainerInterfacesByPod
func (mr *MockInterfaceStoreMockRecorder) GetContainerInterfacesByPod(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContainerInterfacesByPod", reflect.TypeOf((*MockInterfaceStore)(nil).GetContainerInterfacesByPod), arg0, arg1)
}

// GetInterface mocks base method
func (m *MockInterfaceStore) GetInterface(arg0 string) (*interfacestore.InterfaceConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterface", arg0)
	ret0, _ := ret[0].(*interfacestore.InterfaceConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetInterface indicates an expected call of GetInterface
func (mr *MockInterfaceStoreMockRecorder) GetInterface(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterface", reflect.TypeOf((*MockInterfaceStore)(nil).GetInterface), arg0)
}

// GetInterfaceByIP mocks base method
func (m *MockInterfaceStore) GetInterfaceByIP(arg0 string) (*interfacestore.InterfaceConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterfaceByIP", arg0)
	ret0, _ := ret[0].(*interfacestore.InterfaceConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetInterfaceByIP indicates an expected call of GetInterfaceByIP
func (mr *MockInterfaceStoreMockRecorder) GetInterfaceByIP(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterfaceByIP", reflect.TypeOf((*MockInterfaceStore)(nil).GetInterfaceByIP), arg0)
}

// GetInterfaceByName mocks base method
func (m *MockInterfaceStore) GetInterfaceByName(arg0 string) (*interfacestore.InterfaceConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterfaceByName", arg0)
	ret0, _ := ret[0].(*interfacestore.InterfaceConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetInterfaceByName indicates an expected call of GetInterfaceByName
func (mr *MockInterfaceStoreMockRecorder) GetInterfaceByName(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterfaceByName", reflect.TypeOf((*MockInterfaceStore)(nil).GetInterfaceByName), arg0)
}

// GetInterfaceKeysByType mocks base method
func (m *MockInterfaceStore) GetInterfaceKeysByType(arg0 interfacestore.InterfaceType) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterfaceKeysByType", arg0)
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetInterfaceKeysByType indicates an expected call of GetInterfaceKeysByType
func (mr *MockInterfaceStoreMockRecorder) GetInterfaceKeysByType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterfaceKeysByType", reflect.TypeOf((*MockInterfaceStore)(nil).GetInterfaceKeysByType), arg0)
}

// GetInterfacesByType mocks base method
func (m *MockInterfaceStore) GetInterfacesByType(arg0 interfacestore.InterfaceType) []*interfacestore.InterfaceConfig {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInterfacesByType", arg0)
	ret0, _ := ret[0].([]*interfacestore.InterfaceConfig)
	return ret0
}

// GetInterfacesByType indicates an expected call of GetInterfacesByType
func (mr *MockInterfaceStoreMockRecorder) GetInterfacesByType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInterfacesByType", reflect.TypeOf((*MockInterfaceStore)(nil).GetInterfacesByType), arg0)
}

// GetNodeTunnelInterface mocks base method
func (m *MockInterfaceStore) GetNodeTunnelInterface(arg0 string) (*interfacestore.InterfaceConfig, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeTunnelInterface", arg0)
	ret0, _ := ret[0].(*interfacestore.InterfaceConfig)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetNodeTunnelInterface indicates an expected call of GetNodeTunnelInterface
func (mr *MockInterfaceStoreMockRecorder) GetNodeTunnelInterface(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeTunnelInterface", reflect.TypeOf((*MockInterfaceStore)(nil).GetNodeTunnelInterface), arg0)
}

// Initialize mocks base method
func (m *MockInterfaceStore) Initialize(arg0 []*interfacestore.InterfaceConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Initialize", arg0)
}

// Initialize indicates an expected call of Initialize
func (mr *MockInterfaceStoreMockRecorder) Initialize(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialize", reflect.TypeOf((*MockInterfaceStore)(nil).Initialize), arg0)
}

// Len mocks base method
func (m *MockInterfaceStore) Len() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Len")
	ret0, _ := ret[0].(int)
	return ret0
}

// Len indicates an expected call of Len
func (mr *MockInterfaceStoreMockRecorder) Len() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Len", reflect.TypeOf((*MockInterfaceStore)(nil).Len))
}
