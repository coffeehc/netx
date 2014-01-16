// coffeenet project coffeenet.go
package coffeenet

import (
	"net"
)

type BootStrap struct {
	host                         string
	netType                      string
	channelHandlerContextFactory *ChannelHandlerContextFactory
	open                         bool
	workConcurrent               int
	workPool                     chan int
}

func (this *BootStrap) initWorkPool() {
	if this.workConcurrent < 0 {
		panic("工作并发不能小于0")
	}
	if this.workConcurrent == 0 {
		this.workConcurrent = 1
	}
	this.workPool = make(chan int, this.workConcurrent)
}

type ChannelHandlerContextFactory struct {
	idSeed          int
	group           map[int]ChannelHandlerContext
	initContextFunc func(content *ChannelHandlerContext)
}

func (this *BootStrap) SetChannelHandlerContextFactory(factory *ChannelHandlerContextFactory) {
	this.channelHandlerContextFactory = factory
}

func NewChannelHandlerContextFactory(initContextFunc func(context *ChannelHandlerContext)) *ChannelHandlerContextFactory {
	this := new(ChannelHandlerContextFactory)
	this.group = make(map[int]ChannelHandlerContext)
	this.initContextFunc = initContextFunc
	return this
}

func (this *ChannelHandlerContextFactory) CreatChannelHandlerContext(conn net.Conn, workPool chan int) *ChannelHandlerContext {
	this.idSeed++
	channelHandlerContext := NewChannelHandlerContext(this.idSeed, conn, workPool)
	this.initContextFunc(channelHandlerContext)
	if channelHandlerContext.headProtocol == nil {
		p := newChannelProtocolWarp(new(defaultChannelProtocol))
		channelHandlerContext.headProtocol = p
		channelHandlerContext.tailProtocol = p
	}
	return channelHandlerContext
}
