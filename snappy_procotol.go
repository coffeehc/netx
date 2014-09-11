// snappy_procotol
package coffeenet

import (
	"logger"

	"code.google.com/p/snappy-go/snappy"
)

type SnappyProtocol struct {
}

func (this *SnappyProtocol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.(string); ok {
		data = []byte(v)
	}
	if v, ok := data.([]byte); ok {
		var err error
		data, err = snappy.Encode(nil, v)
		if err != nil {
			logger.Warn("snappy压缩出错:%s", err)
			return
		}
	}
	warp.FireNextWrite(context, data)
}
func (this *SnappyProtocol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		var err error
		data, err = snappy.Decode(nil, v)
		if err != nil {
			logger.Warn("snappy压缩出错:%s", err)
			return
		}
	}
	warp.FireNextRead(context, data)
}
