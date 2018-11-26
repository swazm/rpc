package rpc

import "net/http"

func NewContext(request *http.Request) *Context {
	return &Context{
		store:   make(map[string]interface{}),
		Request: request,
	}
}

type Context struct {
	Request *http.Request
	store   map[string]interface{}
}

func (c *Context) Get(key string) interface{} {
	return c.store[key]
}
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

func (c *Context) Exists(key string) bool {
	if c.store[key] != nil {
		return true
	}
	return false
}
