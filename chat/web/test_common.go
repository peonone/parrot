package web

import (
	"reflect"

	"github.com/stretchr/testify/mock"
)

type mockUserClient struct {
	mock.Mock
}

func (c *mockUserClient) ReadJSON(v interface{}) error {
	returnVals := c.Called(v)
	// TODO use the first of return values as the value need to update the arg pointer
	// need find the best practice to do it
	ptrVal := reflect.ValueOf(v).Elem()
	vv := reflect.ValueOf(returnVals.Get(0))
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	ptrVal.Set(vv)
	return returnVals.Error(1)
}

func (c *mockUserClient) WriteJSON(v interface{}) error {
	returnVals := c.Called(v)
	return returnVals.Error(0)
}
