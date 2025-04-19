package version

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/MarsRH/PeachDoss/es"
)

/*处理版本搜索*/
func Handler(w http.ResponseWriter, r *http.Request) {
	// 非 GET 方法时响应方法不允许
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 其实是分页参数，一页最多有 1000 条记录，默认从第 0 条开始往后取数据
	// 当返回值的长度不等于 1000 时，则说明后续没有数据了，直接返回
	// 当返回值等于 1000 时，说明后续可能有数据， from 则从 1000 条开始往后取数据
	from := 0
	size := 1000
	// 若未指定名字，则切割 URL 之后名字为空字符串
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	for {
		metas, e := es.SearchAllVersions(name, from, size)
		if e != nil {
			log.Println(e)
			// 服务器内部错误
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 遍历结果集
		for i := range metas {
			// 格式化为 json 返回
			b, _ := json.Marshal(metas[i])
			w.Write(b)
			w.Write([]byte("\n"))
		}

		if len(metas) != size {
			return
		}
		from += size
	}
}
