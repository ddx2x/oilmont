package uri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"unsafe"
)

const (
	SEPARATE   = "/"
	NAMESPACES = "namespaces"
	PODS       = "pods"
	CONTAINERS = "containers"
)

type Spec struct {
	//// multiple config microservices fields
	//Cluster string `json:"config"`
	// service config microservices fields
	Service string `json:"service"`

	// resource fields
	Api       string `json:"api"`
	Group     string `json:"group"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
	Resource  string `json:"resource"`
	Name      string `json:"name"`
	// custom op code
	Op string `json:"op"`
}

func (u *Spec) String() string {
	b, _ := json.Marshal(u)
	return hackString(b)
}

// URI is only via this project url spec.
type URI struct {
	index  uint
	count  int
	offset int64
	url    string
	method string
	// watch url need resource version
	ResourceVersion string `json:"resource_version"`
	// Spec base specified
	Cluster string
	Tenant  string
	Spec
}

func (u *URI) Source() string { return u.url }

// Parser URI general interface analysis
type Parser interface {
	ParseOp(method, url string, operationMap map[string]string) (*URI, error)
	ParseWatch(url *url.URL) ([]*URI, error)
}

// NewURIParser return default parser
func NewURIParser() Parser { return &parser{} }

var _ Parser = (*parser)(nil)

type parser struct{}

type Content struct {
	ResourceVersion string
	Cluster         string
	Tenant          string // every tenant a schema
}

func parseWatchPathAndResourceVersion(inputUrl *url.URL) (map[string]Content, error) {
	// get api=? query
	rawQuerys, err := url.ParseQuery(inputUrl.RawQuery)
	if err != nil {
		return nil, err
	}

	apis, exist := rawQuerys["api"]
	if !exist {
		return nil, fmt.Errorf("not found api query")
	}

	result := make(map[string]Content)
	for _, api := range apis {
		apiURL, err := url.Parse(api)
		if err != nil {
			return nil, err
		}
		queryData, err := url.ParseQuery(apiURL.RawQuery)
		if err != nil {
			return nil, err
		}
		result[apiURL.Path] = Content{
			ResourceVersion: queryData.Get("resourceVersion"),
			Cluster:         queryData.Get("cluster"),
			Tenant:          queryData.Get("tenant"),
		}
	}

	return result, nil
}

func (p *parser) ParseWatch(url *url.URL) ([]*URI, error) {
	uris := make([]*URI, 0)
	contentMap, err := parseWatchPathAndResourceVersion(url)
	if err != nil {
		return nil, err
	}
	for uRL, content := range contentMap {
		uri, err := parseWatch(uRL)
		if err != nil {
			return nil, err
		}
		uri.ResourceVersion = content.ResourceVersion
		uri.Cluster = content.Cluster
		uri.Tenant = content.Tenant
		uri.Spec.Op = "watch"
		uris = append(uris, uri)
	}

	return uris, nil
}

func parseWatch(inputURL string) (*URI, error) {
	parsedWatchURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, err
	}
	uri := &URI{
		url:   parsedWatchURL.Path,
		count: strings.Count(parsedWatchURL.Path, SEPARATE),
	}
	if err := uri.parseWatch(); err != nil {
		return nil, err
	}
	return uri, nil
}

func (u *URI) parseWatch() error {
	for index := 1; index <= u.count; index++ {
		item, err := u.shift()
		if err != nil {
			return err
		}
		if index == 1 {
			u.Api = item
			continue
		}
		switch u.Api {
		case "api":
			switch index {
			case 2:
				u.Version = item
				continue
			case 3:
				u.Resource = item
			case 4:
				if u.Resource == "namespaces" {
					u.Resource = item
				}
				u.Namespace = item
				continue
			case 5:
				u.Resource = item
				if u.Resource != "namespaces" {
					u.Resource = ""
				}
				continue
			case 6:
				u.Resource = item
				continue
			}

		case "apis":
			switch index {
			case 2:
				u.Group = item
				continue
			case 3:
				u.Version = item
				continue
			case 4:
				if item == "namespaces" {
					continue
				}
				u.Resource = item
				continue
			case 5:
				u.Namespace = item
				continue
			case 6:
				u.Resource = item
				continue
			case 7:
				u.Name = item
				continue
			}
		}
	}

	return nil
}

func (p *parser) ParseOp(method, url string, operationMap map[string]string) (*URI, error) {
	return parse(method, url, operationMap)
}

func parse(_method, _url string, operationMap map[string]string) (*URI, error) {
	_URL, err := url.Parse(_url)
	if err != nil {
		return nil, err
	}

	uri := &URI{
		method: _method,
		url:    _URL.Path,
		count:  strings.Count(_URL.Path, SEPARATE),
	}

	if err := uri.parse(); err != nil {
		return nil, err
	}

	if uri.Op == "" {
		uri.Op = operationMap[_method]
		/*
			switch method {
			case http.MethodGet:
				uri.Op = "view"
			case http.MethodPost:
				if uri.Resource == "metrics" {
					uri.Op = "metrics"
					uri.Namespace = _URL.Query().GetGVR("kubernetes_namespace")
				} else {
					uri.Op = "apply"
				}
			case http.MethodPut:
				uri.Op = "apply"
			case http.MethodDelete:
				uri.Op = "delete"
			}

		*/

	}

	return uri, nil
}

func (u *URI) parse() error {
	lastOp := false
	for index := 1; index <= u.count; index++ {
		item, err := u.shift()
		if err != nil {
			return err
		}
		switch index {
		case 1:
			u.Spec.Service = item
			continue

		case 2:
			switch item {
			case "metrics":
				u.Spec.Resource = "metrics"
				continue
			case "attach":
				u.Spec.Resource = "pods"
				u.Spec.Op = "attach"
				continue
			default:
				u.Api = item
			}
		}

		if item == "op" {
			lastOp = true
			continue
		}
		if lastOp {
			u.Op = item
			continue
		}

		if u.Spec.Op == "attach" && index == 4 {
			u.Spec.Namespace = item
		} else if u.Op == "attach" && index == 6 {
			u.Spec.Name = item
		}

		switch u.Spec.Api {
		case "api":
			switch index {
			case 3:
				if item == "metrics" {
					u.Spec.Resource = item
					continue
				}
				u.Spec.Version = item
				continue
			case 4:
				u.Spec.Resource = item
			case 5:
				if u.Spec.Resource == "namespaces" {
					u.Spec.Name = item
				}
				u.Spec.Namespace = item
				continue
			case 6:
				u.Spec.Resource = item
				if u.Resource != "namespaces" {
					u.Spec.Name = ""
				}
				continue
			case 7:
				u.Spec.Name = item
				continue
			case 8:
				u.Spec.Op = item
				continue
			}

		case "apis":
			switch index {
			case 3:
				u.Spec.Group = item
				continue
			case 4:
				u.Spec.Version = item
				continue
			case 5:
				if item == "namespaces" {
					continue
				}
				u.Spec.Resource = item
				continue
			case 6:
				u.Spec.Namespace = item
				continue
			case 7:
				u.Spec.Resource = item
				continue
			case 8:
				u.Spec.Name = item
				continue
			case 9:
				u.Spec.Op = item
				continue
			}
		}
	}
	return nil
}

func (u *URI) shift() (item string, err error) {
	itemBytes, err := u.shiftItem()
	if err != nil {
		return "", err
	}
	item = hackString(itemBytes)
	return
}

func (u *URI) shiftItem() (item []byte, err error) {
	item, u.offset, err = readItem(u.url, int64(u.offset))
	if err != nil {
		return nil, err
	}
	return
}

func readItem(uri string, offset int64) (item []byte, nextOffset int64, err error) {
	bytesReader := bytes.NewReader(hackSlice(uri))
	if nextOffset, err = bytesReader.Seek(offset, io.SeekCurrent); err != nil {
		return
	}
	prefix := false
	for {
		b, err := bytesReader.ReadByte()
		if err != nil {
			if err == io.EOF {
				return item, nextOffset, nil
			}
			return nil, nextOffset, err
		}
		if b == '/' && !prefix {
			prefix = true
			nextOffset++
			continue
		}

		if b == '/' && prefix {
			break
		}
		item = append(item, b)
		nextOffset++
	}

	return
}

func hackString(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

func hackSlice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}
