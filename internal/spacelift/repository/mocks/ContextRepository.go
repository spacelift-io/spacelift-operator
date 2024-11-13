// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	mock "github.com/stretchr/testify/mock"

	v1beta1 "github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

// ContextRepository is an autogenerated mock type for the ContextRepository type
type ContextRepository struct {
	mock.Mock
}

type ContextRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *ContextRepository) EXPECT() *ContextRepository_Expecter {
	return &ContextRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1
func (_m *ContextRepository) Create(_a0 context.Context, _a1 *v1beta1.Context) (*models.Context, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *models.Context
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) (*models.Context, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) *models.Context); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Context)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Context) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContextRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ContextRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Context
func (_e *ContextRepository_Expecter) Create(_a0 interface{}, _a1 interface{}) *ContextRepository_Create_Call {
	return &ContextRepository_Create_Call{Call: _e.mock.On("Create", _a0, _a1)}
}

func (_c *ContextRepository_Create_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Context)) *ContextRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Context))
	})
	return _c
}

func (_c *ContextRepository_Create_Call) Return(_a0 *models.Context, _a1 error) *ContextRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContextRepository_Create_Call) RunAndReturn(run func(context.Context, *v1beta1.Context) (*models.Context, error)) *ContextRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *ContextRepository) Get(_a0 context.Context, _a1 *v1beta1.Context) (*models.Context, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Context
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) (*models.Context, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) *models.Context); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Context)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Context) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContextRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ContextRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Context
func (_e *ContextRepository_Expecter) Get(_a0 interface{}, _a1 interface{}) *ContextRepository_Get_Call {
	return &ContextRepository_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *ContextRepository_Get_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Context)) *ContextRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Context))
	})
	return _c
}

func (_c *ContextRepository_Get_Call) Return(_a0 *models.Context, _a1 error) *ContextRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContextRepository_Get_Call) RunAndReturn(run func(context.Context, *v1beta1.Context) (*models.Context, error)) *ContextRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0, _a1
func (_m *ContextRepository) Update(_a0 context.Context, _a1 *v1beta1.Context) (*models.Context, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *models.Context
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) (*models.Context, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Context) *models.Context); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Context)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Context) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContextRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ContextRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Context
func (_e *ContextRepository_Expecter) Update(_a0 interface{}, _a1 interface{}) *ContextRepository_Update_Call {
	return &ContextRepository_Update_Call{Call: _e.mock.On("Update", _a0, _a1)}
}

func (_c *ContextRepository_Update_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Context)) *ContextRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Context))
	})
	return _c
}

func (_c *ContextRepository_Update_Call) Return(_a0 *models.Context, _a1 error) *ContextRepository_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ContextRepository_Update_Call) RunAndReturn(run func(context.Context, *v1beta1.Context) (*models.Context, error)) *ContextRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewContextRepository creates a new instance of ContextRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewContextRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ContextRepository {
	mock := &ContextRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
