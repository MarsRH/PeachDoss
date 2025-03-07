package locate

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MarsRH/PeachDoss/rabbitmq"
	"github.com/MarsRH/PeachDoss/sys"
)

/*定位存储对象*/
func Locate(name string) string {
	// 创建临时消息队列
	q := rabbitmq.New(os.Getenv(sys.RabbitmqServer))
	// 向 data 网络层群发这个存储对象的名字
	q.Publish(sys.DataServersExchange, name)
	// 获取信息管道
	c := q.Consume()
	// 休眠一秒之后将临时消息队列关闭，防止超时阻塞
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()
	// 从管道中读取定位信息
	msg := <-c
	s, _ := strconv.Unquote(string(msg.Body))
	return s
}

/*判断存储对象是否存在*/
func Exist(name string) bool {
	return Locate(name) != ""
}

/*处理定位请求*/
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	// 非 GET 方法
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	objInfo := Locate(strings.Split(r.URL.EscapedPath(), "/")[2])
	// 未找到该存储对象
	if len(objInfo) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	obj, _ := json.Marshal(objInfo)
	w.Write(obj)
}
