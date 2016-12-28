package netx

import (
	"context"
	"fmt"
	"net"
	"time"
)

//Client  net client interface
type Client interface {
	GetConnContext() ConnContext
	Connect(timeout time.Duration) error
	Close() error
}

//Client  net Client
type _Client struct {
	host        string
	netType     string
	connContext ConnContext
	bootstrap   *_bootstrap
}

//GetConnContext 获取该客户端的上下文
func (client *_Client) GetConnContext() ConnContext {
	return client.connContext
}

//Connect 使用指定的方式连接指定的地址
func (client *_Client) Connect(timeout time.Duration) error {
	var d net.Dialer
	if timeout != 0 {
		d = net.Dialer{Timeout: timeout}
	}
	conn, err := d.Dial(client.netType, client.host)
	if err != nil {
		return fmt.Errorf("connect出现错误:%s", err)
	}
	//TODO Timeout set
	client.connContext, err = client.bootstrap.connection(conn)
	return err
}

//Close 关闭 Client
func (client *_Client) Close() error {
	if client.connContext != nil {
		timeoutCxt, cancleFunc := context.WithTimeout(context.Background(), time.Second*30)
		defer cancleFunc()
		return client.connContext.Close(timeoutCxt)
	}
	return nil
}
