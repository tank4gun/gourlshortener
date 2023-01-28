// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/tank4gun/gourlshortener/internal/app/storage (interfaces: Repository)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	storage "github.com/tank4gun/gourlshortener/internal/app/storage"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetAllURLsByUserID mocks base method.
func (m *MockRepository) GetAllURLsByUserID(arg0 uint, arg1 string) ([]storage.FullInfoURLResponse, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllURLsByUserID", arg0, arg1)
	ret0, _ := ret[0].([]storage.FullInfoURLResponse)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetAllURLsByUserID indicates an expected call of GetAllURLsByUserID.
func (mr *MockRepositoryMockRecorder) GetAllURLsByUserID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllURLsByUserID", reflect.TypeOf((*MockRepository)(nil).GetAllURLsByUserID), arg0, arg1)
}

// GetNextIndex mocks base method.
func (m *MockRepository) GetNextIndex() (uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNextIndex")
	ret0, _ := ret[0].(uint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNextIndex indicates an expected call of GetNextIndex.
func (mr *MockRepositoryMockRecorder) GetNextIndex() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNextIndex", reflect.TypeOf((*MockRepository)(nil).GetNextIndex))
}

// GetValueByKeyAndUserID mocks base method.
func (m *MockRepository) GetValueByKeyAndUserID(arg0, arg1 uint) (string, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValueByKeyAndUserID", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// GetValueByKeyAndUserID indicates an expected call of GetValueByKeyAndUserID.
func (mr *MockRepositoryMockRecorder) GetValueByKeyAndUserID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValueByKeyAndUserID", reflect.TypeOf((*MockRepository)(nil).GetValueByKeyAndUserID), arg0, arg1)
}

// InsertBatchValues mocks base method.
func (m *MockRepository) InsertBatchValues(arg0 []string, arg1, arg2 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertBatchValues", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertBatchValues indicates an expected call of InsertBatchValues.
func (mr *MockRepositoryMockRecorder) InsertBatchValues(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertBatchValues", reflect.TypeOf((*MockRepository)(nil).InsertBatchValues), arg0, arg1, arg2)
}

// InsertValue mocks base method.
func (m *MockRepository) InsertValue(arg0 string, arg1 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertValue indicates an expected call of InsertValue.
func (mr *MockRepositoryMockRecorder) InsertValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertValue", reflect.TypeOf((*MockRepository)(nil).InsertValue), arg0, arg1)
}

// MarkBatchAsDeleted mocks base method.
func (m *MockRepository) MarkBatchAsDeleted(arg0 []uint, arg1 uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkBatchAsDeleted", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkBatchAsDeleted indicates an expected call of MarkBatchAsDeleted.
func (mr *MockRepositoryMockRecorder) MarkBatchAsDeleted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkBatchAsDeleted", reflect.TypeOf((*MockRepository)(nil).MarkBatchAsDeleted), arg0, arg1)
}

// Ping mocks base method.
func (m *MockRepository) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockRepositoryMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockRepository)(nil).Ping))
}
