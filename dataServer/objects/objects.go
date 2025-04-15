package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MarsRH/PeachDoss/sys"
)

/*
PUT/GET 业务处理函数
*/
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

func put(w http.ResponseWriter, r *http.Request) {

	// 处理文件路径，将文件保存在 D:/uploadFile 目录下，文件名为客户端 PUT 请求时的名字，文件名应为转义后的名字
	// 文件路径可以使用 os.Getenv(var) 通过环境变量来指定会更加灵活
	log.Println(r.URL.EscapedPath())
	fileName := os.Getenv(sys.StorageRoot) + strings.Split(r.URL.EscapedPath(), "/")[2]
	// 创建文件
	f, e := os.Create(fileName)
	if e != nil {
		log.Println(e)
		// 返回服务器错误状态码
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 将客户端 PUT 请求的文件数据保存到服务端创建的文件 f 中
	io.Copy(f, r.Body)
	// 关闭文件
	defer f.Close()
}

func get(w http.ResponseWriter, r *http.Request) {
	// 处理文件路径，从 D:/uploadFile 目录下获取文件，文件名为客户端 GET 请求时的名字，文件名应为转义后的名字
	// 文件路径可以使用 os.Getenv(var) 通过环境变量来指定会更加灵活
	log.Println(r.URL.EscapedPath())
	fileName := os.Getenv(sys.StorageRoot) + strings.Split(r.URL.EscapedPath(), "/")[2]
	// 创建文件
	f, e := os.Open(fileName)
	if e != nil {
		log.Println(e)
		// 返回文件未找到状态码
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 将服务端的文件 f 复制到响应体 w 中
	io.Copy(w, f)
	// 关闭文件
	defer f.Close()
}

/*
注意：
f 本身类型是 *os.File，它是一个指向 os.File 结构体的指针，而 os.File 结构体同时实现了 io.Writer 和 io.Reader 两个接口，
因此它既是一个 io.Writer 又是 一个 io.Reader，可以在 io.Copy 中作为不同的参数使用。
*/
