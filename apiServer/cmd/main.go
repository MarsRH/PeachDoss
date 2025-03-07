package main

import (
	"net/http"
	"os"

	"github.com/MarsRH/PeachDoss/apiServer/heartbeat"
	"github.com/MarsRH/PeachDoss/apiServer/locate"
	"github.com/MarsRH/PeachDoss/apiServer/objects"
	"github.com/MarsRH/PeachDoss/sys"
)

func main() {
	// 监听数据服务节点心跳
	go heartbeat.ListenHeartbeat()
	// 处理对象请求，实际上是将对象请求转发给数据服务
	http.HandleFunc("/handleObjs/", objects.Handler)
	// 处理定位请求
	http.HandleFunc("/locateObj/", locate.Handler)
	// 启动并监听服务
	http.ListenAndServe(os.Getenv(sys.ListenAddress), nil)
}
