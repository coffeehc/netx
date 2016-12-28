package netx

import (
	"context"
)

//ContextListen  listen
type ContextListen interface {
	OnActive(context.Context, ConnContext)
	OnClose(context.Context, ConnContext)
	OnException(context.Context, ConnContext, error)
}
