// procotol
package coffeenet

import (
	"fmt"
	"time"

	"github.com/coffeehc/logger"
)

type ChannelProtocol interface {
	Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{})
	Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{})
}

type ChannelProtocolDestroy interface {
	Destroy()
}

type ChannelProtocolWarp struct {
	protocol ChannelProtocol
	prve     *ChannelProtocolWarp
	next     *ChannelProtocolWarp
}

func newChannelProtocolWarp(protocol ChannelProtocol) *ChannelProtocolWarp {
	warp := new(ChannelProtocolWarp)
	warp.protocol = protocol
	return warp
}

func (this *ChannelProtocolWarp) read(context *ChannelHandlerContext, data interface{}) {
	this.protocol.Decode(context, this, data)
}

func (this *ChannelProtocolWarp) FireNextRead(context *ChannelHandlerContext, data interface{}) {
	if data == nil {
		return
	}
	warp := this.next
	if warp != nil {
		warp.read(context, data)
	} else {
		select {
		case context.workPool <- 0:
			go func(context *ChannelHandlerContext) {
				//TODO 这里可以加上统计数据
				defer func() {
					<-context.workPool
					if err := recover(); err != nil {
						logger.Error("处理数据时出现了不可恢复的异常:%s", err)
						context.Close()
					}
				}()
				t1 := time.Now()
				context.handler.ChannelRead(context, data)
				context.handlerStat.AcceptData(time.Since(t1))
			}(context)
		case <-time.After(time.Second * 3):
			logger.Warn("进程池无可用进程,丢弃读取的数据")
		}

	}
}

func (this *ChannelProtocolWarp) write(context *ChannelHandlerContext, data interface{}) {
	this.protocol.Encode(context, this, data)
}

func (this *ChannelProtocolWarp) FireNextWrite(context *ChannelHandlerContext, data interface{}) {
	if data == nil {
		return
	}
	warp := this.prve
	if warp != nil {
		warp.write(context, data)
	} else {
		if v, ok := data.([]byte); ok {
			context.write(v)
		} else {
			context.fireException(fmt.Errorf("发送的数据不能转换为byte数组"))
		}
	}
}

func (this *ChannelProtocolWarp) Destroy() {
	if v, ok := this.protocol.(ChannelProtocolDestroy); ok {
		v.Destroy()
	}
	if this.next != nil {
		this.next.Destroy()
	}
}

type defaultChannelProtocol struct {
}

func (this *defaultChannelProtocol) Encode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	warp.FireNextWrite(context, data)
}
func (this *defaultChannelProtocol) Decode(context *ChannelHandlerContext, warp *ChannelProtocolWarp, data interface{}) {
	warp.FireNextRead(context, data)
}
