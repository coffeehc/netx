# coffeenet --简单的一个Net框架

***

## 需要实现的接口

**协议接口**


`	type ChannelProtocol interface {`
`		Encode(context *ChannelHandlerContext, warp *ChannekProtocolWarp, data interface{})`
`		Decode(context *ChannelHandlerContext, warp *ChannekProtocolWarp, data interface{})`
`	}`


**处理接口**

`
	type ChannelHandler interface {
		Active(context *ChannelHandlerContext)
		Exception(context *ChannelHandlerContext, err error)
		ChannelRead(context *ChannelHandlerContext, data interface{}) error
		ChannelClose(context *ChannelHandlerContext)
	}
`
## 例子:

`
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
		time.Sleep(time.Second * 200)
	}
`

具体使用可以参考server_test.go