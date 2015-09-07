// content
package coffeenet

import (
	"math"
	"net"
	"sync/atomic"
)

type ContextFactory struct {
	//初始化上下文的方法
	initContextFunc func(content *Context)
	//获取ChannelId的种子
	idSeed int64
	//ChannelHandler组映射
	group map[int64]*Context
	//处理goroutine的个数限制
	workPool chan int64
	//是否顺序处理消息,默认false,即可以并发处理消息
	orderHandler bool
	handlerStat  *HanderStat
}

//初始化一个ContextFactory
func NewContextFactory(initContextFunc func(context *Context)) *ContextFactory {
	return &ContextFactory{initContextFunc, 0, make(map[int64]*Context), nil, false, nil}
}

//获取下一个Id
func (this *ContextFactory) nextId() int64 {
	atomic.CompareAndSwapInt64(&this.idSeed, math.MaxInt64, 0)
	return atomic.AddInt64(&this.idSeed, 1)
}

//关闭Factory ,将关闭下属所有的context
func (this *ContextFactory) Close() {
	if this.group != nil {
		for _, v := range this.group {
			v.Close()
		}
	}
}

//当连接建立的时候创建Channel上下文
func (this *ContextFactory) creatContext(conn net.Conn) *Context {
	id := this.nextId()
	context := newContext(id, conn, this.workPool, this.orderHandler)
	context.handlerStat = this.handlerStat
	this.group[id] = context
	this.initContextFunc(context)
	if context.headProtocol == nil {
		p := newProtocolWarp(new(defaultProtocol))
		context.headProtocol = p
		context.tailProtocol = p
	}
	return context
}
