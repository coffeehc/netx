package protocol

import (
	"github.com/coffeehc/coffeenet"
	"github.com/coffeehc/logger"
	"github.com/golang/snappy"
)

type Snappy_Protocol struct {
}

func (this *Snappy_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
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
	warp.FireNextEncode(context, data)
}
func (this *Snappy_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		var err error
		data, err = snappy.Decode(nil, v)
		if err != nil {
			logger.Warn("snappy解码出错:%s", err)
			return
		}
	}
	warp.FireNextDecode(context, data)
}
