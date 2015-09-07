package coffeenet

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	host      string
	netType   string
	context   *Context
	bootstrap Bootstrap
}

//获取该客户端的上下文
func (this *Client) GetContext() *Context {
	return this.context
}

//使用指定的方式连接指定的地址
func (this *Client) Connect(timeout time.Duration) error {
	var d net.Dialer
	if timeout != 0 {
		d = net.Dialer{Timeout: timeout}
	}
	conn, err := d.Dial(this.netType, this.host)
	if err != nil {
		return fmt.Errorf("connect出现错误:%s", err)
	}
	this.context, err = this.bootstrap.Connection(conn)
	return nil
}

func (this *Client) Close() error {
	if this.context != nil {
		return this.context.Close()
	}
	return nil
}
