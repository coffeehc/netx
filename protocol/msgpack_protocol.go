package protocol

import (
	"context"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
	"github.com/ugorji/go/codec"
)

type msgpackProtocol struct {
	hander *codec.MsgpackHandle
	interf func() interface{}
}

//NewMsgpackProcotol cteate a Msgpack Protocol implement
func NewMsgpackProcotol(interfFunc func() interface{}) netx.Protocol {
	p := &msgpackProtocol{hander: new(codec.MsgpackHandle)}
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (mp *msgpackProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	var b []byte
	encode := codec.NewEncoderBytes(&b, mp.hander)
	err := encode.Encode(data)
	if err != nil {
		logger.Error("Msgpack序列化错误:%s", err)
		return
	}
	chain.Fire(cxt, connContext, b)
}

func (mp *msgpackProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := mp.interf()
		decode := codec.NewDecoderBytes(v, mp.hander)
		err := decode.Decode(obj)
		if err != nil {
			logger.Error("Msgpack反序列化失败:%s", err)
			return
		}
		data = obj
	}
	chain.Fire(cxt, connContext, data)
}

func (mp *msgpackProtocol) EncodeDestroy() {}

func (mp *msgpackProtocol) DecodeDestroy() {}
