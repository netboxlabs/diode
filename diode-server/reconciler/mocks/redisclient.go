// Code generated by mockery v2.42.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	redis "github.com/redis/go-redis/v9"
)

// RedisClient is an autogenerated mock type for the RedisClient type
type RedisClient struct {
	mock.Mock
}

type RedisClient_Expecter struct {
	mock *mock.Mock
}

func (_m *RedisClient) EXPECT() *RedisClient_Expecter {
	return &RedisClient_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *RedisClient) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RedisClient_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type RedisClient_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *RedisClient_Expecter) Close() *RedisClient_Close_Call {
	return &RedisClient_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *RedisClient_Close_Call) Run(run func()) *RedisClient_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *RedisClient_Close_Call) Return(_a0 error) *RedisClient_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_Close_Call) RunAndReturn(run func() error) *RedisClient_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Del provides a mock function with given fields: ctx, keys
func (_m *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	_va := make([]interface{}, len(keys))
	for _i := range keys {
		_va[_i] = keys[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Del")
	}

	var r0 *redis.IntCmd
	if rf, ok := ret.Get(0).(func(context.Context, ...string) *redis.IntCmd); ok {
		r0 = rf(ctx, keys...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.IntCmd)
		}
	}

	return r0
}

// RedisClient_Del_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Del'
type RedisClient_Del_Call struct {
	*mock.Call
}

// Del is a helper method to define mock.On call
//   - ctx context.Context
//   - keys ...string
func (_e *RedisClient_Expecter) Del(ctx interface{}, keys ...interface{}) *RedisClient_Del_Call {
	return &RedisClient_Del_Call{Call: _e.mock.On("Del",
		append([]interface{}{ctx}, keys...)...)}
}

func (_c *RedisClient_Del_Call) Run(run func(ctx context.Context, keys ...string)) *RedisClient_Del_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *RedisClient_Del_Call) Return(_a0 *redis.IntCmd) *RedisClient_Del_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_Del_Call) RunAndReturn(run func(context.Context, ...string) *redis.IntCmd) *RedisClient_Del_Call {
	_c.Call.Return(run)
	return _c
}

// Do provides a mock function with given fields: ctx, args
func (_m *RedisClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Do")
	}

	var r0 *redis.Cmd
	if rf, ok := ret.Get(0).(func(context.Context, ...interface{}) *redis.Cmd); ok {
		r0 = rf(ctx, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.Cmd)
		}
	}

	return r0
}

// RedisClient_Do_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Do'
type RedisClient_Do_Call struct {
	*mock.Call
}

// Do is a helper method to define mock.On call
//   - ctx context.Context
//   - args ...interface{}
func (_e *RedisClient_Expecter) Do(ctx interface{}, args ...interface{}) *RedisClient_Do_Call {
	return &RedisClient_Do_Call{Call: _e.mock.On("Do",
		append([]interface{}{ctx}, args...)...)}
}

func (_c *RedisClient_Do_Call) Run(run func(ctx context.Context, args ...interface{})) *RedisClient_Do_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *RedisClient_Do_Call) Return(_a0 *redis.Cmd) *RedisClient_Do_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_Do_Call) RunAndReturn(run func(context.Context, ...interface{}) *redis.Cmd) *RedisClient_Do_Call {
	_c.Call.Return(run)
	return _c
}

// Ping provides a mock function with given fields: ctx
func (_m *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 *redis.StatusCmd
	if rf, ok := ret.Get(0).(func(context.Context) *redis.StatusCmd); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StatusCmd)
		}
	}

	return r0
}

// RedisClient_Ping_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ping'
type RedisClient_Ping_Call struct {
	*mock.Call
}

// Ping is a helper method to define mock.On call
//   - ctx context.Context
func (_e *RedisClient_Expecter) Ping(ctx interface{}) *RedisClient_Ping_Call {
	return &RedisClient_Ping_Call{Call: _e.mock.On("Ping", ctx)}
}

