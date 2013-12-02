package coffeenet

import (
	"fmt"
	"logger"
	"net"
)

type Client struct {
	BootStrap
	conn net.Conn
}

func NewClient(host string, netType string) *Client {
	client := new(Client)
	client.host = host
	client.netType = netType
	return client
}

func (this *Client) Connect() (*ChannelHandlerContext, error) {
	conn, err := net.Dial(this.netType, this.host)
	if err != nil {
		return nil, fmt.Errorf("connect出现错误:%s", err)
	}
	this.conn = conn
	logger.Infof("已经connect:[%s]%s->%s", this.netType, conn.LocalAddr(), conn.RemoteAddr())
	channelHandlerContext := this.channelHandlerContextFactory.CreatChannelHandlerContext(conn)
	go channelHandlerContext.handle()
	return channelHandlerContext, nil
}
