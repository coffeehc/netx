
package protocol

import (
	"context"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
	"github.com/golang/protobuf/proto"
)

type protoBufProtocol struct {
	getMessageImpl func() proto.Message
}

//NewProtoBufProcotol cteate a ProtoBuf Protocol implement
func NewProtoBufProcotol(getMessageImpl func() proto.Message) netx.Protocol {
	p := &protoBufProtocol{}
	p.getMessageImpl = getMessageImpl
	if p.getMessageImpl == nil {
		logger.Error("没有指定proto.Message实例化接口.")
		return nil
	}
	return p
}

func (pp *protoBufProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.(proto.Message); ok {
		buf, err := proto.Marshal(v)
		if err != nil {
			logger.Error("ProtoBuf 序列化失败: ", err)
			return
		}
		data = buf
	}
	chain.Process(cxt, connContext, data)
}

func (pp *protoBufProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		message := pp.getMessageImpl()
		err := proto.Unmarshal(v, message)
		if err != nil {
			logger.Error("ProtoBuf反序列化失败:%s", err)
			return
		}
		data = message
	}
	chain.Process(cxt, connContext, data)
}

func (pp *protoBufProtocol) EncodeDestroy() {}

func (pp *protoBufProtocol) DecodeDestroy() {}
