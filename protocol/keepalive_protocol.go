// keepalive_procotol
package protocol

import (
	"bytes"
	"time"

	"github.com/coffeehc/coffeenet"
)

type KeepAlive_Protocol struct {
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	readChan     chan bool
	writeChan    chan bool
	isDestroy    bool
}

var KEEP_ALIVE_MSG = []byte{0XFE, 0xFF, 'k', 'e', 'e', 'p', 'a', 'l', 'i', 'v', 'e'}

func NewKeepAliveProtocol(readTimeOut, writeTimeOut time.Duration) *KeepAlive_Protocol {
	keeper := &KeepAlive_Protocol{readTimeOut: readTimeOut, writeTimeOut: writeTimeOut}
	keeper.readChan = make(chan bool)
	keeper.writeChan = make(chan bool)
	return keeper
}

func (this *KeepAlive_Protocol) Encode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
	this.readChan <- true
	warp.FireNextEncode(context, data)
}
func (this *KeepAlive_Protocol) Decode(context *coffeenet.Context, warp *coffeenet.ProtocolWarp, data interface{}) {
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
	warp.FireNextDecode(context, data)
}
func (this *KeepAlive_Protocol) Destrop() {
	this.isDestroy = true
	close(this.readChan)
	close(this.writeChan)
}

func (this *KeepAlive_Protocol) SetSelfWarp(context *coffeenet.Context, warp *coffeenet.ProtocolWarp) {
	if this.readTimeOut != 0 {
		go func() {
			timer := time.NewTimer(0)
			for !this.isDestroy {
				timer.Reset(this.readTimeOut)
				select {
				case <-timer.C:
					warp.FireNextEncode(context, KEEP_ALIVE_MSG)
				case <-this.readChan:
				}
			}
			timer.Stop()
		}()
	}
	if this.writeTimeOut != 0 {
		go func() {
			timer := time.NewTimer(0)
			for !this.isDestroy {
				timer.Reset(this.writeTimeOut)
				select {
				case <-timer.C:
					warp.FireNextEncode(context, KEEP_ALIVE_MSG)
				case <-this.writeChan:
				}
			}
			timer.Stop()
		}()
	}
}
