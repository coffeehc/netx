package protocol

import (
	"context"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
	"github.com/golang/snappy"
)

//NewSnappyProtocol cteate a Snappy Protocol implement
func NewSnappyProtocol() netx.Protocol {
	return &snappyProtocol{}
}

type snappyProtocol struct {
}

func (sp *snappyProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.(string); ok {
		data = []byte(v)
	}
	if v, ok := data.([]byte); ok {
		data = snappy.Encode(nil, v)
		if data == nil {
			logger.Warn("snappy编码出错:%s")
			return
		}
	}
	chain.Process(cxt, connContext, data)
}
func (sp *snappyProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		var err error
		data, err = snappy.Decode(nil, v)
		if err != nil {
			logger.Warn("snappy解码出错:%s", err)
			return
		}
	}
	chain.Process(cxt, connContext, data)
}

func (sp *snappyProtocol) EncodeDestroy() {}

func (sp *snappyProtocol) DecodeDestroy() {}
