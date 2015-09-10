# coffeenet --简单的一个Net框架
[![GoDoc](https://godoc.org/github.com/coffeehc/coffeenet?status.png)](http://godoc.org/github.com/coffeehc/coffeenet)

## 2.0.0 架构
一个全新版本,将代码进行了大的整理,每个功能都单独定义到具体的go源文件中,并且重新定义了接口以及架构

```bash
                    --------       --------
                   | server |     | client |
                    --------       --------
                        |             |
                         -------------
                               |
                          -----------
                         | bootstrap |
                          -----------
                               |
                          -----------
                         | protocols |    
                          -----------
                               |
                           ---------
                          | handler | 
                           --------- 
```

## Bootstrap

 > 接口定义
 
 ```go
    type Bootstrap interface {
	   //创建一个新的Server
	   NewServer(netType, host string) *Server
	   //创建一个信的Client
	   NewClient(netType, host string) *Client
	   //当连接创立的时候需要被调用,可用于自定义扩展
	   Connection(conn net.Conn) (*Context, error)
	   //关闭多有的链接
	   Close() error
    }
 ```
 
 > 创建Bootstrap
 
```go
func NewBootStrap(config *Config, contextFactory *ContextFactory, connectionSetting func(conn net.Conn)) Bootstrap
```


其中Config主要用于设置Bootstrap连接数限制以及是否并发处理,将来可能还需要扩展

```go
 type Config struct {
	//最大连接数
	MaxConnection int
	//最大并发处理个数
	MaxConcurrentHandler int
	//是否顺序处理消息,默认false,即可以并发处理消息
	OrderHandler bool
}
```

``func connectionSetting(conn net.Conn)``方法主要是在创建连接的时候对net.Conn的属性进行自定义的扩展,没有默认方法,将来优化后可能会有默认方法

contextFactory的创建需要使用

```go
func NewContextFactory(initContextFunc func(context *Context)) *ContextFactory
```
在建立连接后创建Context的时候对Context进行初始化设置,如设置protocol,handler等

> Bootstrap关闭

调用Close()方法,将会关闭由Bootstrap管理的所有Connection.

## Server&Client
> 创建Server

调用``Bootstrap.NewServer()``来创建,==目前仅支持TCP==

>Server的启动
    
调用Server.Bind()来启动监听.如果想要监听多个端口,请创建多了Server,可以共用一个Bootstrap来管理链接

>创建Client

调用``Bootstrap.NewClient()``来创建,==目前仅支持TCP==

>Client启动

调用``Client.Client(timeout time.Duration)``

## 例子

```go
func TestServer(t *testing.T) {
	contextFactory := NewContextFactory(func(context *Context) {
		//设置
		context.SetProtocols([]Protocol{new(defaultProtocol)})
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
```


