package netx_test

import (
	"fmt"
	"testing"
	"time"

	"context"

	"github.com/coffeehc/logger"
	"github.com/coffeehc/netx"
)

type testHandler struct {
}

func (this *testHandler) Active(cxt context.Context, connContext netx.ConnContext) {
	logger.Debug("已经激活了一个连接")
	//context.Write([]byte("欢迎您的到来\n"))
}
func (this *testHandler) Exception(cxt context.Context, connContext netx.ConnContext, err error) {
	logger.Error("接收到一个异常:%s", err)
}
func (this *testHandler) Read(cxt context.Context, connContext netx.ConnContext, data interface{}) {
	//	logger.Debug("接收到的消息是:%s", data)
	if fmt.Sprintf("%s", data) == "next" {
		msg := fmt.Sprintf("现在时间:\n%s\nnext\n", time.Now().Format(logger.LoggerTimeformatNanosecond))
		//logger.Debug("发送消息")
		connContext.Write(cxt, []byte(msg))
	}
}
func (this *testHandler) Close(cxt context.Context, connContext netx.ConnContext) {
	logger.Debug("连接关闭掉了")
}

func TestServer(t *testing.T) {
	initContextFactoryFunc := func(cxt context.Context,connContext netx.ConnContext) {
		connContext.SetHandler(new(testHandler))
	}
	bootstrap := netx.NewBootStrap(new(netx.Config), initContextFactoryFunc, nil)
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
	connContext := client.GetConnContext()
	for i := 0; i < 10; i++ {
		connContext.Write(context.Background(), []byte(fmt.Sprintln("开始了\nnext")))
	}
	time.Sleep(time.Millisecond * 300)
	connContext.Close(context.Background())
	server.Close()
	bootstrap.Close()
}
