// procotol
package coffeenet

import (
	"fmt"

	"github.com/coffeehc/logger"
)

type Protocol interface {
	//数据编码
	Encode(context *Context, warp *ProtocolWarp, data interface{})
	//数据解码
	Decode(context *Context, warp *ProtocolWarp, data interface{})
}
type ProtocolDestroy interface {
	//回收
	Destroy()
}

//协议包装,用于调用下一个协议编码或者下一个协议解码
type ProtocolWarp struct {
	protocol Protocol
	prve     *ProtocolWarp
	next     *ProtocolWarp
}

func newProtocolWarp(protocol Protocol) *ProtocolWarp {
	warp := new(ProtocolWarp)
	warp.protocol = protocol
	return warp
}

func (this *ProtocolWarp) decode(context *Context, data interface{}) {
	this.protocol.Decode(context, this, data)
}

//调用下一个协议读取数据
func (this *ProtocolWarp) FireNextDecode(context *Context, data interface{}) {
	if data == nil {
		logger.Warn("Data is nil")
		return
	}
	warp := this.next
	if warp != nil {
		warp.decode(context, data)
	} else {
		context.handle(data)
	}
}

func (this *ProtocolWarp) encode(context *Context, data interface{}) {
	this.protocol.Encode(context, this, data)
}

//调用下一个协议写入数据
func (this *ProtocolWarp) FireNextEncode(context *Context, data interface{}) {
	if data == nil {
		return
	}
	warp := this.prve
	if warp != nil {
		warp.encode(context, data)
	} else {
		if v, ok := data.([]byte); ok {
			context.write(v)
		} else {
			context.fireException(fmt.Errorf("发送的数据不能转换为byte数组"))
		}
	}
}

//回收,主要用于buf的回收
func (this *ProtocolWarp) Destroy() {
	if v, ok := this.protocol.(ProtocolDestroy); ok {
		v.Destroy()
	}
	if this.next != nil {
		this.next.Destroy()
	}
}

type defaultProtocol struct {
}

func (this *defaultProtocol) Encode(context *Context, warp *ProtocolWarp, data interface{}) {
	warp.FireNextEncode(context, data)
}
func (this *defaultProtocol) Decode(context *Context, warp *ProtocolWarp, data interface{}) {
	warp.FireNextDecode(context, data)
}
