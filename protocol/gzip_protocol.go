package protocol

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/coffeehc/coffeenet"
)

type Gizp_Protocol struct {
}

func (this *Gizp_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	if b, ok := data.([]byte); ok {
		buf := bytes.NewBuffer(nil)
		writer := gzip.NewWriter(buf)
		_, err := writer.Write(b)
		writer.Close()
		if err == nil {
			data = buf.Bytes()
		}
	}
	warp.FireNextEncode(context, data)
}

func (this *Gizp_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
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
	warp.FireNextDecode(context, data)
}
