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

## epoll
四个关键函数

    func runtime_pollServerInit() // 启动epoll实例，初始化
    func runtime_pollOpen(fd uintptr) (uintptr, int) // 为fd注册一个epoll事件
    func runtime_pollWait(ctx uintptr, mode int) int // 等待fd有IO发生，触发注册的事件回调

### 1.runtime_pollServerInit

    // 进程启动时调用一次，不会重复调用
    func netpollGenericInit() {
    	if atomic.Load(&netpollInited) == 0 {
    		lock(&netpollInitLock)
    		if netpollInited == 0 {
    			netpollinit()
    			atomic.Store(&netpollInited, 1)
    		}
    		unlock(&netpollInitLock)
    	}
    }
    
### 2.runtime_pollOpen

    func poll_runtime_pollOpen(fd uintptr) (*pollDesc, int) {
        pd := pollcache.alloc()
        lock(&pd.lock)
        if pd.wg != 0 && pd.wg != pdReady {
            throw("runtime: blocked write on free polldesc")
        }
        if pd.rg != 0 && pd.rg != pdReady {
            throw("runtime: blocked read on free polldesc")
        }
        pd.fd = fd
        pd.closing = false
        pd.everr = false
        pd.rseq++
        pd.rg = 0
        pd.rd = 0
        pd.wseq++
        pd.wg = 0
        pd.wd = 0
        unlock(&pd.lock)
    
        var errno int32
        errno = netpollopen(fd, pd)
        return pd, int(errno)
    }
    
### 3.runtime_pollWait

    func poll_runtime_pollWait(pd *pollDesc, mode int) int {
        // 省略校验部分
        // 等待IO事件
        for !netpollblock(pd, int32(mode), false) {
            err = netpollcheckerr(pd, int32(mode))
            if err != 0 {
                return err
            }
            // Can happen if timeout has fired and unblocked us,
            // but before we had a chance to run, timeout has been reset.
            // Pretend it has not happened and retry.
        }
        return 0
    }
    
    // returns true if IO is ready, or false if timedout or closed
    // waitio - wait only for completed IO, ignore errors
    func netpollblock(pd *pollDesc, mode int32, waitio bool) bool {
    	gpp := &pd.rg
    	if mode == 'w' {
    		gpp = &pd.wg
    	}
    
    	// 设置gpp=WAIT
    	// 校验pd是否参数正常，如果校验通过就gopark,将当前的g放入sleep队列，进行IO等待，有IO事件时再把g放入等待队列中
    	if waitio || netpollcheckerr(pd, mode) == 0 {
    		gopark(netpollblockcommit, unsafe.Pointer(gpp), waitReasonIOWait, traceEvGoBlockNet, 5)
    	}
    	// 再次唤醒goroutine时需要坚持pd是否正常
    	old := atomic.Xchguintptr(gpp, 0)
    	if old > pdWait {
    		throw("runtime: corrupted polldesc")
    	}
    	return old == pdReady
    }

### 4.gopark

    // Puts the current goroutine into a waiting state and calls unlockf.
    // If unlockf returns false, the goroutine is resumed.
    // unlockf must not access this G's stack, as it may be moved between
    // the call to gopark and the call to unlockf.
    // Reason explains why the goroutine has been parked.
    // It is displayed in stack traces and heap dumps.
    // Reasons should be unique and descriptive.
    // Do not re-use reasons, add new ones.
    func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceEv byte, traceskip int) {
        if reason != waitReasonSleep {
            checkTimeouts() // timeouts may expire while two goroutines keep the scheduler busy
        }
        mp := acquirem() // 获取当前g对应的物理线程m
        gp := mp.curg // 当前的g
        status := readgstatus(gp)
        // g的状态必须是running才能gopark
        if status != _Grunning && status != _Gscanrunning {
            throw("gopark: bad g status")
        }
        mp.waitlock = lock
        mp.waitunlockf = unlockf
        gp.waitreason = reason
        mp.waittraceev = traceEv
        mp.waittraceskip = traceskip
        releasem(mp) 
        // can't do anything that might move the G between Ms here.
        mcall(park_m)
    }
    
    func park_m(gp *g) {
    	_g_ := getg()
    
    	if trace.enabled {
    		traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
    	}
    
        // running->waiting
    	casgstatus(gp, _Grunning, _Gwaiting)
    	dropg() // 将_g_从PM中移除，解绑
        // 释放G后执行回调函数 waitunlockf
    	if fn := _g_.m.waitunlockf; fn != nil {
    		ok := fn(gp, _g_.m.waitlock)
    		_g_.m.waitunlockf = nil
    		_g_.m.waitlock = nil
    		// 回调函数执行失败，需要再把G状态改回去
    		if !ok {
    			if trace.enabled {
    				traceGoUnpark(gp, 2)
    			}
    			casgstatus(gp, _Gwaiting, _Grunnable)
    			execute(gp, true) // Schedule it back, never returns.
    		}
    	}
    	// 让出调度资源
    	schedule()
    }

