# tcp

## example
### 开启server
```
cd tcp/examples 
go test -v -test.run Test_tcpServer
```
```
=== RUN   Test_tcpServer
in hello world! 5577006791947779410 n=32
out hello world! 5577006791947779410 n=32
in hello world! 8674665223082153551 n=32
out hello world! 8674665223082153551 n=32
in hello world! 6129484611666145821 n=32
```
### 开启server(再开一个Terminal)
```
cd tcp/examples
go test -v -test.run Test_tcpClient
```
```
=== RUN   Test_tcpClient
in hello world! 5577006791947779410 n=32
out hello world! 5577006791947779410 n=32
in hello world! 8674665223082153551 n=32
out hello world! 8674665223082153551 n=32
in hello world! 6129484611666145821 n=32
out hello world! 6129484611666145821 n=32

```

## tcp server启动过程
* 监听端口,获取一个对象TCPListener{fd}
* TCPListener.fd基于epoll注册一个事件
* TCPListener开始accept tcp client的连接，针对每一个client过来的连接都会创建1个fd并且注册到epoll
* 后续每个连接的io读写都通过对应fd的epoll_wait异步完成

### 关键数据结构如下
    // poll.FD
    type FD struct {
        Sysfd int
        // I/O poller.
        pd pollDesc
    }
    
    // net.netFD
    // Network file descriptor.
    type netFD struct {
    	pfd poll.FD
    }
    
    // server listener
    type TCPListener struct {
        fd *netFD
        lc ListenConfig
    }
    
    // net.conn
    type conn struct {
    	fd *netFD
    }
通过上面的数据结构可以看出，不管是listener还是连接conn都由一个关键的成员fd组成，
这个fd就是注册进epoll的一个网络描述符，epoll_wait时，每次发现新的事件就会调用对应fd的
回调函数进行执行。

# todo epoll的具体实现及gopark原理

    
