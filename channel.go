// channle.go
package coffeenet

import (
	"fmt"
	"io"
	"logger"
	"net"
	"sync/atomic"
	"time"
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
	ChannelRead(context *ChannelHandlerContext, data interface{})
	ChannelClose(context *ChannelHandlerContext)
}

type ChannelHandlerContext struct {
	id           int32
	conn         net.Conn
	handler      ChannelHandler
	headProtocol *ChannelProtocolWarp
	tailProtocol *ChannelProtocolWarp
	isOpen       bool
	listens      []ChannelListen
	remortAddr   net.Addr
	workPool     chan int
	writing      int32
	attr         map[string]interface{}
}

func (this *ChannelHandlerContext) GetId() int32 {
	return this.id
}

func NewChannelHandlerContext(id int32, conn net.Conn, workPool chan int) *ChannelHandlerContext {
	channelHandlerContext := new(ChannelHandlerContext)
	channelHandlerContext.id = id
	channelHandlerContext.conn = conn
	channelHandlerContext.listens = make([]ChannelListen, 0)
	channelHandlerContext.remortAddr = conn.RemoteAddr()
	channelHandlerContext.workPool = workPool
	channelHandlerContext.writing = 0
	channelHandlerContext.attr = make(map[string]interface{})
	return channelHandlerContext
}

func (this *ChannelHandlerContext) SetAttr(key string, value interface{}) {
	this.attr[key] = value
}

func (this *ChannelHandlerContext) GetAttr(key string) interface{} {
	return this.attr[key]
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
	}
	this.tailProtocol = curWarp
	//TODO 这里需要处理

}

func (this *ChannelHandlerContext) handle() {
	this.isOpen = true
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("处理数据出现了异常:%s", err)
		}
	}()
	this.handler.Active(this)
	go func(this *ChannelHandlerContext) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("系统异常:%s", err)
			}
		}()
		for _, l := range this.listens {
			l.OnActive(this)
		}
	}(this)
	bytes := make([]byte, 1500)
	//TODO 加入读取或者写入超时的限制
	for this.isOpen {
		i, err := this.conn.Read(bytes)
		if err != nil {
			if err == io.EOF {
				this.Close()
				continue
			}
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Timeout() || opErr.Temporary() {
					continue
				}
				// 链接重置请参考：https://github.com/cyfdecyf/cow/blob/master/proxy_unix.go
				//https://github.com/cyfdecyf/cow/blob/master/proxy_windows.go
				this.Close()
			} else {
				this.fireException(fmt.Errorf("接受内容异常,%#v", err))
			}
			continue
		}
		if i > 0 {
			this.headProtocol.read(this, bytes[:i])
		}
	}
}

func (this *ChannelHandlerContext) IsOpen() bool {
	return this.isOpen
}

/*
	处理异常
*/
func (this *ChannelHandlerContext) fireException(err error) {
	logger.Debug("获取了一个异常事件:%s", err)
	this.handler.Exception(this, err)
	go func(this *ChannelHandlerContext, err error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("系统异常:%s", err)
			}
		}()
		for _, l := range this.listens {
			l.OnException(this, err)
		}
	}(this, err)
}

func (this *ChannelHandlerContext) Write(data interface{}) {
	if !this.isOpen {
		logger.Warn("通道已经关闭,不能发送数据")
		return
	}
	atomic.AddInt32(&this.writing, 1)
	defer func() {
		if err := recover(); err != nil {
			logger.Error("发送数据异常,%s", err)
		}
		atomic.AddInt32(&this.writing, -1)
	}()
	this.tailProtocol.write(this, data)
}

func (this *ChannelHandlerContext) Close() {
	if this.isOpen {
		logger.Debug("开始关闭连接")
		this.isOpen = false
		if this.writing != 0 {
			for i := 0; i <= 1000; i++ {
				time.Sleep(time.Millisecond * 10)
				if this.writing == 0 {
					break
				}
			}
		}
		err := this.conn.Close()
		if err != nil {
			this.fireException(err)
		}
		logger.Debug("关闭了连接,%s", this.conn.RemoteAddr().String())
		this.handler.ChannelClose(this)
		this.headProtocol.Destroy()
		go func(this *ChannelHandlerContext) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("系统异常:%s", err)
				}
			}()
			for _, l := range this.listens {
				l.OnClose(this)
			}
		}(this)
	}
}

func (this *ChannelHandlerContext) write(data []byte) {
	_, err := this.conn.Write(data)
	if err != nil {
		go this.fireException(err)
	}
}
