// msgpack_procotol
package coffeenet

import (
	"github.com/coffeehc/logger"
	"github.com/ugorji/go/codec"
)

type MsgpackProcotol struct {
	hander *codec.MsgpackHandle
	interf func() interface{}
}

func NewMsgpackProcotol(interfFunc func() interface{}) *MsgpackProcotol {
	p := &MsgpackProcotol{hander: new(codec.MsgpackHandle)}
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (this *MsgpackProcotol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	var b []byte
	encode := codec.NewEncoderBytes(&b, this.hander)
	err := encode.Encode(data)
	if err != nil {
		logger.Error("Msgpack序列化错误:%s", err)
		return
	}
	warp.FireNextWrite(context, b)
}

func (this *MsgpackProcotol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := this.interf()
		decode := codec.NewDecoderBytes(v, this.hander)
		err := decode.Decode(obj)
		if err != nil {
			logger.Error("Msgpack反序列化失败:%s", err)
			return
		}
		warp.FireNextRead(context, obj)
	} else {
		warp.FireNextRead(context, data)
	}

}
