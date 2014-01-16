package coffeenet

import (
	"bytes"
)

type TerminalProtocol struct {
	ChannelProtocol
	buf *bytes.Buffer
}

func NewTerminalProtocol() *TerminalProtocol {
	p := new(TerminalProtocol)
	p.buf = bytes.NewBuffer(nil)
	return p
}

func (this *TerminalProtocol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	warp.FireNextWrite(context, data)
}
func (this *TerminalProtocol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		for _, d := range v {
			if d == '\n' {
				bs := make([]byte, this.buf.Len())
				this.buf.Read(bs)
				warp.FireNextRead(context, bs)
			} else {
				this.buf.WriteByte(d)
			}
		}
	} else {
		warp.FireNextRead(context, data)
	}
}
