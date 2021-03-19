// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pingcap/kvproto/pkg/import_kvpb (interfaces: ImportKVClient,ImportKV_WriteEngineClient)

// $ mockgen -package mock github.com/pingcap/kvproto/pkg/import_kvpb ImportKVClient,ImportKV_WriteEngineClient

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	import_kvpb "github.com/pingcap/kvproto/pkg/import_kvpb"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

// MockImportKVClient is a mock of ImportKVClient interface
type MockImportKVClient struct {
	ctrl     *gomock.Controller
	recorder *MockImportKVClientMockRecorder
}

// MockImportKVClientMockRecorder is the mock recorder for MockImportKVClient
type MockImportKVClientMockRecorder struct {
	mock *MockImportKVClient
}

// NewMockImportKVClient creates a new mock instance
func NewMockImportKVClient(ctrl *gomock.Controller) *MockImportKVClient {
	mock := &MockImportKVClient{ctrl: ctrl}
	mock.recorder = &MockImportKVClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockImportKVClient) EXPECT() *MockImportKVClientMockRecorder {
	return m.recorder
}

// CleanupEngine mocks base method
func (m *MockImportKVClient) CleanupEngine(arg0 context.Context, arg1 *import_kvpb.CleanupEngineRequest, arg2 ...grpc.CallOption) (*import_kvpb.CleanupEngineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CleanupEngine", varargs...)
	ret0, _ := ret[0].(*import_kvpb.CleanupEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CleanupEngine indicates an expected call of CleanupEngine
func (mr *MockImportKVClientMockRecorder) CleanupEngine(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupEngine", reflect.TypeOf((*MockImportKVClient)(nil).CleanupEngine), varargs...)
}

// CloseEngine mocks base method
func (m *MockImportKVClient) CloseEngine(arg0 context.Context, arg1 *import_kvpb.CloseEngineRequest, arg2 ...grpc.CallOption) (*import_kvpb.CloseEngineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CloseEngine", varargs...)
	ret0, _ := ret[0].(*import_kvpb.CloseEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloseEngine indicates an expected call of CloseEngine
func (mr *MockImportKVClientMockRecorder) CloseEngine(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseEngine", reflect.TypeOf((*MockImportKVClient)(nil).CloseEngine), varargs...)
}

// CompactCluster mocks base method
func (m *MockImportKVClient) CompactCluster(arg0 context.Context, arg1 *import_kvpb.CompactClusterRequest, arg2 ...grpc.CallOption) (*import_kvpb.CompactClusterResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CompactCluster", varargs...)
	ret0, _ := ret[0].(*import_kvpb.CompactClusterResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompactCluster indicates an expected call of CompactCluster
func (mr *MockImportKVClientMockRecorder) CompactCluster(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompactCluster", reflect.TypeOf((*MockImportKVClient)(nil).CompactCluster), varargs...)
}

// GetMetrics mocks base method
func (m *MockImportKVClient) GetMetrics(arg0 context.Context, arg1 *import_kvpb.GetMetricsRequest, arg2 ...grpc.CallOption) (*import_kvpb.GetMetricsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMetrics", varargs...)
	ret0, _ := ret[0].(*import_kvpb.GetMetricsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetrics indicates an expected call of GetMetrics
func (mr *MockImportKVClientMockRecorder) GetMetrics(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetrics", reflect.TypeOf((*MockImportKVClient)(nil).GetMetrics), varargs...)
}

// GetVersion mocks base method
func (m *MockImportKVClient) GetVersion(arg0 context.Context, arg1 *import_kvpb.GetVersionRequest, arg2 ...grpc.CallOption) (*import_kvpb.GetVersionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetVersion", varargs...)
	ret0, _ := ret[0].(*import_kvpb.GetVersionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersion indicates an expected call of GetVersion
func (mr *MockImportKVClientMockRecorder) GetVersion(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersion", reflect.TypeOf((*MockImportKVClient)(nil).GetVersion), varargs...)
}

// ImportEngine mocks base method
func (m *MockImportKVClient) ImportEngine(arg0 context.Context, arg1 *import_kvpb.ImportEngineRequest, arg2 ...grpc.CallOption) (*import_kvpb.ImportEngineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ImportEngine", varargs...)
	ret0, _ := ret[0].(*import_kvpb.ImportEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ImportEngine indicates an expected call of ImportEngine
func (mr *MockImportKVClientMockRecorder) ImportEngine(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImportEngine", reflect.TypeOf((*MockImportKVClient)(nil).ImportEngine), varargs...)
}

// OpenEngine mocks base method
func (m *MockImportKVClient) OpenEngine(arg0 context.Context, arg1 *import_kvpb.OpenEngineRequest, arg2 ...grpc.CallOption) (*import_kvpb.OpenEngineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "OpenEngine", varargs...)
	ret0, _ := ret[0].(*import_kvpb.OpenEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenEngine indicates an expected call of OpenEngine
func (mr *MockImportKVClientMockRecorder) OpenEngine(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenEngine", reflect.TypeOf((*MockImportKVClient)(nil).OpenEngine), varargs...)
}

// SwitchMode mocks base method
func (m *MockImportKVClient) SwitchMode(arg0 context.Context, arg1 *import_kvpb.SwitchModeRequest, arg2 ...grpc.CallOption) (*import_kvpb.SwitchModeResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SwitchMode", varargs...)
	ret0, _ := ret[0].(*import_kvpb.SwitchModeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SwitchMode indicates an expected call of SwitchMode
func (mr *MockImportKVClientMockRecorder) SwitchMode(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SwitchMode", reflect.TypeOf((*MockImportKVClient)(nil).SwitchMode), varargs...)
}

// WriteEngine mocks base method
func (m *MockImportKVClient) WriteEngine(arg0 context.Context, arg1 ...grpc.CallOption) (import_kvpb.ImportKV_WriteEngineClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "WriteEngine", varargs...)
	ret0, _ := ret[0].(import_kvpb.ImportKV_WriteEngineClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WriteEngine indicates an expected call of WriteEngine
func (mr *MockImportKVClientMockRecorder) WriteEngine(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteEngine", reflect.TypeOf((*MockImportKVClient)(nil).WriteEngine), varargs...)
}

// WriteEngineV3 mocks base method
func (m *MockImportKVClient) WriteEngineV3(arg0 context.Context, arg1 *import_kvpb.WriteEngineV3Request, arg2 ...grpc.CallOption) (*import_kvpb.WriteEngineResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "WriteEngineV3", varargs...)
	ret0, _ := ret[0].(*import_kvpb.WriteEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WriteEngineV3 indicates an expected call of WriteEngineV3
func (mr *MockImportKVClientMockRecorder) WriteEngineV3(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteEngineV3", reflect.TypeOf((*MockImportKVClient)(nil).WriteEngineV3), varargs...)
}

// MockImportKV_WriteEngineClient is a mock of ImportKV_WriteEngineClient interface
type MockImportKV_WriteEngineClient struct {
	ctrl     *gomock.Controller
	recorder *MockImportKV_WriteEngineClientMockRecorder
}

// MockImportKV_WriteEngineClientMockRecorder is the mock recorder for MockImportKV_WriteEngineClient
type MockImportKV_WriteEngineClientMockRecorder struct {
	mock *MockImportKV_WriteEngineClient
}

// NewMockImportKV_WriteEngineClient creates a new mock instance
func NewMockImportKV_WriteEngineClient(ctrl *gomock.Controller) *MockImportKV_WriteEngineClient {
	mock := &MockImportKV_WriteEngineClient{ctrl: ctrl}
	mock.recorder = &MockImportKV_WriteEngineClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockImportKV_WriteEngineClient) EXPECT() *MockImportKV_WriteEngineClientMockRecorder {
	return m.recorder
}

// CloseAndRecv mocks base method
func (m *MockImportKV_WriteEngineClient) CloseAndRecv() (*import_kvpb.WriteEngineResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseAndRecv")
	ret0, _ := ret[0].(*import_kvpb.WriteEngineResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloseAndRecv indicates an expected call of CloseAndRecv
func (mr *MockImportKV_WriteEngineClientMockRecorder) CloseAndRecv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseAndRecv", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).CloseAndRecv))
}

// CloseSend mocks base method
func (m *MockImportKV_WriteEngineClient) CloseSend() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseSend")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseSend indicates an expected call of CloseSend
func (mr *MockImportKV_WriteEngineClientMockRecorder) CloseSend() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseSend", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).CloseSend))
}

// Context mocks base method
func (m *MockImportKV_WriteEngineClient) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockImportKV_WriteEngineClientMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).Context))
}

// Header mocks base method
func (m *MockImportKV_WriteEngineClient) Header() (metadata.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Header indicates an expected call of Header
func (mr *MockImportKV_WriteEngineClientMockRecorder) Header() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).Header))
}

// RecvMsg mocks base method
func (m *MockImportKV_WriteEngineClient) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockImportKV_WriteEngineClientMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockImportKV_WriteEngineClient) Send(arg0 *import_kvpb.WriteEngineRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockImportKV_WriteEngineClientMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).Send), arg0)
}

// SendMsg mocks base method
func (m *MockImportKV_WriteEngineClient) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockImportKV_WriteEngineClientMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).SendMsg), arg0)
}

// Trailer mocks base method
func (m *MockImportKV_WriteEngineClient) Trailer() metadata.MD {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Trailer")
	ret0, _ := ret[0].(metadata.MD)
	return ret0
}

// Trailer indicates an expected call of Trailer
func (mr *MockImportKV_WriteEngineClientMockRecorder) Trailer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trailer", reflect.TypeOf((*MockImportKV_WriteEngineClient)(nil).Trailer))
}
