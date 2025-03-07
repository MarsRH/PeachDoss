package main

import (
	"net/http"
	"os"

	"github.com/MarsRH/PeachDoss/dataServer/heartbeat"
	"github.com/MarsRH/PeachDoss/dataServer/locate"
	"github.com/MarsRH/PeachDoss/dataServer/objects"
	"github.com/MarsRH/PeachDoss/sys"
)

func main() {
	// 向 apiServers exchange 发送心跳
	go heartbeat.StartHeartbeat()
	// 监听定位信息
	go locate.StartLocate()
	// 注册URL与逻辑处理函数
	http.HandleFunc("/handleObjs/", objects.Handler)
	// 启动并监听服务
	http.ListenAndServe(os.Getenv(sys.ListenAddress), nil)
}
