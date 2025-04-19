package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/MarsRH/PeachDoss/apiServer/heartbeat"
	"github.com/MarsRH/PeachDoss/apiServer/locate"
	"github.com/MarsRH/PeachDoss/apiServer/objects"
	"github.com/MarsRH/PeachDoss/apiServer/version"
	"github.com/MarsRH/PeachDoss/sys"
)

func main() {
	err := godotenv.Load(".apiServer.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 监听数据服务节点心跳
	go heartbeat.ListenHeartbeat()
	// 处理对象请求，实际上是将对象请求转发给数据服务
	http.HandleFunc("/handleObjs/", objects.Handler)
	// 处理定位请求
	http.HandleFunc("/locateObj/", locate.Handler)
	// 处理版本信息
	http.HandleFunc("/versions/", version.Handler)
	// 启动并监听服务
	http.ListenAndServe(os.Getenv(sys.ListenAddress), nil)
}
