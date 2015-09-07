package coffeenet

import (
	"fmt"
	"testing"
	"time"

	"github.com/coffeehc/logger"
)

type testHandler struct {
}

func (this *testHandler) Active(context *Context) {
	logger.Debug("已经激活了一个连接")
	//context.Write([]byte("欢迎您的到来\n"))
}
func (this *testHandler) Exception(context *Context, err error) {
	logger.Error("接收到一个异常:%s", err)
}
func (this *testHandler) Read(context *Context, data interface{}) {
	//	logger.Debug("接收到的消息是:%s", data)
	if fmt.Sprintf("%s", data) == "next" {
		msg := fmt.Sprintf("现在时间:\n%s\nnext\n", time.Now().Format(logger.LOGGER_TIMEFORMAT_NANOSECOND))
		//logger.Debug("发送消息")
		context.Write([]byte(msg))
	}
}
func (this *testHandler) Close(context *Context) {
	logger.Debug("连接关闭掉了")
}

func TestServer(t *testing.T) {
	contextFactory := NewContextFactory(func(context *Context) {
		context.SetHandler(new(testHandler))
	})
	bootstrap := NewBootStrap(new(Config), contextFactory, nil)
	server := bootstrap.NewServer("tcp", "127.0.0.1:9991")
	err := server.Bind()
	if err != nil {
		t.Fatalf("启动服务器出现错误:%s", err)
	}
	client := bootstrap.NewClient("tcp", "127.0.0.1:9991")
	err = client.Connect(3 * time.Second)
	if err != nil {
		t.Fatalf("连接服务器出现错误:%s", err)
	}
	context := client.GetContext()
	for i := 0; i < 10; i++ {
		context.Write([]byte(fmt.Sprintln("开始了\nnext")))
	}
	time.Sleep(time.Millisecond * 300)
	context.Close()
	server.Close()
	bootstrap.Close()
}
