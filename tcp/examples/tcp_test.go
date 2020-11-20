package examples

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

/*
tcp server将tcp client发送的内容直接返回
*/

type httpService struct {
	conn net.Conn
}

func (s httpService) run() {
	/*
		根据请求内容做相应处理，返回结果给client
	*/
	for {
		in := make([]byte, 1024)
		n, err := s.conn.Read(in)
		if err != nil {
			panic(err)
		}
		in = in[:n]
		fmt.Println(fmt.Sprintf("in %s n=%d", string(in), n))

		out := in
		nn, err := s.conn.Write(out)
		if err != nil {
			panic(err)
		}
		out = out[:n]
		fmt.Println(fmt.Sprintf("out %s n=%d", string(out), nn))
	}
}

func Test_tcpServer(t *testing.T) {
	addr := "0.0.0.0:8888"

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		srv := httpService{conn}
		// 开启协程处理单个连接的网络请求
		go srv.run()
	}
}

func Test_tcpClient(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Second * 2)
		rd := rand.Int()
		in := fmt.Sprintf("hello world! %d", rd)
		n, err := conn.Write([]byte(in))
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("in %s n=%d", string(in), n))

		out := make([]byte, 1024)
		nn, err := conn.Read(out)
		if err != nil {
			panic(err)
		}
		out = out[:n]
		fmt.Println(fmt.Sprintf("out %s n=%d", string(out), nn))
	}
}
