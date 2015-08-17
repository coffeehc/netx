// coffeenet project coffeenet.go
package coffeenet

import (
	"net"
	"sync"
	"sync/atomic"
)

type BootStrap struct {
	idSeed         int32
	group          map[int32]*ChannelHandlerContext
	open           *sync.Once
	workConcurrent int
	workPool       chan int
	listen         *closeListen
}

func (this *BootStrap) GetSeed() int32 {
	return atomic.AddInt32(&this.idSeed, 1)
}

//初始化Bootstrap
func (this *BootStrap) init() {
	if this.workConcurrent < 0 {
		panic("工作并发不能小于0")
	}
	if this.workConcurrent == 0 {
		this.workConcurrent = 1
	}
	this.open = new(sync.Once)
	this.workPool = make(chan int, this.workConcurrent)
	this.listen = new(closeListen)
	this.idSeed = 0
}

type ChannelHandlerContextFactory struct {
	initContextFunc func(content *ChannelHandlerContext)
	bootStrap       *BootStrap
}

func (this *BootStrap) Close() {
	if this.group != nil {
		for _, v := range this.group {
			v.Close()
		}
	}
}

type closeListen struct {
	SimpleChannelListen
	server *Server
}

func (this *closeListen) OnActive(context *ChannelHandlerContext) {
	this.server.group[context.id] = context
}

func (this *closeListen) OnClose(context *ChannelHandlerContext) {
	delete(this.server.group, context.id)
	context.listens = nil
}

func NewChannelHandlerContextFactory(initContextFunc func(context *ChannelHandlerContext)) *ChannelHandlerContextFactory {
	this := new(ChannelHandlerContextFactory)
	this.initContextFunc = initContextFunc
	return this
}

func (this *ChannelHandlerContextFactory) CreatChannelHandlerContext(conn net.Conn, workPool chan int) *ChannelHandlerContext {
	seed := this.bootStrap.GetSeed()
	channelHandlerContext := NewChannelHandlerContext(seed, conn, workPool)
	this.initContextFunc(channelHandlerContext)
	if channelHandlerContext.headProtocol == nil {
		p := newChannelProtocolWarp(new(defaultChannelProtocol))
		channelHandlerContext.headProtocol = p
		channelHandlerContext.tailProtocol = p
	}
	this.bootStrap.group[seed] = channelHandlerContext
	return channelHandlerContext
}
