// keepalive_procotol
package coffeenet

import (
	"bytes"
	"time"
)

type KeepAliveProcotol struct {
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	readChan     chan bool
	writeChan    chan bool
	isDestroy    bool
}

var KEEP_ALIVE_MSG = []byte{0XFE, 0xFF, 'k', 'e', 'e', 'p', 'a', 'l', 'i', 'v', 'e'}

func NewKeepAliveProcotol(readTimeOut, writeTimeOut time.Duration) *KeepAliveProcotol {
	keeper := &KeepAliveProcotol{readTimeOut: readTimeOut, writeTimeOut: writeTimeOut}
	keeper.readChan = make(chan bool)
	keeper.writeChan = make(chan bool)
	return keeper
}

func (this *KeepAliveProcotol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	this.readChan <- true
	warp.FireNextWrite(context, data)
}
func (this *KeepAliveProcotol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	this.readChan <- true
	if v, ok := data.([]byte); ok {
		i := bytes.Index(v, KEEP_ALIVE_MSG)
		if i >= 0 {
			if i == 0 {
				this.Decode(context, warp, v[len(KEEP_ALIVE_MSG):])
			} else {
				this.Decode(context, warp, append(v[:i], v[i+len(KEEP_ALIVE_MSG):]...))
			}
			return
		}
	}
	warp.FireNextRead(context, data)
}
func (this *KeepAliveProcotol) Destrop() {
	this.isDestroy = true
	close(this.readChan)
	close(this.writeChan)
}

func (this *KeepAliveProcotol) SetSelfWarp(context *ChannelHandlerContext, warp ChannelProtocolWarp) {
	if this.readTimeOut != 0 {
		go func() {
			for !this.isDestroy {
				select {
				case <-time.After(this.readTimeOut):
					warp.FireNextWrite(context, KEEP_ALIVE_MSG)
				case <-this.readChan:
				}
			}
		}()
	}
	if this.writeTimeOut != 0 {
		go func() {
			for !this.isDestroy {
				select {
				case <-time.After(this.writeTimeOut):
					warp.FireNextWrite(context, KEEP_ALIVE_MSG)
				case <-this.writeChan:
				}
			}
		}()
	}
}
