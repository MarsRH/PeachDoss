package objects

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/MarsRH/PeachDoss/apiServer/heartbeat"
	"github.com/MarsRH/PeachDoss/apiServer/locate"
	"github.com/MarsRH/PeachDoss/apiServer/objectStream"
	"github.com/MarsRH/PeachDoss/sys"
)

/*接口服务的 PUT 和 GET 请求是将 HTTP 请求转发到数据服务，实际上是调用数据服务的 PUT 和 GET 方法*/
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method

	// PUT 方法时，创建或者替换资源
	if m == http.MethodPut {
		put(w, r)
		return
	}

	// GET 方法时，获取资源
	if m == http.MethodGet {
		get(w, r)
		return
	}

	// 其他方式时，返回状态码，方法不允许
	w.WriteHeader(http.StatusMethodNotAllowed)
}

/*处理接口服务 PUT 请求*/
func put(w http.ResponseWriter, r *http.Request) {
	// 获取存储对象名字
	obj := strings.Split(r.URL.EscapedPath(), "/")[2]
	// 存储请求数据体
	httpStatus, e := storeObject(r.Body, obj)
	if e != nil {
		log.Fatalln(e)
	}
	// 返回存储结果
	w.WriteHeader(httpStatus)
}

func storeObject(r io.Reader, obj string) (int, error) {
	// 获取接口服务节点存储对象的流
	stream, e := putStream(obj)
	if e != nil {
		return http.StatusServiceUnavailable, e
	}
	// 将请求数据体拷贝到流 stream
	io.Copy(stream, r)
	// 关闭流
	e = stream.Close()
	if e != nil {
		return http.StatusInternalServerError, e
	}
	// 返回成功状态码
	return http.StatusOK, nil
}

func putStream(obj string) (*objectStream.PutStream, error) {
	// 随机选择一个数据服务节点
	server := heartbeat.ChooseRandomDataServer()
	// 若没有可用的数据服务节点则返回错误信息
	if server == "" {
		return nil, fmt.Errorf(sys.DataServerNotFound)
	}
	// 返回数据服务节点存储对象的流
	return objectStream.NewPutStream(server, obj), nil
}

/*处理接口服务 GET 请求*/
func get(w http.ResponseWriter, r *http.Request) {
	// 获取存储对象名称
	obj := strings.Split(r.URL.EscapedPath(), "/")[2]
	// 获取存储对象数据流
	stream, e := getStream(obj)
	if e != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 将数据流拷贝到响应流 w
	io.Copy(w, stream)
}

func getStream(obj string) (io.Reader, error) {
	// 根据存储对象名称进行定位
	server := locate.Locate(obj)
	// 未找到该存储对象时返回定位失败错误
	if server == "" {
		return nil, fmt.Errorf(sys.DataServerLocateFail, obj)
	}
	// 定位到存储对象时，返回该对象的数据流
	return objectStream.NewGetStream(server, obj)
}
