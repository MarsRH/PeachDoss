package heartbeat

import (
	"os"
	"time"

	"github.com/MarsRH/PeachDoss/rabbitmq"
	"github.com/MarsRH/PeachDoss/sys"
)

/*每 3s 向 apiServers exchange 发送一次心跳，心跳信息为该节点的监听地址*/
func StartHeartbeat() {
	q := rabbitmq.New(os.Getenv(sys.RabbitmqServer))
	defer q.Close()
	for {
		q.Publish(sys.ApiServersExchange, os.Getenv(sys.ListenAddress))
		// 休眠 3s
		time.Sleep(3 * time.Second)
	}
}
