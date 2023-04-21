package GeeRPC

import (
	"GeeRPC/codec"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Call 代表一个活跃的 RPC 请求
type Call struct {
	Seq           uint64      // 请求的序号, 用来区分不同的请求; 由客户端选定;
	ServiceMethod string      // "Service.Method";
	Args          interface{} // 调用方法的参数;
	Reply         interface{} // 调用方法的返回值;
	Error         error       // 错误信息;
	Done          chan *Call  // 调用完成时的信号;
}

// 调用结束时，通知调用方;
func (call *Call) done() {
	call.Done <- call
}

// Client 代表一个 RPC 的客户端
type Client struct {
	cc       codec.Codec      // 编码解码器;
	opt      *Option          // Option 部分;
	sending  sync.Mutex       // 保证请求的有序发送, 防止多个请求报文混淆;
	header   codec.Header     // Header 部分;
	mu       sync.Mutex       // 保证写 pending 时的互斥锁;
	seq      uint64           // 请求的序号, 每个请求的序号是唯一的;
	pending  map[uint64]*Call // seq <-> Call;
	closing  bool             // 客户端结束, 用户主动关闭;
	shutdown bool             // 服务端结束, 用户被动关闭;
}

// 保证 Client 实现了 io.Closer 接口
var _ io.Closer = (*Client)(nil)

var ErrShutDown = errors.New("connection is shut down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		// 不是客户端主动关闭的请求
		return ErrShutDown
	}
	client.closing = true
	// 需要关闭连接
	return client.cc.Close()
}

// IsAvailable 检查客户端是否可用
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

// registerCall 注册一个 RPC Call
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutDown
	}
	call.Seq = client.seq
	client.pending[client.seq] = call
	client.seq += 1
	return call.Seq, nil
}

// removeCall 删除一个 RPC Call
// 根据 seq，从 client.pending 中移除对应的 call，并返回。
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// terminateCalls 服务端或客户端发生错误时调用
// 将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call。
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		// 从 pending 中找到对应的 RPC call
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// call 不存在, 可能是请求没有发送完整或者因为其他原因被取消, 但是服务端仍然处理了;
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			// call 存在, h.Error 不为空, 说明服务端错误;
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			// call 存在, 服务端处理正常, 读取 Reply
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	client.terminateCalls(err)
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		return nil, err
	}

	// 和服务端进行协议交换
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("RPC Client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     opt,
		seq:     1,
		pending: make(map[uint64]*Call),
	}
	// 子协程接收响应
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	// 如果为空则使用默认值
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	// 只能有一个 option
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial 创建 RPC 连接, 返回 Client 实例
func Dial(network string, address string, opts ...*Option) (client *Client, err error) {
	//opt, err := parseOptions(opts...)
	//if err != nil {
	//	return nil, err
	//}
	//conn, err := net.Dial(network, address)
	//if err != nil {
	//	return nil, err
	//}
	//
	//defer func() {
	//	if client == nil {
	//		_ = conn.Close()
	//	}
	//}()
	//return NewClient(conn, opt)
	return dialTimeout(NewClient, network, address, opts...)
}

// send 发送 RPC Call 请求
func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	// 登记 RPC Call
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 准备请求头
	client.header.Seq = seq
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Error = ""

	// 编码并发送请求
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go 异步调用函数
// 返回表示调用的 Call 的结构体
func (client *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("RPC Client: done channel is unbuffered")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

// Call 调用服务端的方法, 并返回错误信息
//func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
//	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
//	return call.Error
//}

func (client *Client) Call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("RPC Client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network string, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()

	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()

	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("RPC Client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("Unexpected HTTP Response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network string, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, address, opts...)
}

func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("RPC Client Error: wrong format %q, expect protocol@addr", rpcAddr)
	}

	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		return Dial(protocol, addr, opts...)
	}
}