func (_c *RedisClient_Ping_Call) Run(run func(ctx context.Context)) *RedisClient_Ping_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *RedisClient_Ping_Call) Return(_a0 *redis.StatusCmd) *RedisClient_Ping_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_Ping_Call) RunAndReturn(run func(context.Context) *redis.StatusCmd) *RedisClient_Ping_Call {
	_c.Call.Return(run)
	return _c
}

// Scan provides a mock function with given fields: ctx, cursor, match, count
func (_m *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	ret := _m.Called(ctx, cursor, match, count)

	if len(ret) == 0 {
		panic("no return value specified for Scan")
	}

	var r0 *redis.ScanCmd
	if rf, ok := ret.Get(0).(func(context.Context, uint64, string, int64) *redis.ScanCmd); ok {
		r0 = rf(ctx, cursor, match, count)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.ScanCmd)
		}
	}

	return r0
}

// RedisClient_Scan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Scan'
type RedisClient_Scan_Call struct {
	*mock.Call
}

// Scan is a helper method to define mock.On call
//   - ctx context.Context
//   - cursor uint64
//   - match string
//   - count int64
func (_e *RedisClient_Expecter) Scan(ctx interface{}, cursor interface{}, match interface{}, count interface{}) *RedisClient_Scan_Call {
	return &RedisClient_Scan_Call{Call: _e.mock.On("Scan", ctx, cursor, match, count)}
}

func (_c *RedisClient_Scan_Call) Run(run func(ctx context.Context, cursor uint64, match string, count int64)) *RedisClient_Scan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(string), args[3].(int64))
	})
	return _c
}

func (_c *RedisClient_Scan_Call) Return(_a0 *redis.ScanCmd) *RedisClient_Scan_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_Scan_Call) RunAndReturn(run func(context.Context, uint64, string, int64) *redis.ScanCmd) *RedisClient_Scan_Call {
	_c.Call.Return(run)
	return _c
}

// XAck provides a mock function with given fields: ctx, stream, group, ids
func (_m *RedisClient) XAck(ctx context.Context, stream string, group string, ids ...string) *redis.IntCmd {
	_va := make([]interface{}, len(ids))
	for _i := range ids {
		_va[_i] = ids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, stream, group)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for XAck")
	}

	var r0 *redis.IntCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...string) *redis.IntCmd); ok {
		r0 = rf(ctx, stream, group, ids...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.IntCmd)
		}
	}

	return r0
}

// RedisClient_XAck_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'XAck'
type RedisClient_XAck_Call struct {
	*mock.Call
}

// XAck is a helper method to define mock.On call
//   - ctx context.Context
//   - stream string
//   - group string
//   - ids ...string
func (_e *RedisClient_Expecter) XAck(ctx interface{}, stream interface{}, group interface{}, ids ...interface{}) *RedisClient_XAck_Call {
	return &RedisClient_XAck_Call{Call: _e.mock.On("XAck",
		append([]interface{}{ctx, stream, group}, ids...)...)}
}

func (_c *RedisClient_XAck_Call) Run(run func(ctx context.Context, stream string, group string, ids ...string)) *RedisClient_XAck_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(string), variadicArgs...)
	})
	return _c
}

func (_c *RedisClient_XAck_Call) Return(_a0 *redis.IntCmd) *RedisClient_XAck_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_XAck_Call) RunAndReturn(run func(context.Context, string, string, ...string) *redis.IntCmd) *RedisClient_XAck_Call {
	_c.Call.Return(run)
	return _c
}

// XDel provides a mock function with given fields: ctx, stream, ids
func (_m *RedisClient) XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd {
	_va := make([]interface{}, len(ids))
	for _i := range ids {
		_va[_i] = ids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, stream)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for XDel")
	}

	var r0 *redis.IntCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, ...string) *redis.IntCmd); ok {
		r0 = rf(ctx, stream, ids...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.IntCmd)
		}
	}

	return r0
}

// RedisClient_XDel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'XDel'
type RedisClient_XDel_Call struct {
	*mock.Call
}

// XDel is a helper method to define mock.On call
//   - ctx context.Context
//   - stream string
//   - ids ...string
func (_e *RedisClient_Expecter) XDel(ctx interface{}, stream interface{}, ids ...interface{}) *RedisClient_XDel_Call {
	return &RedisClient_XDel_Call{Call: _e.mock.On("XDel",
		append([]interface{}{ctx, stream}, ids...)...)}
}

