// Code generated by MockGen. DO NOT EDIT.
// Source: auth.go

// Package mock_service is a generated GoMock package.
package mock_storage

import (
	context "context"
	reflect "reflect"

	models "github.com/dkrasnykh/gophkeeper/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockUserProvider is a mock of UserProvider interface.
type MockUserProvider struct {
	ctrl     *gomock.Controller
	recorder *MockUserProviderMockRecorder
}

// MockUserProviderMockRecorder is the mock recorder for MockUserProvider.
type MockUserProviderMockRecorder struct {
	mock *MockUserProvider
}

// NewMockUserProvider creates a new mock instance.
func NewMockUserProvider(ctrl *gomock.Controller) *MockUserProvider {
	mock := &MockUserProvider{ctrl: ctrl}
	mock.recorder = &MockUserProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserProvider) EXPECT() *MockUserProviderMockRecorder {
	return m.recorder
}

// SaveUser mocks base method.
func (m *MockUserProvider) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, email, passHash)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockUserProviderMockRecorder) SaveUser(ctx, email, passHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockUserProvider)(nil).SaveUser), ctx, email, passHash)
}

// User mocks base method.
func (m *MockUserProvider) User(ctx context.Context, email string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "User", ctx, email)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// User indicates an expected call of User.
func (mr *MockUserProviderMockRecorder) User(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "User", reflect.TypeOf((*MockUserProvider)(nil).User), ctx, email)
}

// MockAppProvider is a mock of AppProvider interface.
type MockAppProvider struct {
	ctrl     *gomock.Controller
	recorder *MockAppProviderMockRecorder
}

// MockAppProviderMockRecorder is the mock recorder for MockAppProvider.
type MockAppProviderMockRecorder struct {
	mock *MockAppProvider
}

// NewMockAppProvider creates a new mock instance.
func NewMockAppProvider(ctrl *gomock.Controller) *MockAppProvider {
	mock := &MockAppProvider{ctrl: ctrl}
	mock.recorder = &MockAppProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppProvider) EXPECT() *MockAppProviderMockRecorder {
	return m.recorder
}

// App mocks base method.
func (m *MockAppProvider) App(ctx context.Context, id int) (models.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "App", ctx, id)
	ret0, _ := ret[0].(models.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// App indicates an expected call of App.
func (mr *MockAppProviderMockRecorder) App(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "App", reflect.TypeOf((*MockAppProvider)(nil).App), ctx, id)
}
