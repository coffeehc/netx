// protobuf_protocol
package protocol

import (
	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/logger"
	"github.com/golang/protobuf/proto"
)

type ProtoBuf_Protocol struct {
	getMessageImpl func() proto.Message
}

func NewProtoBufProcotol(getMessageImpl func() proto.Message) *ProtoBuf_Protocol {
	p := &ProtoBuf_Protocol{}
	p.getMessageImpl = getMessageImpl
	if p.getMessageImpl == nil {
		logger.Error("没有指定proto.Message实例化接口.")
		return nil
	}
	return p
}

func (this *ProtoBuf_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.(proto.Message); ok {
		buf, err := proto.Marshal(v)
		if err != nil {
			logger.Error("ProtoBuf 序列化失败: ", err)
			return
		}
		data = buf
	}
	warp.FireNextEncode(context, data)
}

func (this *ProtoBuf_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		message := this.getMessageImpl()
		err := proto.Unmarshal(v, message)
		if err != nil {
			logger.Error("ProtoBuf反序列化失败:%s", err)
			return
		}
		data = message
	}
	warp.FireNextDecode(context, data)
}
