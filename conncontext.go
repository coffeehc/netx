package netx

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"context"
	"sync"

	"github.com/coffeehc/logger"
)

const (
	contextKeyID          = "__ID__"
	contextKeySyncHandler = "__syncHandler__"
	//contextKey
)

//ConnContext 建立建立后的上下文
type ConnContext interface {
	GetID() int64
	AddListen(name string, listen ContextListen) error
	RemoveListen(name string)
	GetListens() map[string]ContextListen
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	IsOpen() bool
	SetHandler(handler Handler)
	SetProtocols(protocols ...Protocol)
	Start(cxt context.Context)
	Close(cxt context.Context) error
	Write(cxt context.Context, data interface{})
	FireException(cxt context.Context, err error)
}

//初始化一个新的上下文
//id:指定的上下文编号
//conn:网络连接
//workPool:用于控制并行处理工作池,主要用于标记
//orderHandler:标记该上下文处理数据的时候是否顺序处理,默认为并行,否则为当前数据处理完之后再处理接下来的数据
func newConnContext(cxt context.Context, id int64, conn net.Conn, workPool chan int64, syncHandle bool) ConnContext {
	encoder, decoder := newProtocolChain(new(emptyProtocol))
	connContext := &_ConnContext{
		id:         id,
		conn:       conn,
		listens:    make(map[string]ContextListen, 0),
		workPool:   workPool,
		writing:    0,
		once:       new(sync.Once),
		encoder:    encoder,
		decoder:    decoder,
		syncHandle: syncHandle,
	}
	return connContext
}

type _ConnContext struct {
	conn       net.Conn // socket connection
	handler    Handler  //biz Handler
	encoder    *_ProtocolHandler
	decoder    *_ProtocolHandler
	isOpen     bool //通道是否打开
	listens    map[string]ContextListen
	workPool   chan int64
	writing    int32 //关闭的时候用于标记剩余多少数据没有写完
	once       *sync.Once
	id         int64 //channel id
	syncHandle bool
}

//GetID 获取上下文对应的ID
func (c *_ConnContext) GetID() int64 {
	return c.id
}

func (c *_ConnContext) GetListens() map[string]ContextListen {
	return c.listens
}

//AddListen 添加监听器
func (c *_ConnContext) AddListen(name string, listen ContextListen) error {
	if _, ok := c.listens[name]; ok {
		return fmt.Errorf("监听器[%s]已经存在,不能添加", name)
	}
	c.listens[name] = listen
	return nil
}

//RemoveListen 删除监听器
func (c *_ConnContext) RemoveListen(name string) {
	delete(c.listens, name)
}

//RemoteAddr 获取远程地址
func (c *_ConnContext)RemoteAddr() net.Addr{
	return c.conn.RemoteAddr()
}

//LocalAddr 获取本地地址
func (c *_ConnContext) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

//IsOpen 上下文是否打开
func (c *_ConnContext) IsOpen() bool {
	return c.isOpen
}

//SetHandler 设置Handler
func (c *_ConnContext) SetHandler(handler Handler) {
	c.handler = handler
}

//SetProtocols 设置Protocol
func (c *_ConnContext) SetProtocols(protocols ...Protocol) {
	for _, protocol := range protocols {
		encoder, decoder := newProtocolChain(protocol)
		c.encoder.addNextChain(encoder)
		c.decoder.addNextChain(decoder)
	}
}

//开始处理上下文
func (c *_ConnContext) Start(cxt context.Context) {
	c.once.Do(func() {
		c.isOpen = true
		defer func() {
			if err := recover(); err != nil {
				logger.Error("处理数据出现了异常:%s", err)
			}
		}()
		c.handler.Active(cxt, c)
		go func(connContext *_ConnContext) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("系统异常:%s", err)
				}
			}()
			for _, l := range connContext.listens {
				l.OnActive(cxt, connContext)
			}
		}(c)
		go func() {
			//TODO 此处要优化,最好使用buf池
			bytes := make([]byte, 1500)
			//TODO 加入读取或者写入超时的限制
			for c.isOpen {
				i, err := c.conn.Read(bytes)
				if err != nil {
					if err == io.EOF {
						c.Close(cxt)
						continue
					}
					if opErr, ok := err.(*net.OpError); ok {
						if !opErr.Timeout() && !opErr.Temporary() {
							logger.Error("接收到不可恢复的异常,关闭连接,%s", err)
							c.Close(cxt)
						}
					} else {
						c.FireException(cxt, fmt.Errorf("接收内容异常,%#v", err))
					}
					continue
				}
				if i > 0 {
					c.decoder.Fire(cxt, c, bytes[:i])
				}
			}
		}()
	})
}

/*
	处理异常
*/
func (c *_ConnContext) FireException(cxt context.Context, err error) {
	logger.Debug("获取了一个异常事件:%s", err)
	c.handler.Exception(cxt, c, err)
	go func(cxt context.Context, connContext ConnContext, err error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("系统异常:%s", err)
			}
		}()
		for _, l := range connContext.GetListens() {
			l.OnException(cxt, connContext, err)
		}
	}(cxt, c, err)
}

//Write 写入数据
func (c *_ConnContext) Write(cxt context.Context, data interface{}) {
	if !c.isOpen {
		logger.Warn("通道已经关闭,不能发送数据")
		return
	}
	atomic.AddInt32(&c.writing, 1)
	defer func() {
		if err := recover(); err != nil {
			logger.Error("发送数据异常,%s", err)
		}
		atomic.AddInt32(&c.writing, -1)
	}()
	c.encoder.Fire(cxt, c, data)
}

//Close 关闭上下文,包括关闭连接等
func (c *_ConnContext) Close(cxt context.Context) error {
	if c.isOpen {
		logger.Info("开始关闭连接")
		c.isOpen = false
		if c.writing != 0 {
			for i := 0; i <= 1000; i++ {
				time.Sleep(time.Millisecond * 10)
				if c.writing == 0 {
					break
				}
			}
		}
		err := c.conn.Close()
		if err != nil {
			c.FireException(cxt, err)
			return err
		}
		logger.Info("关闭了连接,%s", c.conn.RemoteAddr().String())
		c.handler.Close(cxt, c)
		c.encoder.Destroy()
		go func(this *_ConnContext) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("系统异常:%s", err)
				}
			}()
			for _, l := range this.listens {
				l.OnClose(cxt, this)
			}
		}(c)
	}
	return nil
}