func (_c *RedisClient_XDel_Call) Run(run func(ctx context.Context, stream string, ids ...string)) *RedisClient_XDel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *RedisClient_XDel_Call) Return(_a0 *redis.IntCmd) *RedisClient_XDel_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_XDel_Call) RunAndReturn(run func(context.Context, string, ...string) *redis.IntCmd) *RedisClient_XDel_Call {
	_c.Call.Return(run)
	return _c
}

// XGroupCreateMkStream provides a mock function with given fields: ctx, stream, group, start
func (_m *RedisClient) XGroupCreateMkStream(ctx context.Context, stream string, group string, start string) *redis.StatusCmd {
	ret := _m.Called(ctx, stream, group, start)

	if len(ret) == 0 {
		panic("no return value specified for XGroupCreateMkStream")
	}

	var r0 *redis.StatusCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *redis.StatusCmd); ok {
		r0 = rf(ctx, stream, group, start)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.StatusCmd)
		}
	}

	return r0
}

// RedisClient_XGroupCreateMkStream_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'XGroupCreateMkStream'
type RedisClient_XGroupCreateMkStream_Call struct {
	*mock.Call
}

// XGroupCreateMkStream is a helper method to define mock.On call
//   - ctx context.Context
//   - stream string
//   - group string
//   - start string
func (_e *RedisClient_Expecter) XGroupCreateMkStream(ctx interface{}, stream interface{}, group interface{}, start interface{}) *RedisClient_XGroupCreateMkStream_Call {
	return &RedisClient_XGroupCreateMkStream_Call{Call: _e.mock.On("XGroupCreateMkStream", ctx, stream, group, start)}
}

func (_c *RedisClient_XGroupCreateMkStream_Call) Run(run func(ctx context.Context, stream string, group string, start string)) *RedisClient_XGroupCreateMkStream_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *RedisClient_XGroupCreateMkStream_Call) Return(_a0 *redis.StatusCmd) *RedisClient_XGroupCreateMkStream_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_XGroupCreateMkStream_Call) RunAndReturn(run func(context.Context, string, string, string) *redis.StatusCmd) *RedisClient_XGroupCreateMkStream_Call {
	_c.Call.Return(run)
	return _c
}

// XReadGroup provides a mock function with given fields: ctx, a
func (_m *RedisClient) XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	ret := _m.Called(ctx, a)

	if len(ret) == 0 {
		panic("no return value specified for XReadGroup")
	}

	var r0 *redis.XStreamSliceCmd
	if rf, ok := ret.Get(0).(func(context.Context, *redis.XReadGroupArgs) *redis.XStreamSliceCmd); ok {
		r0 = rf(ctx, a)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.XStreamSliceCmd)
		}
	}

	return r0
}

// RedisClient_XReadGroup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'XReadGroup'
type RedisClient_XReadGroup_Call struct {
	*mock.Call
}

// XReadGroup is a helper method to define mock.On call
//   - ctx context.Context
//   - a *redis.XReadGroupArgs
func (_e *RedisClient_Expecter) XReadGroup(ctx interface{}, a interface{}) *RedisClient_XReadGroup_Call {
	return &RedisClient_XReadGroup_Call{Call: _e.mock.On("XReadGroup", ctx, a)}
}

func (_c *RedisClient_XReadGroup_Call) Run(run func(ctx context.Context, a *redis.XReadGroupArgs)) *RedisClient_XReadGroup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*redis.XReadGroupArgs))
	})
	return _c
}

func (_c *RedisClient_XReadGroup_Call) Return(_a0 *redis.XStreamSliceCmd) *RedisClient_XReadGroup_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *RedisClient_XReadGroup_Call) RunAndReturn(run func(context.Context, *redis.XReadGroupArgs) *redis.XStreamSliceCmd) *RedisClient_XReadGroup_Call {
	_c.Call.Return(run)
	return _c
}

// NewRedisClient creates a new instance of RedisClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRedisClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *RedisClient {
	mock := &RedisClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
