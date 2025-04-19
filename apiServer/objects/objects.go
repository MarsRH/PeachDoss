package objects

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"strconv"

	"github.com/MarsRH/PeachDoss/apiServer/heartbeat"
	"github.com/MarsRH/PeachDoss/apiServer/locate"
	"github.com/MarsRH/PeachDoss/es"
	"github.com/MarsRH/PeachDoss/objectStream"
	"github.com/MarsRH/PeachDoss/sys"
	"github.com/MarsRH/PeachDoss/utils"
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

	// 版本删除
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	// 其他方式时，返回状态码，方法不允许
	w.WriteHeader(http.StatusMethodNotAllowed)
}

/*处理接口服务 PUT 请求*/
func put(w http.ResponseWriter, r *http.Request) {
	// 按以前的步骤，这里应该获取存储对象名字，不过从 header 中取对象的散列值作为名字
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println(sys.MissingObjectHash)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 存储请求数据，散列值要作转义
	httpStatus, e := storeObject(r.Body, url.PathEscape(hash))
	if e != nil {
		log.Println(e)
		w.WriteHeader(httpStatus)
		return
	}
	if httpStatus != http.StatusOK {
		w.WriteHeader(httpStatus)
		return
	}

	// 获取名字和大小，新增一个对象版本
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	size := utils.GetSizeFromHeader(r.Header)
	e = es.AddVersion(name, hash, size)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// 返回结果
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
	// 获取存储对象名称和版本号
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	versionId := r.URL.Query()["version"]
	version := 0
	var e error
	if len(versionId) != 0 {
		// 版本号字符串转数字
		version, e = strconv.Atoi(versionId[0])
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	// 根据名字和版本号来获取元数据
	meta, e := es.GetMetadata(name, version)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 元数据散列值为空则无该对象
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 散列值要作 URL 转移
	object := url.PathEscape(meta.Hash)
	// 根据散列值获取对象数据
	stream, e := getStream(object)
	if e != nil {
		log.Println(e)
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

/*处理接口服务 DELETE 请求*/
func del(w http.ResponseWriter, r *http.Request) {
	// 获取名字
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	v, e := es.SearchLatestVersion(name)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 插入一条新的元数据作删除标记
	e = es.PutMetadata(name, v.Version+1, 0, "")
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
