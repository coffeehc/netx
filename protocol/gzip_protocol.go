package protocol

import (
	"bytes"
	"compress/gzip"
	"io"

	"context"

	"github.com/coffeehc/netx"
)

//NewGizpProtocol cteate a gzip protocol implement
func NewGizpProtocol() netx.Protocol {
	return &gizpProtocol{}
}

type gizpProtocol struct {
}

func (gp *gizpProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if b, ok := data.([]byte); ok {
		buf := bytes.NewBuffer(nil)
		writer := gzip.NewWriter(buf)
		_, err := writer.Write(b)
		writer.Close()
		if err == nil {
			data = buf.Bytes()
		}
	}
	chain.Fire(cxt, connContext, data)
}

func (gp *gizpProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	if b, ok := data.([]byte); ok {
		buf := bytes.NewBuffer(nil)
		reader, err := gzip.NewReader(bytes.NewBuffer(b))
		if err == nil {
			_, err = io.Copy(buf, reader)
			reader.Close()
			if err == nil {
				data = buf.Bytes()
			}
		}
	}
	chain.Fire(cxt, connContext, data)
}

func (gp *gizpProtocol) EncodeDestroy() {}

func (gp *gizpProtocol) DecodeDestroy() {}
