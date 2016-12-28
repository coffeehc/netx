package netx

import (
	"context"
	"math"
	"net"
	"sync/atomic"
)

//InitConnContextFunc init connContext
type InitConnContextFunc func(cxt context.Context, connContext ConnContext)

//_ContextFactory _ContextFactory
type _ContextFactory struct {
	//初始化上下文的方法
	initContextFunc InitConnContextFunc
	//获取ChannelId的种子
	idSeed *int64
	//ChannelHandler组映射
	group map[int64]ConnContext
	//处理goroutine的个数限制i
	workPool chan int64
	//是否顺序处理消息,默认false,即可以并发处理消息
	syncHandler bool

	rootCxt context.Context
}

//newContextFactory 初始化一个ContextFactory
func newContextFactory(initContextFunc InitConnContextFunc) *_ContextFactory {
	var idSeed = int64(0)
	return &_ContextFactory{
		initContextFunc: initContextFunc,
		idSeed:          &idSeed,
		group:           make(map[int64]ConnContext),
		workPool:        nil,
		syncHandler:     false,
		rootCxt:         context.Background(),
	}
}

//获取下一个Id
func (factory *_ContextFactory) nextID() int64 {
	atomic.CompareAndSwapInt64(factory.idSeed, math.MaxInt64, 0)
	return atomic.AddInt64(factory.idSeed, 1)
}

//close 关闭Factory ,将关闭下属所有的context
func (factory *_ContextFactory) close() {
	if factory.group != nil {
		for _, v := range factory.group {
			//TODO set time out
			v.Close(context.Background())
		}
	}
}

//当连接建立的时候创建Channel上下文
func (factory *_ContextFactory) createConnContext(conn net.Conn) ConnContext {
	id := factory.nextID()
	connContext := newConnContext(context.Background(),id, conn, factory.workPool, factory.syncHandler)
	factory.group[id] = connContext
	factory.initContextFunc(factory.rootCxt, connContext)
	return connContext
}
