package protocol

import (
	"bytes"
	"context"

	"github.com/coffeehc/netx"
)

type terminalProtocol struct {
	buf *bytes.Buffer
}

//NewTerminalProtocol cteate a Terminal Protocol implement
func NewTerminalProtocol() netx.Protocol {
	p := new(terminalProtocol)
	p.buf = bytes.NewBuffer(nil)
	return p
}

func (tp *terminalProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	chain.Fire(cxt, connContext, data)
}
func (tp *terminalProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if v, ok := data.([]byte); ok {
		for _, d := range v {
			if d == '\n' {
				bs := make([]byte, tp.buf.Len())
				tp.buf.Read(bs)
				chain.Fire(cxt, connContext, bs)
			} else {
				tp.buf.WriteByte(d)
			}
		}
	} else {
		chain.Fire(cxt, connContext, data)
	}
}

func (tp *terminalProtocol) EncodeDestroy() {}

func (tp *terminalProtocol) DecodeDestroy() {}
