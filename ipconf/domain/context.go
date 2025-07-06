package domain

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type ClientContext struct {
	IP string `json:"ip"`
}

type IPConfContext struct {
	Ctx       *context.Context
	AppCtx    *app.RequestContext
	ClientCtx *ClientContext
}

// BuildIPConfContext constructs an IPConfContext from the given context and app.RequestContext.
func BuildIPConfContext(c *context.Context, ctx *app.RequestContext) *IPConfContext {
	return &IPConfContext{
		Ctx:       c,
		AppCtx:    ctx,
		ClientCtx: &ClientContext{IP: ctx.ClientIP()},
	}
}
