package objectStream

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MarsRH/PeachDoss/sys"
)

type PutStream struct {
	writer *io.PipeWriter // writer 用于实现 Write 方法
	c      chan error     // c 用于将 goroutine 传输数据过程中抛出的错误传回主协程
}

/* 构建接口服务存储对象流，当数据通过写入到该 PutStream.writer 时，向数据服务节点请求的 reader 就会读到管道中的数据 */
func NewPutStream(server, obj string) *PutStream {
	// 创建一对管道互连的 reader 和 writer，写入 writer 中的数据可以从 reader 中读出来
	reader, writer := io.Pipe()
	c := make(chan error)

	// 由于管道读写是阻塞的，因此需要另起一个 goroutine 来向数据服务节点发起 PUT 请求，此时接口服务是一个 client 端
	go func() {
		// 构建请求
		request, _ := http.NewRequest(http.MethodPut, "http://"+server+"/handleObjs/"+obj, reader)
		client := http.Client{}
		// 发起请求
		r, e := client.Do(request)
		// 其他异常
		if e == nil && r.StatusCode != http.StatusOK {
			e = fmt.Errorf(sys.DataServerError, r.StatusCode)
		}
		// 将异常写入管道
		c <- e
	}()

	// 返回接口服务存储对象流
	return &PutStream{
		writer: writer,
		c:      c,
	}
}

/*实现 io.Writer 接口*/
func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

/*让管道另一端的 reader 读到 io.EOF ，否则 client.Do(request) 将一直处于请求阻塞状态*/
func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}

type GetStream struct {
	reader io.Reader
}

func newGetStream(url string) (*GetStream, error) {
	// 向数据服务节点发起 GET 请求
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(sys.DataServerError, r.StatusCode)
	}
	// 返回请求响应数据
	return &GetStream{
		reader: r.Body,
	}, nil
}

func NewGetStream(server, obj string) (*GetStream, error) {
	// 当数据服务节点为空或者请求对象名字为空，则抛出异常
	if server == "" || obj == "" {
		return nil, fmt.Errorf(sys.InvalidServerOrObject, server, obj)
	}
	// 调用数据服务节点 GET 接口，返回请求响应数据
	return newGetStream("http://" + server + "/handleObjs/" + obj)
}

/*实现 io.Reader 接口*/
func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
