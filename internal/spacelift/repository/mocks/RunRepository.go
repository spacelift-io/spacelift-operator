// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	context "context"

	v1beta1 "github.com/spacelift-io/spacelift-operator/api/v1beta1"
	repository "github.com/spacelift-io/spacelift-operator/internal/spacelift/repository"
	mock "github.com/stretchr/testify/mock"
)

// RunRepository is an autogenerated mock type for the RunRepository type
type RunRepository struct {
	mock.Mock
}

type RunRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *RunRepository) EXPECT() *RunRepository_Expecter {
	return &RunRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1
func (_m *RunRepository) Create(_a0 context.Context, _a1 *v1beta1.Run) (*repository.CreateRunOutput, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *repository.CreateRunOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Run) (*repository.CreateRunOutput, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Run) *repository.CreateRunOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.CreateRunOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Run) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RunRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type RunRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Run
func (_e *RunRepository_Expecter) Create(_a0 interface{}, _a1 interface{}) *RunRepository_Create_Call {
	return &RunRepository_Create_Call{Call: _e.mock.On("Create", _a0, _a1)}
}

func (_c *RunRepository_Create_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Run)) *RunRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Run))
	})
	return _c
}

func (_c *RunRepository_Create_Call) Return(_a0 *repository.CreateRunOutput, _a1 error) *RunRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *RunRepository_Create_Call) RunAndReturn(run func(context.Context, *v1beta1.Run) (*repository.CreateRunOutput, error)) *RunRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *RunRepository) Get(_a0 context.Context, _a1 *v1beta1.Run) (*repository.GetRunOutput, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *repository.GetRunOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Run) (*repository.GetRunOutput, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Run) *repository.GetRunOutput); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.GetRunOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Run) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RunRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type RunRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Run
func (_e *RunRepository_Expecter) Get(_a0 interface{}, _a1 interface{}) *RunRepository_Get_Call {
	return &RunRepository_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *RunRepository_Get_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Run)) *RunRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Run))
	})
	return _c
}

func (_c *RunRepository_Get_Call) Return(_a0 *repository.GetRunOutput, _a1 error) *RunRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *RunRepository_Get_Call) RunAndReturn(run func(context.Context, *v1beta1.Run) (*repository.GetRunOutput, error)) *RunRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewRunRepository creates a new instance of RunRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRunRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *RunRepository {
	mock := &RunRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}