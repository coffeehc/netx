package protocol

import (
	"bytes"
	"context"
	"time"

	"github.com/coffeehc/netx"
)

type keepAliveProtocol struct {
	readTimeOut  time.Duration
	writeTimeOut time.Duration
	readChan     chan bool
	writeChan    chan bool
	isDestroy    bool
	msg          []byte
}

var defaultKeepAliveMsg = []byte{0XFE, 0xFF, 'k', 'e', 'e', 'p', 'a', 'l', 'i', 'v', 'e'}

//NewKeepAliveProtocol cteate a KeepAlive Protocol implement
func NewKeepAliveProtocol(readTimeOut, writeTimeOut time.Duration, msg []byte) netx.Protocol {
	keeper := &keepAliveProtocol{
		readTimeOut:  readTimeOut,
		writeTimeOut: writeTimeOut,
		readChan:     make(chan bool),
		writeChan:    make(chan bool),
	}
	if msg == nil || len(msg) == 0 {
		msg = defaultKeepAliveMsg
	}
	keeper.msg = msg

	//TODO schedule task
	return keeper
}

func (kp *keepAliveProtocol) Encode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	kp.readChan <- true
	chain.Process(cxt, connContext, data)
}
func (kp *keepAliveProtocol) Decode(cxt context.Context, connContext netx.ConnContext, chain netx.ProtocolChain, data interface{}) {
	kp.readChan <- true
	if v, ok := data.([]byte); ok {
		i := bytes.Index(v, kp.msg)
		if i >= 0 {
			if i == 0 {
				kp.Decode(cxt, connContext, chain, v[len(kp.msg):])
			} else {
				kp.Decode(cxt, connContext, chain, append(v[:i], v[i+len(kp.msg):]...))
			}
			return
		}
	}
	chain.Process(cxt, connContext, data)
}

func (kp *keepAliveProtocol) EncodeDestroy() {
	kp.isDestroy = true
	close(kp.writeChan)
}

func (kp *keepAliveProtocol) DecodeDestroy() {
	kp.isDestroy = true
	close(kp.readChan)
}

// func (this *KeepAlive_Protocol) SetSelfWarp(cxt context.Context, connContext netx.ConnContext, warp *netx.ProtocolWarp) {
// 	if this.readTimeOut != 0 {
// 		go func() {
// 			timer := time.NewTimer(0)
// 			for !this.isDestroy {
// 				timer.Reset(this.readTimeOut)
// 				select {
// 				case <-timer.C:
// 					warp.FireNextEncode(cxt, connContext, KEEP_ALIVE_MSG)
// 				case <-this.readChan:
// 				}
// 			}
// 			timer.Stop()
// 		}()
// 	}
// 	if this.writeTimeOut != 0 {
// 		go func() {
// 			timer := time.NewTimer(0)
// 			for !this.isDestroy {
// 				timer.Reset(this.writeTimeOut)
// 				select {
// 				case <-timer.C:
// 					warp.FireNextEncode(cxt, connContext, KEEP_ALIVE_MSG)
// 				case <-this.writeChan:
// 				}
// 			}
// 			timer.Stop()
// 		}()
// 	}
// }
