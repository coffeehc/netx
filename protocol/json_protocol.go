package protocol

import (
	"context"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
	"github.com/pquerna/ffjson/ffjson"
)

type jsonProtocol struct {
	interf func() interface{}
}

//NewJSONProtocol cteate a json Protocol implement
func NewJSONProtocol(interfFunc func() interface{}) netx.Protocol {
	p := new(jsonProtocol)
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (jp *jsonProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	b, err := ffjson.Marshal(data)
	if err != nil {
		logger.Error("Json序列化错误:%s", err)
		return
	}
	chain.Fire(cxt, connContext, b)
}

func (jp *jsonProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := jp.interf()
		err := ffjson.Unmarshal(v, obj)
		if err != nil {
			logger.Error("Json反序列化失败:%s", err)
			return
		}
		data = obj
	}
	chain.Fire(cxt, connContext, data)

}

func (jp *jsonProtocol) EncodeDestroy() {}

func (jp *jsonProtocol) DecodeDestroy() {}
