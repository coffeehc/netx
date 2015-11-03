package protocol

import (
	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/logger"
	"github.com/ugorji/go/codec"
)

type Msgpack_Protocol struct {
	hander *codec.MsgpackHandle
	interf func() interface{}
}

func NewMsgpackProcotol(interfFunc func() interface{}) *Msgpack_Protocol {
	p := &Msgpack_Protocol{hander: new(codec.MsgpackHandle)}
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (this *Msgpack_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	var b []byte
	encode := codec.NewEncoderBytes(&b, this.hander)
	err := encode.Encode(data)
	if err != nil {
		logger.Error("Msgpack序列化错误:%s", err)
		return
	}
	warp.FireNextEncode(context, b)
}

func (this *Msgpack_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := this.interf()
		decode := codec.NewDecoderBytes(v, this.hander)
		err := decode.Decode(obj)
		if err != nil {
			logger.Error("Msgpack反序列化失败:%s", err)
			return
		}
		data = obj
	}
	warp.FireNextDecode(context, data)
}
