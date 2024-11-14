// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/spacelift-io/spacelift-operator/internal/spacelift/models"
	mock "github.com/stretchr/testify/mock"

	v1beta1 "github.com/spacelift-io/spacelift-operator/api/v1beta1"
)

// PolicyRepository is an autogenerated mock type for the PolicyRepository type
type PolicyRepository struct {
	mock.Mock
}

type PolicyRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *PolicyRepository) EXPECT() *PolicyRepository_Expecter {
	return &PolicyRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: _a0, _a1
func (_m *PolicyRepository) Create(_a0 context.Context, _a1 *v1beta1.Policy) (*models.Policy, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *models.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) (*models.Policy, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) *models.Policy); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Policy) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type PolicyRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Policy
func (_e *PolicyRepository_Expecter) Create(_a0 interface{}, _a1 interface{}) *PolicyRepository_Create_Call {
	return &PolicyRepository_Create_Call{Call: _e.mock.On("Create", _a0, _a1)}
}

func (_c *PolicyRepository_Create_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Policy)) *PolicyRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Policy))
	})
	return _c
}

func (_c *PolicyRepository_Create_Call) Return(_a0 *models.Policy, _a1 error) *PolicyRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyRepository_Create_Call) RunAndReturn(run func(context.Context, *v1beta1.Policy) (*models.Policy, error)) *PolicyRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *PolicyRepository) Get(_a0 context.Context, _a1 *v1beta1.Policy) (*models.Policy, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *models.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) (*models.Policy, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) *models.Policy); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Policy) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyRepository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type PolicyRepository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Policy
func (_e *PolicyRepository_Expecter) Get(_a0 interface{}, _a1 interface{}) *PolicyRepository_Get_Call {
	return &PolicyRepository_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *PolicyRepository_Get_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Policy)) *PolicyRepository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Policy))
	})
	return _c
}

func (_c *PolicyRepository_Get_Call) Return(_a0 *models.Policy, _a1 error) *PolicyRepository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyRepository_Get_Call) RunAndReturn(run func(context.Context, *v1beta1.Policy) (*models.Policy, error)) *PolicyRepository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0, _a1
func (_m *PolicyRepository) Update(_a0 context.Context, _a1 *v1beta1.Policy) (*models.Policy, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *models.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) (*models.Policy, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1beta1.Policy) *models.Policy); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1beta1.Policy) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyRepository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type PolicyRepository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v1beta1.Policy
func (_e *PolicyRepository_Expecter) Update(_a0 interface{}, _a1 interface{}) *PolicyRepository_Update_Call {
	return &PolicyRepository_Update_Call{Call: _e.mock.On("Update", _a0, _a1)}
}

func (_c *PolicyRepository_Update_Call) Run(run func(_a0 context.Context, _a1 *v1beta1.Policy)) *PolicyRepository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1beta1.Policy))
	})
	return _c
}

func (_c *PolicyRepository_Update_Call) Return(_a0 *models.Policy, _a1 error) *PolicyRepository_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyRepository_Update_Call) RunAndReturn(run func(context.Context, *v1beta1.Policy) (*models.Policy, error)) *PolicyRepository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewPolicyRepository creates a new instance of PolicyRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPolicyRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *PolicyRepository {
	mock := &PolicyRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
