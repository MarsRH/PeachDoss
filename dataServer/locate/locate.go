package locate

import (
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/MarsRH/PeachDoss/rabbitmq"
	"github.com/MarsRH/PeachDoss/sys"
)

// 定位对象
func Locate(name string) bool {
	// 访问磁盘上对应的文件名
	_, e := os.Stat(name)
	// 判读文件名是否存在
	return !os.IsNotExist(e)
}

// 监听定位信息
func StartLocate() {
	q := rabbitmq.New(os.Getenv(sys.RabbitmqServer))
	defer q.Close()
	// 绑定 data 网络层
	q.Bind(sys.DataServersExchange)
	// 获取信息管道
	c := q.Consume()
	// 从管道中遍历信息，msg 为需要定位的存储对象名字
	for msg := range c {
		// 去掉 json 序列化的双引号
		obj, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			log.Fatalln(e)
		}
		// 存储根目录拼接文件名，定位存储对象，名字需要 URL 转义
		if Locate(os.Getenv(sys.StorageRoot) + url.PathEscape(obj)) {
			// 如果存储对象存在，则回送本节点监听地址，已告知存储对象在该节点
			q.Send(msg.ReplyTo, os.Getenv(sys.ListenAddress))
		}
	}
}
