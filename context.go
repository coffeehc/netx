// context
package coffeenet

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/coffeehc/logger"
)

//建立建立后的上下文
type Context struct {
	id           int64    //channel id
	conn         net.Conn // socket connection
	handler      Handler  //biz Handler
	headProtocol *ProtocolWarp
	tailProtocol *ProtocolWarp
	isOpen       bool //通道是否打开
	listens      map[string]ContextListen
	workPool     chan int64
	writing      int32 //关闭的时候用于标记剩余多少数据没有写完
	attr         map[string]interface{}
	orderHandler bool
	handlerStat  *HanderStat
}

//获取上下文对应的ID
func (this *Context) GetId() int64 {
	return this.id
}

//初始化一个新的上下文
//id:指定的上下文编号
//conn:网络连接
//workPool:用于控制并行处理工作池,主要用于标记
//orderHandler:标记该上下文处理数据的时候是否顺序处理,默认为并行,否则为当前数据处理完之后再处理接下来的数据
func newContext(id int64, conn net.Conn, workPool chan int64, orderHandler bool) *Context {
	context := new(Context)
	context.id = id
	context.conn = conn
	context.listens = make(map[string]ContextListen, 0)
	context.workPool = workPool
	context.writing = 0
	context.attr = make(map[string]interface{})
	context.orderHandler = orderHandler
	return context
}

//设置属性
func (this *Context) SetAttr(key string, value interface{}) {
	this.attr[key] = value
}

//获取属性
func (this *Context) GetAttr(key string) interface{} {
	return this.attr[key]
}

//添加监听器
func (this *Context) AddListen(name string, listen ContextListen) error {
	if _, ok := this.listens[name]; ok {
		return fmt.Errorf("监听器[%s]已经存在,不能添加", name)
	}
	this.listens[name] = listen
	return nil
}

//删除监听器
func (this *Context) RemoveListen(name string) {
	delete(this.listens, name)
}

//获取远程地址
func (this *Context) RemortAddr() net.Addr {
	return this.conn.RemoteAddr()
}

//获取本地地址
func (this *Context) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

//上下文是否打开
func (this *Context) IsOpen() bool {
	return this.isOpen
}

//设置Handler
func (this *Context) SetHandler(handler Handler) {
	this.handler = handler
}

//设置Protocol
func (this *Context) SetProtocols(protocols []Protocol) {
	var curWarp *ProtocolWarp
	for _, protocol := range protocols {
		warp := newProtocolWarp(protocol)
		if this.headProtocol == nil {
			this.headProtocol = warp
		} else {
			curWarp.next = warp
		}
		warp.prve = curWarp
		curWarp = warp
	}
	this.tailProtocol = curWarp
}

//开始处理上下文
func (this *Context) process(wait chan<- bool) {
	this.isOpen = true
	wait <- this.isOpen
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("处理数据出现了异常:%s", err)
		}
	}()
	this.handler.Active(this)
	go func(this *Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("系统异常:%s", err)
			}
		}()
		for _, l := range this.listens {
			l.OnActive(this)
		}
	}(this)
	//TODO 此处要优化,最好使用buf池
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
				if !opErr.Timeout() && !opErr.Temporary() {
					logger.Error("接收到不可恢复的异常,关闭连接,%s", err)
					this.Close()
				}
			} else {
				this.fireException(fmt.Errorf("接收内容异常,%#v", err))
			}
			continue
		}
		if i > 0 {
			this.headProtocol.decode(this, bytes[:i])
		}
	}
}

/*
	处理异常
*/
func (this *Context) fireException(err error) {
	logger.Debug("获取了一个异常事件:%s", err)
	this.handler.Exception(this, err)
	go func(this *Context, err error) {
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

//写入数据
func (this *Context) Write(data interface{}) {
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
	this.tailProtocol.encode(this, data)
}

//关闭上下文,包括关闭连接等
func (this *Context) Close() error {
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
			return err
		}
		logger.Debug("关闭了连接,%s", this.conn.RemoteAddr().String())
		this.handler.Close(this)
		this.headProtocol.Destroy()
		go func(this *Context) {
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
	return nil
}

//向对端发送数据
func (this *Context) write(data []byte) {
	_, err := this.conn.Write(data)
	if err != nil {
		if err == io.EOF {
			this.Close()
			return
		}
		if opErr, ok := err.(*net.OpError); ok {
			if !opErr.Timeout() && !opErr.Temporary() {
				logger.Error("接收到不可恢复的异常,关闭连接,%s", err)
				this.Close()
			}
		} else {
			this.fireException(fmt.Errorf("接收内容异常,%#v", err))
		}
	}
}

//处理封装好的数据
func (this *Context) handle(data interface{}) {
	this.workPool <- this.id
	if this.orderHandler {
		_hanle(this, data)
	} else {
		go _hanle(this, data)
	}
}

//处理数据
func _hanle(context *Context, data interface{}) {
	t1 := time.Now()
	defer func() {
		<-context.workPool
		context.handlerStat.acceptData(time.Since(t1))
		if err := recover(); err != nil {
			logger.Error("处理数据时出现了不可恢复的异常:%s", err)
			context.Close()
		}
	}()
	context.handler.Read(context, data)
}
