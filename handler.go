package netx

import (
	"context"
)

//Handler 处理器接口
type Handler interface {
	Active(cxt context.Context, context ConnContext) 
	Exception(cxt context.Context, connContext ConnContext, err error)
	Read(cxt context.Context, connContext ConnContext, data interface{})
	Close(context.Context, ConnContext)
}
