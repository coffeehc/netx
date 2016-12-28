package signal

import (
	"context"
	"net"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
)

type netHandler struct {
	factory    *initFactory
	remortAddr net.Addr
}

func (nh *netHandler) Active(cxt context.Context, connContext netx.ConnContext) {
}
func (nh *netHandler) Exception(cxt context.Context, connContext netx.ConnContext, err error) {
	if opErr, ok := err.(*net.OpError); ok {
		logger.Error("出现网络异常:%s", opErr)
		connContext.Close(cxt)
	} else {
		logger.Error("出现了业务异常:%s", err)
	}
}
func (nh *netHandler) Read(cxt context.Context, connContext netx.ConnContext, data interface{}) {
	if signal, ok := data.(*Signal); ok {
		handler := nh.factory.getHandler(signal.GetSignal())
		if handler != nil {
			handler.Handle(connContext, signal)
			return
		}
		logger.Error("信令[0x%X]没有对应的处理类", signal.GetSignal())
		return
	}
	logger.Error("处理的对象并非Signal类型:[%T]%#v", data, data)

}
func (nh *netHandler) Close(cxt context.Context, connContext netx.ConnContext) {
}
