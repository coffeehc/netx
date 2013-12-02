package coffeenet

import (
	"fmt"
	"logger"
	//"runtime"
	"testing"
	"time"
)

type NetMessage struct {
	Code int32  `json:"code"`
	Data string `json:"data"`
}

type testHandler struct {
	ChannelHandler
}

func (this *testHandler) Active(context *ChannelHandlerContext) {
	logger.Debugf("已经激活了一个连接")
	//context.Write([]byte("欢迎您的到来\n"))
}
func (this *testHandler) Exception(context *ChannelHandlerContext, err error) {
	logger.Errorf("接收到一个异常:%s", err)
}
func (this *testHandler) ChannelRead(context *ChannelHandlerContext, data interface{}) error {
	logger.Debugf("接收到的消息是:%s", data)
	if fmt.Sprintf("%s", data) == "next" {
		msg := fmt.Sprintf("现在时间:\n%s\nnext\n", time.Now().Format(logger.LOGGER_TIMEFORMAT_NANOSECOND))
		//logger.Debug("发送消息")
		context.Write([]byte(msg))
	}
	return nil
}
func (this *testHandler) ChannelClose(context *ChannelHandlerContext) {
	logger.Debug("连接关闭掉了")
}

func TestServer(t *testing.T) {
	server := NewServer("127.0.0.1:800", "tcp")
	channelHandlerContextFactory := NewChannelHandlerContextFactory(func(context *ChannelHandlerContext) {
		context.SetProtocols([]ChannelProtocol{NewLengthFieldProtocol(4), NewTerminalProtocol()})
		context.SetHandler(new(testHandler))
	})
	server.SetChannelHandlerContextFactory(channelHandlerContextFactory)
	err := server.Bind()
	if err != nil {
		t.Fatalf("启动服务器出现错误:%s", err)
	}
	client := NewClient("127.0.0.1:800", "tcp")
	client.SetChannelHandlerContextFactory(channelHandlerContextFactory)
	for i := 0; i < 1; i++ {
		context, err := client.Connect()
		if err != nil {
			t.Fatalf("连接服务器出现错误:%s", err)
		}
		context.Write([]byte(fmt.Sprintln("开始了\nnext")))
	}
	//for {
	//	logger.Debugf("当前运行的goroutune数量为:%d", runtime.NumGoroutine())
	//	memStat := new(runtime.MemStats)
	//	runtime.ReadMemStats(memStat)
	//	logger.Debugf("当前运行的GC次数:%d,总内存:%d", memStat.NumGC, memStat.Alloc)
	//	time.Sleep(500 * time.Millisecond)
	//}
	time.Sleep(time.Second * 200)

}
