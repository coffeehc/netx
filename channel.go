// channle.go
package coffeenet

import (
	"fmt"
	"io"
	"net"

	"github.com/coffeehc/logger"
)

const (
	DEFAULT_BUF_SIZE int = 512
)

type ChannelListen interface {
	OnActive(context *ChannelHandlerContext)
	OnClose(context *ChannelHandlerContext)
	OnException(context *ChannelHandlerContext, err error)
}

type SimpleChannelListen struct {
	ChannelListen
}

func (this *SimpleChannelListen) OnActive(context *ChannelHandlerContext) {
	//do Nothing
}

func (this *SimpleChannelListen) OnClose(context *ChannelHandlerContext) {
	//do Nothing
}
func (this *SimpleChannelListen) OnException(context *ChannelHandlerContext, err error) {
	//do Nothing
}

type ChannelHandler interface {
	Active(context *ChannelHandlerContext)
	Exception(context *ChannelHandlerContext, err error)
	ChannelRead(context *ChannelHandlerContext, data interface{}) error
	ChannelClose(context *ChannelHandlerContext)
}

type ChannelHandlerContext struct {
	id           int
	conn         net.Conn
	handler      ChannelHandler
	headProtocol *ChannelProtocolWarp
	tailProtocol *ChannelProtocolWarp
	isOpen       bool
	listens      []ChannelListen
	remortAddr   net.Addr
	workPool     chan int
}

func NewChannelHandlerContext(id int, conn net.Conn, workPool chan int) *ChannelHandlerContext {
	channelHandlerContext := new(ChannelHandlerContext)
	channelHandlerContext.id = id
	channelHandlerContext.conn = conn
	channelHandlerContext.listens = make([]ChannelListen, 0)
	channelHandlerContext.remortAddr = conn.RemoteAddr()
	channelHandlerContext.workPool = workPool
	return channelHandlerContext
}

func (this *ChannelHandlerContext) AddListen(listen ChannelListen) {
	this.listens = append(this.listens, listen)
}

func (this *ChannelHandlerContext) RemortAddr() net.Addr {
	return this.remortAddr
}

func (this *ChannelHandlerContext) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *ChannelHandlerContext) SetHandler(handler ChannelHandler) {
	this.handler = handler
}

func (this *ChannelHandlerContext) SetProtocols(protocols []ChannelProtocol) {
	var curWarp *ChannelProtocolWarp
	for _, protocol := range protocols {
		warp := newChannelProtocolWarp(protocol)
		if this.headProtocol == nil {
			this.headProtocol = warp
		} else {
			curWarp.next = warp
		}
		warp.prve = curWarp
		curWarp = warp
		warp.bridge(this)
	}
	this.tailProtocol = curWarp
	//TODO 这里需要处理

}

func (this *ChannelHandlerContext) handle() {
	logger.Debugf("已经建立连接:%s->%s", this.conn.LocalAddr(), this.conn.RemoteAddr())
	this.handler.Active(this)
	go func(this *ChannelHandlerContext) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("系统异常:%s", err)
			}
		}()
		for _, l := range this.listens {
			l.OnActive(this)
		}
	}(this)
	this.isOpen = true
	bytes := make([]byte, 1024)
	defer func() {
		if err := recover(); err != nil {
			logger.Debugf("处理连接出现了异常:%s", err)
		}
	}()
	for this.isOpen {
		i, err := this.conn.Read(bytes)
		if err != nil {
			this.fireException(fmt.Errorf("接受内容异常,%s", err))
			if err == io.EOF {
				this.Close()
			}
		}
		if i > 0 {
			this.headProtocol.read(this, bytes[:i])
		}
	}
}

func (this *ChannelHandlerContext) Close() {
	if this.isOpen {
		this.isOpen = false
		err := this.conn.Close()
		if err != nil {
			this.fireException(err)
		}
		logger.Debugf("关闭了连接,%s", this.conn.RemoteAddr().String())
		this.handler.ChannelClose(this)
		this.headProtocol.Destroy()
		go func(this *ChannelHandlerContext) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("系统异常:%s", err)
				}
			}()
			for _, l := range this.listens {
				l.OnClose(this)
			}
		}(this)
	}
}

func (this *ChannelHandlerContext) IsOpen() bool {
	return this.isOpen
}

/*
	处理异常
*/
func (this *ChannelHandlerContext) fireException(err error) {
	logger.Debugf("获取了一个异常事件:%s", err)
	this.handler.Exception(this, err)
	go func(this *ChannelHandlerContext, err error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("系统异常:%s", err)
			}
		}()
		for _, l := range this.listens {
			l.OnException(this, err)
		}
	}(this, err)
}

func (this *ChannelHandlerContext) Write(data interface{}) {
	this.tailProtocol.write(this, data)
}

func (this *ChannelHandlerContext) write(data []byte) {
	_, err := this.conn.Write(data)
	if err != nil {
		this.fireException(err)
	}
}
