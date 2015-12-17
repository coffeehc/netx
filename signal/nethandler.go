// nethandler
package signal

import (
	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/logger"
	"net"
)

type netHandler struct {
	factory    *initFactory
	remortAddr net.Addr
}

func (this *netHandler) Active(context *coffeenet.Context) {
}
func (this *netHandler) Exception(context *coffeenet.Context, err error) {
	if opErr, ok := err.(*net.OpError); ok {
		logger.Error("出现网络异常:%s", opErr)
		context.Close()
	} else {
		logger.Error("出现了业务异常:%s", err)
	}
}
func (this *netHandler) Read(context *coffeenet.Context, data interface{}) {
	if signal, ok := data.(*Signal); ok {
		handler := this.factory.getHandler(signal.GetSignal())
		if handler != nil {
			handler.Handle(context, signal)
			return
		}
		logger.Error("信令[0x%X]没有对应的处理类",signal.GetSignal())
		return
	}
	logger.Error("处理的对象并非Signal类型:[%T]%#v", data, data)

}
func (this *netHandler) Close(context *coffeenet.Context) {
}
