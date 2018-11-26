package rpc

type Middleware func(ctx *Context) error
