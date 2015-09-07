package protocol

import (
	"bytes"

	"github.com/coffeehc/coffeenet"
)

type Terminal_Protocol struct {
	buf *bytes.Buffer
}

func NewTerminalProtocol() *Terminal_Protocol {
	p := new(Terminal_Protocol)
	p.buf = bytes.NewBuffer(nil)
	return p
}

func (this *Terminal_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	warp.FireNextEncode(context, data)
}
func (this *Terminal_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if v, ok := data.([]byte); ok {
		for _, d := range v {
			if d == '\n' {
				bs := make([]byte, this.buf.Len())
				this.buf.Read(bs)
				warp.FireNextDecode(context, bs)
			} else {
				this.buf.WriteByte(d)
			}
		}
	} else {
		warp.FireNextDecode(context, data)
	}
}
