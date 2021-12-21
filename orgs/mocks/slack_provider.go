// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SlackProvider is an autogenerated mock type for the SlackProvider type
type SlackProvider struct {
	mock.Mock
}

// SendText provides a mock function with given fields: text
func (_m *SlackProvider) SendText(text string) error {
	ret := _m.Called(text)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(text)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
