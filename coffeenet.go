// coffeenet project coffeenet.go
package coffeenet

import (
	"net"
)

type BootStrap struct {
	host                         string //127.0.0.1:8888
	netType                      string
	channelHandlerContextFactory *ChannelHandlerContextFactory
	open                         bool
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

func (this *ChannelHandlerContextFactory) CreatChannelHandlerContext(conn net.Conn) *ChannelHandlerContext {
	this.idSeed++
	channelHandlerContext := NewChannelHandlerContext(this.idSeed, conn)
	this.initContextFunc(channelHandlerContext)
	if channelHandlerContext.headProtocol == nil {
		p := NewChannelProtocolWarp(new(defaultChannelProtocol))
		channelHandlerContext.headProtocol = p
		channelHandlerContext.tailProtocol = p
	}
	return channelHandlerContext
}
