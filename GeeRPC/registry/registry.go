package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type GeeRegistry struct {
	timeout time.Duration
	mu      sync.Mutex // 保证互斥 servers
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultGeeRegistry = New(defaultTimeout)

// putServer 添加新的服务, 若服务存在则更新 start
func (r *GeeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.servers[addr]
	if ok {
		s.start = time.Now()
	} else {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	}
}

// aliveServers 返回所有可用的服务列表
func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			// 服务未过期
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// 将可用的服务放在请求头中, key = "X-GeeRPC-Servers"
		w.Header().Set("X-GeeRPC-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		// 注册服务时, 也从请求头中获得服务的地址
		addr := req.Header.Get("X-GeeRPC-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *GeeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("RPC Registry Path: ", registryPath)
}

func HandleHTTP() {
	DefaultGeeRegistry.HandleHTTP(defaultPath)
}

func Heartbeat(registry string, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHearbeat(registry, addr)
	go func() {
		// 定期发送心跳
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHearbeat(registry, addr)
		}
	}()

}

func sendHearbeat(registry string, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	// 通过重新发送 POST 请求模拟心跳
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-GeeRPC-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("RPC Server: heart beat error:", err)
		return err
	}
	return nil
}
