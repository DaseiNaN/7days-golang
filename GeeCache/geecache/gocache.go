package geecache

import (
	"fmt"
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 定义函数类型实现 Getter 接口 —— 接口型函数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string // 命名空间
	getter    Getter // 未命中缓存时用来获取数据源的回调函数
	mainCache cache  // 并发缓存
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	if g, ok := groups[name]; ok {
		return g
	}
	return nil
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 命中主存
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 未命中, 去其他节点获取
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer.", err)
			}
		}
		return g.getlocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return
	// 未处理缓存击穿
	// if g.peers != nil {
	// 	if peer, ok := g.peers.PickPeer(key); ok {
	// 		if value, err := g.getFromPeer(peer, key); err == nil {
	// 			return value, nil
	// 		}
	// 		log.Println("[GeeCache] Failed to get from peer.", err)
	// 	}
	// }
	// return g.getlocally(key)
}

func (g *Group) getlocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	// with protobuf
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}

	res := &pb.Response{}

	if err := peer.Get(req, res); err != nil {
		return ByteView{}, err
	} else {
		return ByteView{b: res.Value}, nil
	}

	// without protobuf
	// if bytes, err := peer.Get(g.name, key); err == nil {
	// 	return ByteView{b: bytes}, nil
	// } else {
	// 	return ByteView{}, err
	// }
}
