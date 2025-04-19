package es

/* 该 ES 包封装了以 HTTP 访问 ES 的各种 API 的操作 */
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/MarsRH/PeachDoss/sys"
)

/* 元数据结构体 */
type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
}

type hit struct {
	Source Metadata `json:"_source"`
}

type searchResult struct {
	Hits struct {
		Total int
		Hits  []hit
	}
}

/*根据对象的名称和版本号来获取元数据*/
func getMetadata(name string, versionId int) (meta Metadata, e error) {
	// 索引为 metadata ，类型为 objects，文档 id 为对象名称和版本号的拼接
	url := fmt.Sprintf(sys.GetMetadataUrl, os.Getenv(sys.EsServer), name, versionId)
	// 通过 GET URL 可以直接获取该对象的元数据，免除了耗时的搜索操作
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf(sys.FailToGetMetadata, name, versionId, r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	// 将请求结果反序列化为元数据结构
	json.Unmarshal(result, &meta)
	return
}

/*根据对象名称获取最新版本的元数据*/
func SearchLatestVersion(name string) (meta Metadata, e error) {
	// 构建 url 时需要将名称转移成 url 字符
	url := fmt.Sprintf(sys.SearchLatestVersionUrl, os.Getenv(sys.EsServer), url.PathEscape(name))
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf(sys.FailToSearchLatestMetadata, r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	// 请求结果反序列化
	json.Unmarshal(result, &sr)
	// 如果长度为 0 则没有搜索结果，直接返回
	if len(sr.Hits.Hits) != 0 {
		meta = sr.Hits.Hits[0].Source
	}
	return
}

/*根据对象的名称和版本号来获取元数据*/
func GetMetadata(name string, version int) (Metadata, error) {
	// 没有指定版本号时默认返回最新版本的元数据
	if version == 0 {
		return SearchLatestVersion(name)
	}
	return getMetadata(name, version)
}

/*向 ES 服务上传一个新的元数据*/
func PutMetadata(name string, version int, size int64, hash string) error {
	doc := fmt.Sprintf(sys.MetadataJson, name, version, size, hash)
	client := http.Client{}
	url := fmt.Sprintf(sys.PutMetadataUrl, os.Getenv(sys.EsServer), name, version)
	request, _ := http.NewRequest(http.MethodPut, url, strings.NewReader(doc))
	// 加入 header ，否则报 406
	request.Header.Add("content-type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		return e
	}
	if r.StatusCode == http.StatusConflict {
		return PutMetadata(name, version+1, size, hash)
	}
	if r.StatusCode != http.StatusCreated {
		result, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf(sys.FailToPutMetadata, r.StatusCode, string(result))
	}
	return nil
}

/*版本号加一*/
func AddVersion(name, hash string, size int64) error {
	// 获取目前最新的版本
	version, e := SearchLatestVersion(name)
	if e != nil {
		return e
	}
	// 创建一个最新的版本号
	return PutMetadata(name, version.Version+1, size, hash)
}

/*搜索对象的全部版本*/
func SearchAllVersions(name string, from, size int) ([]Metadata, error) {
	// 不指定名字时则搜索全部对象的全部版本，指定名字时则搜索某个对象的全部版本
	url := fmt.Sprintf(sys.SearchAllVersionsUrl, os.Getenv(sys.EsServer), from, size)
	if name != "" {
		url += "&q=name:" + name
	}
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	metas := make([]Metadata, 0)
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	for i := range sr.Hits.Hits {
		metas = append(metas, sr.Hits.Hits[i].Source)
	}
	return metas, nil
}

/*删除指定的版本*/
func DelMetadata(name string, version int) {
	client := http.Client{}
	url := fmt.Sprintf(sys.DelMetadataUrl, os.Getenv(sys.EsServer), name, version)
	request, _ := http.NewRequest(http.MethodDelete, url, nil)
	client.Do(request)
}

type Bucket struct {
	Key         string
	Doc_count   int
	Min_version struct {
		Value float32
	}
}

type aggregateResult struct {
	Aggregations struct {
		Group_by_name struct {
			Buckets []Bucket
		}
	}
}

/*搜索版本状态*/
func SearchVersionStatus(min_doc_count int) ([]Bucket, error) {
	client := http.Client{}
	url := fmt.Sprintf(sys.SearchVersionStatusUrl, os.Getenv(sys.EsServer))
	body := fmt.Sprintf(sys.SearchVersionStatusJson, min_doc_count)
	request, _ := http.NewRequest(http.MethodGet, url, strings.NewReader(body))
	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var ar aggregateResult
	json.Unmarshal(b, &ar)
	return ar.Aggregations.Group_by_name.Buckets, nil
}

func HasHash(hash string) (bool, error) {
	url := fmt.Sprintf(sys.HasHashUrl, os.Getenv(sys.EsServer), hash)
	r, e := http.Get(url)
	if e != nil {
		return false, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(b, &sr)
	return sr.Hits.Total != 0, nil
}

func SearchHashSize(hash string) (size int64, e error) {
	url := fmt.Sprintf(sys.SearchHashSizeUrl, os.Getenv(sys.EsServer), hash)
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf(sys.FailToSearchHashSize, r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		size = sr.Hits.Hits[0].Source.Size
	}
	return
}
