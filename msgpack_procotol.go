// msgpack_procotol
package coffeenet

import (
	"logger"
	"msgpackgo"
)

type MsgpackProtocol struct {
	interf func() interface{}
}

func NewMsgpackProtocol(interfFunc func() interface{}) *MsgpackProtocol {
	p := new(MsgpackProtocol)
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (this *MsgpackProtocol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	b, err := msgpackgo.Marshal(data)
	if err != nil {
		logger.Error("Msgpack序列化错误:%s", err)
		return
	}
	warp.FireNextWrite(context, b)
}

func (this *MsgpackProtocol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := this.interf()
		err := msgpackgo.Unmarshal(v, obj)
		if err != nil {
			logger.Error("Msgpack反序列化失败:%s", err)
			return
		}
		warp.FireNextRead(context, obj)
	} else {
		warp.FireNextRead(context, data)
	}

}
