package netx

import (
	"context"

	"fmt"
	"io"
	"net"

	"errors"
	"github.com/coffeehc/logger"
)

//Protocol protocol struct
type Protocol interface {
	//用于激活 Protocol 的后台任务
	//Start(cxt context.Context, context ConnContext)// TODO 待考虑
	//数据编码
	Encode(cxt context.Context, context ConnContext, chain ProtocolChain, data interface{})
	//数据解码
	Decode(cxt context.Context, context ConnContext, chain ProtocolChain, data interface{})

	EncodeDestroy()

	DecodeDestroy()
}

//NewProtocolChain 创建调用链
func newProtocolChain(protocol Protocol) (encoder *_ProtocolHandler, decoder *_ProtocolHandler) {
	return &_ProtocolHandler{
			handler: protocol.Encode,
			destroy: protocol.EncodeDestroy,
			do:      _write,
		}, &_ProtocolHandler{
			handler: protocol.Decode,
			destroy: protocol.DecodeDestroy,
			do:      _read,
		}
}

//ProtocolChain Protocol chain interface
type ProtocolChain interface {
	Fire(cxt context.Context, connContext ConnContext, data interface{})
}

//ProtocolChain 协议包装,用于调用下一个协议编码或者下一个协议解码
type _ProtocolHandler struct {
	handler func(cxt context.Context, connContext ConnContext, chain ProtocolChain, data interface{})
	destroy func()
	next    *_ProtocolHandler
	do      func(cxt context.Context, connContext *_ConnContext, data interface{})
}

func (ph *_ProtocolHandler) Fire(cxt context.Context, connContext ConnContext, data interface{}) {
	if ph.next != nil {
		ph.handler(cxt, connContext, ph.next, data)
		return
	}
	ph.do(cxt, connContext.(*_ConnContext), data)
}

func (ph *_ProtocolHandler) addNextChain(next *_ProtocolHandler) {
	if ph.next == nil {
		ph.next = next
		return
	}
	ph.next.addNextChain(next)
}

//Destroy 回收,主要用于buf的回收
func (ph *_ProtocolHandler) Destroy() {
	ph.destroy()
	if ph.next != nil {
		ph.next.Destroy()
	}
}

type emptyProtocol struct {
}

func (*emptyProtocol) Start(cxt context.Context, context ConnContext) {

}

func (*emptyProtocol) Encode(cxt context.Context, connContext ConnContext, chain ProtocolChain, data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("处理数据时出现了不可恢复的异常:%s", err)
			connContext.Close(cxt)
		}
	}()
	chain.Fire(cxt, connContext, data)
}
func (*emptyProtocol) Decode(cxt context.Context, connContext ConnContext, chain ProtocolChain, data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("处理数据时出现了不可恢复的异常:%s", err)
			connContext.Close(cxt)
		}
	}()
	chain.Fire(cxt, connContext, data)
}

func (prorocol *emptyProtocol) EncodeDestroy() {
	//nothing
}

func (prorocol *emptyProtocol) DecodeDestroy() {
	//nothing
}

func _write(cxt context.Context, c *_ConnContext, data interface{}) {
	byteData, ok := data.([]byte)
	if !ok {
		logger.Error("发送的数据类型不是[]byte")
		c.FireException(cxt, errors.New("发送的数据类型不是[]byte"))
		return

	}
	_, err := c.conn.Write(byteData)
	if err != nil {
		if err == io.EOF {
			c.Close(cxt)
		}
		if opErr, ok := err.(*net.OpError); ok {
			if !opErr.Timeout() && !opErr.Temporary() {
				logger.Error("接收到不可恢复的异常,关闭连接,%s", err)
				c.Close(cxt)
			}
		} else {
			c.FireException(cxt, fmt.Errorf("发送内容异常,%#v", err))
		}
	}
}

//处理封装好的数据
func _read(cxt context.Context, c *_ConnContext, data interface{}) {
	c.workPool <- c.id
	if c.syncHandle {
		c.handler.Read(cxt, c, data)
	} else {
		go c.handler.Read(cxt, c, data)
	}
}
