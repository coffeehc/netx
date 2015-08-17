package coffeenet

import (
	"fmt"
	"net"
	"time"

	"github.com/coffeehc/logger"
)

type Client struct {
	BootStrap
}

func NewClient(workPoolSize int) *Client {
	client := new(Client)
	client.group = make(map[int32]*ChannelHandlerContext)
	client.workConcurrent = workPoolSize
	client.init()
	return client
}

func (this *Client) Connect(netType, host string, contextFactory *ChannelHandlerContextFactory, timeout time.Duration, connSetting func(conn net.Conn)) error {
	var d net.Dialer
	if timeout != 0 {
		d = net.Dialer{Timeout: timeout}
	}
	conn, err := d.Dial(netType, host)
	if err != nil {
		return fmt.Errorf("connect出现错误:%s", err)
	}
	if connSetting != nil {
		connSetting(conn)
	}
	logger.Info("已经connect:[%s]%s->%s", netType, conn.LocalAddr(), conn.RemoteAddr())
	contextFactory.bootStrap = &this.BootStrap
	channelHandlerContext := contextFactory.CreatChannelHandlerContext(conn, this.workPool)
	go channelHandlerContext.handle()
	return nil
}
