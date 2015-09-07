package protocol

import (
	"encoding/json"

	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/logger"
)

type JsonProtocol struct {
	interf func() interface{}
}

func NewJsonProtocol(interfFunc func() interface{}) *JsonProtocol {
	p := new(JsonProtocol)
	p.interf = interfFunc
	if p.interf == nil {
		p.interf = func() interface{} {
			var i interface{}
			return &i
		}
	}
	return p
}

func (this *JsonProtocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		logger.Error("Json序列化错误:%s", err)
		return
	}
	warp.FireNextEncode(context, b)
}

func (this *JsonProtocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		obj := this.interf()
		err := json.Unmarshal(v, obj)
		if err != nil {
			logger.Error("Json反序列化失败:%s", err)
			return
		}
		data = obj
	}
	warp.FireNextDecode(context, data)

}
