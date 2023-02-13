package geerpc

import (
	"encoding/json"
	"geerpc/codec/codec"
	"io"
	"log"
	"net"
	"sync"
)

/*
通信过程

客户端与服务端的通信需要协商一些内容
服务端通过解析header就能够知道如何从body中读取需要的信息。对于 RPC 协议来说，这部分协商是需要自主设计的。
为了提升性能，一般在报文的最开始会规划固定的字节，来协商相关的信息。
比如第1个字节用来表示序列化方式
第2个字节表示压缩方式，
第3-6字节表示 header 的长度，
7-10 字节表示 body 的长度。
*/

const MagicNumber = 0x3bef5c

// Option 消息的编解码方式。我们将这部分信息，放到结构体 Option 中承载
type Option struct {
	MagicNumber int
	CodecType   codec.Type // 客户端会选择不同的消息体进行编码
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// Server 一个RPC服务器
type Server struct{}

// NewServer 创建一个新的服务端
func NewServer() *Server {
	return &Server{}
}

// DefaultServer 默认的服务端实例
var DefaultServer = NewServer()

// Accept 为每个监听到的连接提供请求服务
func (server *Server) Accept(lis net.Listener) {
	// for 循环等待 socket 连接建立
	// 并开启子协程处理，处理过程交给了 ServerConn 方法。
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	// 阻塞 直到连接关闭
	defer func() { _ = conn.Close() }()
	var opt Option

	// 反序列化 得到Opt实例
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}

	// 验证magic num
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	f := codec.NewCodeFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // 互斥锁

	// 并发控制方式，它可以让我们的代码等待一组 goroutine 的结束。
	// 比如在主协程中等待几个子协程去做一些耗时的操作
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			// req 空不可恢复，故关闭连接
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.
		}
	}
}
