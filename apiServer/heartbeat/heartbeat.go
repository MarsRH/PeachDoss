package heartbeat

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/MarsRH/PeachDoss/rabbitmq"
	"github.com/MarsRH/PeachDoss/sys"
)

var dataServers = make(map[string]time.Time) // 缓存数据服务节点
var m sync.RWMutex                           // 使用多读写锁，比互斥锁高效

/*监听心跳*/
func ListenHeartbeat() {
	q := rabbitmq.New(os.Getenv(sys.RabbitmqServer))
	defer q.Close()
	// 绑定 api 网络层
	q.Bind(sys.ApiServersExchange)
	c := q.Consume()
	// 移除过期节点
	go removeExpiredDataServer()
	// 监听数据服务节点心跳，将心跳信息写入全局缓存
	for msg := range c {
		dataServer, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			log.Fatalln(e)
		}
		// 写操作互斥，防止多 goroutine 对 dataServers 同时写
		m.Lock()
		dataServers[dataServer] = time.Now()
		m.Unlock()
	}
}

/*移除过期的数据服务节点*/
func removeExpiredDataServer() {
	// 每 3s 扫描一遍缓存的数据服务节点
	// 若当前时间减去心跳时间超过 6s 则判定为节点过期
	for {
		time.Sleep(3 * time.Second)
		// 写操作互斥，防止多 goroutine 对 dataServers 同时写
		m.Lock()
		for s, t := range dataServers {
			if t.Add(6 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		m.Unlock()
	}
}

/*获取全部数据服务节点*/
func GetDataServers() []string {
	// 读锁，可多 goroutine 对 dataServers 同时读
	m.RLock()
	defer m.RUnlock()
	dataServer := make([]string, 0)
	for s, _ := range dataServers {
		dataServer = append(dataServer, s)
	}
	return dataServer
}

/*随机选择一个数据服务节点返回*/
func ChooseRandomDataServer() string {
	dataServer := GetDataServers()
	length := len(dataServer)
	if length == 0 {
		return ""
	}
	return dataServer[rand.Intn(length)]
}
