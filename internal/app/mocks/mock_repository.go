// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/tank4gun/gourlshortener/internal/app/storage (interfaces: IRepository)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/tank4gun/gourlshortener/internal/app/storage"
)

// MockIRepository is a mock of IRepository interface.
type MockIRepository struct {
	ctrl     *gomock.Controller
	recorder *MockIRepositoryMockRecorder
}

// MockIRepositoryMockRecorder is the mock recorder for MockIRepository.
type MockIRepositoryMockRecorder struct {
	mock *MockIRepository
}

// NewMockIRepository creates a new mock instance.
func NewMockIRepository(ctrl *gomock.Controller) *MockIRepository {
	mock := &MockIRepository{ctrl: ctrl}
	mock.recorder = &MockIRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIRepository) EXPECT() *MockIRepositoryMockRecorder {
	return m.recorder
}

// CreateShortURLBatch mocks base method.
func (m *MockIRepository) CreateShortURLBatch(arg0 []storage.BatchURLRequest, arg1 uint, arg2 string) ([]storage.BatchURLResponse, string, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateShortURLBatch", arg0, arg1, arg2)
	ret0, _ := ret[0].([]storage.BatchURLResponse)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(int)
	return ret0, ret1, ret2
}

// CreateShortURLBatch indicates an expected call of CreateShortURLBatch.
func (mr *MockIRepositoryMockRecorder) CreateShortURLBatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateShortURLBatch", reflect.TypeOf((*MockIRepository)(nil).CreateShortURLBatch), arg0, arg1, arg2)
}

// CreateShortURLByURL mocks base method.
func (m *MockIRepository) CreateShortURLByURL(arg0 string, arg1 uint) (string, string, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateShortURLByURL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(int)
	return ret0, ret1, ret2
}

// CreateShortURLByURL indicates an expected call of CreateShortURLByURL.
func (mr *MockIRepositoryMockRecorder) CreateShortURLByURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateShortURLByURL", reflect.TypeOf((*MockIRepository)(nil).CreateShortURLByURL), arg0, arg1)
}

// GetAllURLsByUserID mocks base method.
func (m *MockIRepository) GetAllURLsByUserID(arg0 uint, arg1 string) ([]storage.FullInfoURLResponse, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllURLsByUserID", arg0, arg1)
	ret0, _ := ret[0].([]storage.FullInfoURLResponse)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetAllURLsByUserID indicates an expected call of GetAllURLsByUserID.
func (mr *MockIRepositoryMockRecorder) GetAllURLsByUserID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllURLsByUserID", reflect.TypeOf((*MockIRepository)(nil).GetAllURLsByUserID), arg0, arg1)
}

// GetNextIndex mocks base method.
func (m *MockIRepository) GetNextIndex() (uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNextIndex")
	ret0, _ := ret[0].(uint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNextIndex indicates an expected call of GetNextIndex.
func (mr *MockIRepositoryMockRecorder) GetNextIndex() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNextIndex", reflect.TypeOf((*MockIRepository)(nil).GetNextIndex))
}

// GetStats mocks base method.
func (m *MockIRepository) GetStats() (storage.StatsResponse, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStats")
	ret0, _ := ret[0].(storage.StatsResponse)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetStats indicates an expected call of GetStats.
func (mr *MockIRepositoryMockRecorder) GetStats() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStats", reflect.TypeOf((*MockIRepository)(nil).GetStats))
}

// GetValueByKeyAndUserID mocks base method.
func (m *MockIRepository) GetValueByKeyAndUserID(arg0, arg1 uint) (string, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValueByKeyAndUserID", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetValueByKeyAndUserID indicates an expected call of GetValueByKeyAndUserID.
func (mr *MockIRepositoryMockRecorder) GetValueByKeyAndUserID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValueByKeyAndUserID", reflect.TypeOf((*MockIRepository)(nil).GetValueByKeyAndUserID), arg0, arg1)
}

// InsertBatchValues mocks base method.
func (m *MockIRepository) InsertBatchValues(arg0 []string, arg1, arg2 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertBatchValues", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertBatchValues indicates an expected call of InsertBatchValues.
func (mr *MockIRepositoryMockRecorder) InsertBatchValues(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertBatchValues", reflect.TypeOf((*MockIRepository)(nil).InsertBatchValues), arg0, arg1, arg2)
}

// InsertValue mocks base method.
func (m *MockIRepository) InsertValue(arg0 string, arg1 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertValue indicates an expected call of InsertValue.
func (mr *MockIRepositoryMockRecorder) InsertValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertValue", reflect.TypeOf((*MockIRepository)(nil).InsertValue), arg0, arg1)
}

// MarkBatchAsDeleted mocks base method.
func (m *MockIRepository) MarkBatchAsDeleted(arg0 []uint, arg1 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkBatchAsDeleted", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkBatchAsDeleted indicates an expected call of MarkBatchAsDeleted.
func (mr *MockIRepositoryMockRecorder) MarkBatchAsDeleted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkBatchAsDeleted", reflect.TypeOf((*MockIRepository)(nil).MarkBatchAsDeleted), arg0, arg1)
}

// Ping mocks base method.
func (m *MockIRepository) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockIRepositoryMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockIRepository)(nil).Ping))
}

// Shutdown mocks base method.
func (m *MockIRepository) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockIRepositoryMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockIRepository)(nil).Shutdown))
}
