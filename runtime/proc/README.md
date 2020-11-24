# GMP

## 1.main函数启动过程

```
go build -gcflags "-N -l" -o main main.go     ##为了找到程序入口，需要禁止内联、优化
gdb main
```

    (gdb) info files
    Symbols from "/Users/wendongfang/work/gopath/golang-learning/runtime/proc/main".
    Local exec file:
            `/Users/wendongfang/work/gopath/golang-learning/runtime/proc/main', file type mach-o-x86-64.
            Entry point: 0x105cc30
            0x0000000001001000 - 0x000000000109d033 is .text
            0x000000000109d040 - 0x00000000010ec1b5 is __TEXT.__rodata
            0x00000000010ec1c0 - 0x00000000010ec2e6 is __TEXT.__symbol_stub1
            0x00000000010ec300 - 0x00000000010ece18 is __TEXT.__typelink
            0x00000000010ece18 - 0x00000000010ece88 is __TEXT.__itablink
            0x00000000010ece88 - 0x00000000010ece88 is __TEXT.__gosymtab
            0x00000000010ecea0 - 0x0000000001165c23 is __TEXT.__gopclntab
            0x0000000001166000 - 0x0000000001166020 is __DATA.__go_buildinfo
            0x0000000001166020 - 0x00000000011661a8 is __DATA.__nl_symbol_ptr
            0x00000000011661c0 - 0x00000000011742e0 is __DATA.__noptrdata
            0x00000000011742e0 - 0x000000000117b030 is .data
            0x000000000117b040 - 0x00000000011a4ab0 is .bss
            0x00000000011a4ac0 - 0x00000000011a7188 is __DATA.__noptrbss
    (gdb) b *0x105cc30
    Breakpoint 1 at 0x105cc30
    (gdb) info b
    Num     Type           Disp Enb Address            What
    1       breakpoint     keep y   0x000000000105cc30 <_rt0_amd64_darwin>
    
找到对应的函数是 _rt0_amd64,对应的一段代码

	// set the per-goroutine and per-mach "registers"
	get_tls(BX)
	LEAQ	runtime·g0(SB), CX
	MOVQ	CX, g(BX)
	LEAQ	runtime·m0(SB), AX

	// save m->g0 = g0
	MOVQ	CX, m_g0(AX)
	// save m0 to g0->m
	MOVQ	AX, g_m(CX)

	CLD				// convention is D is always left cleared
	CALL	runtime·check(SB)

	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP)
	CALL	runtime·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOVQ	$runtime·mainPC(SB), AX		// entry
	PUSHQ	AX
	PUSHQ	$0			// arg size
	CALL	runtime·newproc(SB)
	POPQ	AX
	POPQ	AX

	// start this M
	CALL	runtime·mstart(SB)

	CALL	runtime·abort(SB)	// mstart should never return
	RET

* 1.初始m0,g0,调度器
* 2.CALL	runtime·newproc(SB) 启动g0
* 3.init函数执行,gc开始,main函数执行开始


        func main() {
            g := getg() // 当前的g=g0
            g.m.g0.racectx = 0
            lockOSThread() // 锁住当前线程，防止重复初始化main
            if g.m != &m0 {
                throw("runtime.main not on m0")
            }
        
            doInit(&runtime_inittask) // 执行完所有package下的init函数
            if nanotime() == 0 {
                throw("nanotime returning zero")
            }
        
            // Defer unlock so that runtime.Goexit during init does the unlock too.
            needUnlock := true
            defer func() {
                if needUnlock {
                    unlockOSThread()
                }
            }()
        
            // Record when the world started.
            runtimeInitTime = nanotime()
        
            gcenable() // 开启gc相关的goroutine
            if iscgo {
                // cgo需要额外处理
            }
        
            doInit(&main_inittask)
            unlockOSThread()
      
            // 用户编写的main函数，在此处开始执行
            fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
            fn()
           
            exit(0)
            for {
                var x *int32
                *x = 0
            }
        }

关键点:

go的进程里最开始启动了m0,g0,main函数由g0,m0来执行的,gc由新的g并发执行，
程序运行时会不断的通过newproc启动新的g

## 2.GMP

![avatar](https://github.com/fangwendong/golang-learning/blob/main/runtime/proc/images/gmp.png)

图中的m为物理线程,p为调度者,g为逻辑执行的协程。m上面控制着系统资源(内存,cpu),每个m会绑定一个p用来调度G,每个p有一个local队列，按照顺序执行队列里的G。
local队列的G执行完了之后会从全局队列获取G来执行,如果全局队列空了会从其它p的local队列steal G执行。通过这样的调度方式来充分利用系统资源。
另外，网络轮询器发先有IO事件时会唤醒G，放进waiting队列中等待p执行。


通过GODEBUG来追踪下go进程运行时GMP的变化。
```
GODEBUG=schedtrace=1000 go run examples/main.go
```
    SCHED 0ms: gomaxprocs=8 idleprocs=5 threads=5 spinningthreads=1 idlethreads=0 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 0ms: gomaxprocs=8 idleprocs=7 threads=5 spinningthreads=0 idlethreads=3 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 1004ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 1007ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=41 [45 1 1 1 1 1 0 1]
    SCHED 2007ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 2010ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=62 [1 5 5 3 4 4 4 4]
    SCHED 3014ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 3012ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=64 [4 3 2 2 5 4 4 4]
    SCHED 4017ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 4016ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=73 [3 2 4 3 2 2 2 1]
    SCHED 5021ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 5021ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=78 [0 1 2 3 3 1 3 1]
    SCHED 6032ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 6031ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=87 [1 2 1 0 1 0 0 0]
    SCHED 7033ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 7040ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=55 [0 9 2 1 8 10 7 0]
    SCHED 8039ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 8048ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=47 [6 6 5 5 6 7 5 5]
    SCHED 9045ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 9055ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=55 [4 4 5 6 0 7 6 5]
    SCHED 10045ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 10064ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=61 [3 1 7 2 4 5 3 6]
    SCHED 11047ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 11073ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=73 [2 2 2 2 1 2 5 3]
    SCHED 12051ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 12076ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=59 [1 1 11 0 0 3 8 9]
    SCHED 13055ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 13076ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=49 [0 8 0 9 9 2 8 7]
    SCHED 14059ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 14079ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=41 [8 6 9 5 5 6 8 4]
    SCHED 15071ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 15087ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=53 [8 5 3 5 4 6 3 5]
    SCHED 16073ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 16094ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=53 [5 5 6 5 4 4 5 5]
    SCHED 17076ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 17105ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=62 [2 5 3 3 4 6 4 3]
    SCHED 18078ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 18106ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=71 [2 1 5 3 1 4 3 2]
    SCHED 19088ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 19117ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=57 [3 0 8 1 3 9 10 1]
    SCHED 20093ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 20128ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=63 [1 0 9 0 0 11 0 8]
    SCHED 21094ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 21130ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=50 [7 8 7 1 6 5 1 7]
    SCHED 22103ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 22139ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=48 [6 7 0 6 7 7 5 6]
    SCHED 23111ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 23147ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=46 [6 6 7 7 5 5 6 4]
    SCHED 24120ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 24148ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=58 [0 7 5 7 4 4 4 3]
    SCHED 25126ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 25158ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=52 [7 5 3 3 8 5 5 4]
    SCHED 26138ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 26160ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=64 [6 3 3 2 4 4 3 3]
    SCHED 27147ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 27164ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=69 [2 3 2 2 2 4 5 3]
    SCHED 28151ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 28167ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=68 [4 3 2 4 1 2 3 5]
    SCHED 29156ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 29168ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=72 [2 5 3 1 3 3 1 2]
    SCHED 30160ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 30170ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=60 [10 0 9 1 8 0 3 1]
    SCHED 31167ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 31173ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=54 [0 8 7 0 7 8 0 8]
    SCHED 32171ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 32178ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=55 [0 6 7 6 0 5 6 7]
    SCHED 33178ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 33189ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=59 [5 4 3 5 4 4 3 5]
    SCHED 34187ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 34190ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=67 [3 3 3 2 3 4 4 3]
    SCHED 35191ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 35191ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=75 [4 2 1 1 3 3 2 1]
    SCHED 36203ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 36195ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=81 [2 1 2 0 2 1 0 3]
    SCHED 37206ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 37195ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=63 [11 9 1 0 0 0 0 8]
    SCHED 38215ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 38197ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=48 [0 11 8 8 7 0 1 9]
    SCHED 39223ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 39205ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=54 [7 0 7 8 8 0 8 0]
    SCHED 40225ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 40206ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=52 [6 7 4 8 6 5 0 4]
    SCHED 41230ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 41213ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=60 [4 4 3 4 4 4 4 5]
    SCHED 42233ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 42218ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=64 [3 4 2 3 4 4 3 5]
    SCHED 43235ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 43229ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=76 [5 2 2 3 1 1 1 1]
    SCHED 44241ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 44239ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=62 [8 11 1 0 0 1 9 0]
    SCHED 45252ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 45244ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=42 [7 11 8 8 0 6 10 0]
    SCHED 46255ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 46249ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=36 [8 9 6 5 6 7 10 5]
    SCHED 47260ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 47254ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=54 [5 4 5 6 5 4 5 4]
    SCHED 48261ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 48255ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=61 [5 3 5 4 3 2 4 5]
    SCHED 49265ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 49264ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=55 [5 7 5 3 4 7 3 3]
    SCHED 50270ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 50273ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=62 [3 4 3 3 5 4 4 4]
    SCHED 51276ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 51278ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=66 [2 3 4 8 4 0 4 1]
    SCHED 52280ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 52282ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=75 [3 0 1 1 2 0 0 10]
    SCHED 53290ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 53283ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=80 [0 0 11 0 0 0 0 1]
    SCHED 54291ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 54283ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=56 [1 7 7 6 7 0 8 0]
    SCHED 55292ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 55289ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=54 [4 5 6 4 7 6 6 0]
    SCHED 56296ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 56300ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=68 [3 5 3 3 3 2 2 3]
    SCHED 57304ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 57307ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=71 [2 1 0 1 2 3 10 2]
    SCHED 58312ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 58312ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=66 [10 0 9 1 2 3 1 0]
    SCHED 59320ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 59323ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=44 [7 8 6 7 6 0 7 7]
    SCHED 60322ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 60330ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=53 [5 6 5 7 5 5 6 0]
    SCHED 61335ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 61340ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=55 [5 5 4 5 3 5 6 4]
    SCHED 62340ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 62345ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=65 [2 3 2 4 4 4 3 5]
    SCHED 63343ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 63347ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=64 [0 1 2 2 10 2 2 9]
    SCHED 64350ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 64347ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=44 [6 6 6 7 8 9 6 0]
    SCHED 65356ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 65352ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=45 [5 7 6 7 4 6 4 8]
    SCHED 66367ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 66353ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=59 [5 5 4 4 4 4 4 3]
    SCHED 67373ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 67354ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=65 [3 2 5 3 3 3 3 5]
    SCHED 68383ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 68354ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=74 [3 2 3 2 4 2 1 1]
    SCHED 69386ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 69360ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=71 [1 1 0 9 9 1 0 0]
    SCHED 70395ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 70365ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=52 [0 6 5 5 6 7 6 5]
    SCHED 71398ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 71373ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=53 [5 6 6 5 7 5 0 5]
    SCHED 72404ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 72380ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=49 [5 6 5 6 4 5 4 8]
    SCHED 73404ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 73387ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=61 [3 2 2 4 4 6 7 2]
    SCHED 74414ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 74395ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=79 [1 4 0 4 0 0 2 1]
    SCHED 75422ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 75400ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=39 [7 7 7 6 5 7 0 6]
    SCHED 76426ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 76405ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=64 [9 1 1 0 0 1 0 0]
    SCHED 77427ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 77410ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=54 [0 2 1 1 0 2 0 0]
    SCHED 78430ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    SCHED 78418ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=23 [3 0 2 3 3 0 2 2]
    SCHED 79437ms: gomaxprocs=8 idleprocs=8 threads=13 spinningthreads=0 idlethreads=6 runqueue=0 [0 0 0 0 0 0 0 0]
    

* sched：每一行都代表调度器的调试信息，后面提示的毫秒数表示启动到现在的运行时间，输出的时间间隔受 schedtrace 的值影响。
* gomaxprocs：当前的 CPU 核心数（GOMAXPROCS 的当前值）。
* idleprocs：空闲的处理器数量，后面的数字表示当前的空闲数量。
* threads：OS 线程数量，后面的数字表示当前正在运行的线程数量。
* spinningthreads：自旋状态的 OS 线程数量。
* idlethreads：空闲的线程数量。
* runqueue：全局队列中中的 Goroutine 数量，而后面的 [6 7 0 6 7 7 5 6] 则分别代表这 8 个 P 的本地队列正在运行的 Goroutine 数量。

机器cpu=8,开100个协程跑起，观察GMP数量变化，可以发现正在运行的m的数量最多时才13个。
全局队列长度和本地队列的长度一直在动态变化，维持平衡，保证每个m都在干活，不会空置。

可以通过下面命令看一下每一个GMP的详细状态值
```
 GODEBUG=scheddetail=1,schedtrace=1000 go run examples/main.go
```
    SCHED 2015ms: gomaxprocs=8 idleprocs=8 threads=12 spinningthreads=0 idlethreads=5 runqueue=0 gcwaiting=0 nmidlelocked=1 stopwait=0 sysmonwait=0
      P0: status=0 schedtick=316 syscalltick=1728 m=-1 runqsize=0 gfreecnt=0 timerslen=0
      P1: status=0 schedtick=707 syscalltick=81 m=-1 runqsize=0 gfreecnt=0 timerslen=0
      P2: status=0 schedtick=61 syscalltick=2789 m=-1 runqsize=0 gfreecnt=4 timerslen=0
      P3: status=0 schedtick=63 syscalltick=776 m=-1 runqsize=0 gfreecnt=5 timerslen=0
      P4: status=0 schedtick=9 syscalltick=112 m=-1 runqsize=0 gfreecnt=1 timerslen=0
      P5: status=0 schedtick=56 syscalltick=815 m=-1 runqsize=0 gfreecnt=2 timerslen=0
      P6: status=0 schedtick=13 syscalltick=157 m=-1 runqsize=0 gfreecnt=2 timerslen=0
      P7: status=0 schedtick=8 syscalltick=557 m=-1 runqsize=0 gfreecnt=0 timerslen=0
      M11: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M10: p=-1 curg=55 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      M9: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M8: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      M7: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M6: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M5: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M4: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=-1
      M3: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=true lockedg=56
      M2: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M1: p=-1 curg=17 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=17
      M0: p=-1 curg=12 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      G1: status=4(semacquire) m=-1 lockedm=-1
      G17: status=6() m=1 lockedm=1
      G2: status=4(force gc (idle)) m=-1 lockedm=-1
      G3: status=4(GC sweep wait) m=-1 lockedm=-1
      G4: status=4(GC scavenge wait) m=-1 lockedm=-1
      G18: status=4(finalizer wait) m=-1 lockedm=-1
      G34: status=4(GC worker (idle)) m=-1 lockedm=-1
      G5: status=4(GC worker (idle)) m=-1 lockedm=-1
      G6: status=4(GC worker (idle)) m=-1 lockedm=-1
      G7: status=4(GC worker (idle)) m=-1 lockedm=-1
      G35: status=4(GC worker (idle)) m=-1 lockedm=-1
      G8: status=4(GC worker (idle)) m=-1 lockedm=-1
      G19: status=4(GC worker (idle)) m=-1 lockedm=-1
      G36: status=4(GC worker (idle)) m=-1 lockedm=-1
      G9: status=4(select) m=-1 lockedm=-1
      G10: status=4(select) m=-1 lockedm=-1
      G11: status=4(select) m=-1 lockedm=-1
      G12: status=3(chan send) m=0 lockedm=-1
      G13: status=4(select) m=-1 lockedm=-1
      G14: status=4(select) m=-1 lockedm=-1
      G15: status=4(select) m=-1 lockedm=-1
      G16: status=4(select) m=-1 lockedm=-1
      G50: status=6() m=-1 lockedm=-1
      G51: status=6() m=-1 lockedm=-1
      G55: status=3() m=10 lockedm=-1 // 正在进行syscall，绑定在M10执行
      G67: status=6() m=-1 lockedm=-1
      G53: status=6() m=-1 lockedm=-1
      G83: status=6() m=-1 lockedm=-1
      G98: status=6() m=-1 lockedm=-1
      G100: status=6() m=-1 lockedm=-1
      G52: status=6() m=-1 lockedm=-1
      G38: status=6() m=-1 lockedm=-1
      G68: status=6() m=-1 lockedm=-1
      G69: status=6() m=-1 lockedm=-1
      G101: status=6() m=-1 lockedm=-1
      G39: status=6() m=-1 lockedm=-1
      G40: status=6() m=-1 lockedm=-1
      G56: status=4(select) m=-1 lockedm=3
      G57: status=4(chan receive) m=-1 lockedm=-1

    SCHED 2019ms: gomaxprocs=8 idleprocs=0 threads=9 spinningthreads=0 idlethreads=0 runqueue=43 gcwaiting=0 nmidlelocked=0 stopwait=0 sysmonwait=0
      P0: status=1 schedtick=91 syscalltick=0 m=0 runqsize=7 gfreecnt=0 timerslen=0 // 绑定在M0上，本地队列长度为7
      P1: status=1 schedtick=90 syscalltick=0 m=3 runqsize=7 gfreecnt=0 timerslen=0
      P2: status=1 schedtick=89 syscalltick=0 m=2 runqsize=6 gfreecnt=0 timerslen=0
      P3: status=1 schedtick=89 syscalltick=0 m=4 runqsize=7 gfreecnt=0 timerslen=0
      P4: status=1 schedtick=88 syscalltick=0 m=5 runqsize=7 gfreecnt=0 timerslen=0
      P5: status=1 schedtick=88 syscalltick=0 m=6 runqsize=6 gfreecnt=0 timerslen=0
      P6: status=1 schedtick=88 syscalltick=0 m=7 runqsize=8 gfreecnt=0 timerslen=0
      P7: status=1 schedtick=88 syscalltick=0 m=8 runqsize=1 gfreecnt=0 timerslen=0
      M8: p=7 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M7: p=6 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M6: p=5 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M5: p=4 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M4: p=3 curg=10 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      M3: p=1 curg=72 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      M2: p=2 curg=94 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      M1: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 spinning=false blocked=false lockedg=-1
      M0: p=0 curg=38 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 spinning=false blocked=false lockedg=-1
      G1: status=4(semacquire) m=-1 lockedm=-1
      G2: status=4(force gc (idle)) m=-1 lockedm=-1
      G3: status=4(GC sweep wait) m=-1 lockedm=-1 // GC等待
      G4: status=4(GC scavenge wait) m=-1 lockedm=-1
      G5: status=1() m=-1 lockedm=-1
      G6: status=1() m=-1 lockedm=-1
      G7: status=1() m=-1 lockedm=-1
      G8: status=1() m=-1 lockedm=-1
      G9: status=1() m=-1 lockedm=-1
      G10: status=2() m=4 lockedm=-1
      G11: status=1() m=-1 lockedm=-1
      G12: status=1() m=-1 lockedm=-1
      G13: status=1() m=-1 lockedm=-1
      G14: status=1() m=-1 lockedm=-1
      G15: status=1() m=-1 lockedm=-1
      G16: status=1() m=-1 lockedm=-1
      G17: status=1() m=-1 lockedm=-1
      G18: status=1() m=-1 lockedm=-1
      G19: status=1() m=-1 lockedm=-1
      G20: status=1() m=-1 lockedm=-1
      G21: status=1() m=-1 lockedm=-1
      G22: status=1() m=-1 lockedm=-1
      G23: status=1() m=-1 lockedm=-1
      G24: status=1() m=-1 lockedm=-1
      G25: status=1() m=-1 lockedm=-1
      G26: status=1() m=-1 lockedm=-1
      G27: status=1() m=-1 lockedm=-1
      G28: status=1() m=-1 lockedm=-1
      G29: status=1() m=-1 lockedm=-1
      G30: status=1() m=-1 lockedm=-1
      G31: status=1() m=-1 lockedm=-1
      G32: status=1() m=-1 lockedm=-1
      G33: status=1() m=-1 lockedm=-1
      G34: status=1() m=-1 lockedm=-1
      G35: status=1() m=-1 lockedm=-1
      G36: status=1() m=-1 lockedm=-1
      G37: status=1() m=-1 lockedm=-1
      G38: status=2() m=0 lockedm=-1
      G39: status=1() m=-1 lockedm=-1
      G40: status=1() m=-1 lockedm=-1
      G41: status=1() m=-1 lockedm=-1
      G42: status=1() m=-1 lockedm=-1
      G43: status=1() m=-1 lockedm=-1
      G44: status=1() m=-1 lockedm=-1
      G45: status=1() m=-1 lockedm=-1
      G46: status=1() m=-1 lockedm=-1
      G47: status=1() m=-1 lockedm=-1
      G48: status=1() m=-1 lockedm=-1
      G49: status=1() m=-1 lockedm=-1
      G50: status=1() m=-1 lockedm=-1
      G51: status=1() m=-1 lockedm=-1
      G52: status=1() m=-1 lockedm=-1
      G53: status=1() m=-1 lockedm=-1
      G54: status=1() m=-1 lockedm=-1
      G55: status=1() m=-1 lockedm=-1
      G56: status=1() m=-1 lockedm=-1
      G57: status=1() m=-1 lockedm=-1
      G58: status=1() m=-1 lockedm=-1
      G59: status=1() m=-1 lockedm=-1
      G60: status=1() m=-1 lockedm=-1
      G61: status=1() m=-1 lockedm=-1
      G62: status=1() m=-1 lockedm=-1
      G63: status=1() m=-1 lockedm=-1
      G64: status=1() m=-1 lockedm=-1
      G65: status=1() m=-1 lockedm=-1
      G66: status=1() m=-1 lockedm=-1
      G67: status=1() m=-1 lockedm=-1
      G68: status=1() m=-1 lockedm=-1
      G69: status=1() m=-1 lockedm=-1
      G70: status=1() m=-1 lockedm=-1
      G71: status=1() m=-1 lockedm=-1
      G72: status=2() m=3 lockedm=-1
      G73: status=1() m=-1 lockedm=-1
      G74: status=1() m=-1 lockedm=-1
      G75: status=1() m=-1 lockedm=-1
      G76: status=1() m=-1 lockedm=-1
      G77: status=1() m=-1 lockedm=-1
      G78: status=1() m=-1 lockedm=-1
      G79: status=1() m=-1 lockedm=-1
      G80: status=1() m=-1 lockedm=-1
      G81: status=1() m=-1 lockedm=-1
      G82: status=1() m=-1 lockedm=-1
      G83: status=1() m=-1 lockedm=-1
      G84: status=1() m=-1 lockedm=-1
      G85: status=1() m=-1 lockedm=-1
      G86: status=1() m=-1 lockedm=-1
      G87: status=1() m=-1 lockedm=-1
      G88: status=1() m=-1 lockedm=-1
      G89: status=1() m=-1 lockedm=-1
      G90: status=1() m=-1 lockedm=-1
      G91: status=1() m=-1 lockedm=-1
      G92: status=1() m=-1 lockedm=-1
      G93: status=1() m=-1 lockedm=-1
      G94: status=2() m=2 lockedm=-1
      G95: status=1() m=-1 lockedm=-1
      G96: status=1() m=-1 lockedm=-1
      G97: status=1() m=-1 lockedm=-1
      G98: status=1() m=-1 lockedm=-1
      G99: status=1() m=-1 lockedm=-1
      G100: status=1() m=-1 lockedm=-1
      G101: status=1() m=-1 lockedm=-1
      G102: status=1() m=-1 lockedm=-1
      G103: status=1() m=-1 lockedm=-1
      G104: status=1() m=-1 lockedm=-1 // 在队列中，不会绑定m，等待p执行

G状态值枚举

|状态|	值|	含义|
|---|---|---|
_Gidle|	0|	刚刚被分配，还没有进行初始化。|
_Grunnable	|1|	已经在运行队列中，还没有执行用户代码。|
_Grunning	|2|	不在运行队列里中，已经可以执行用户代码，此时已经分配了 M 和 P。|
_Gsyscall	|3|	正在执行系统调用，此时分配了 M。|
_Gwaiting	|4|	在运行时被阻止，没有执行用户代码，也不在运行队列中，此时它正在某处阻塞等待中。|
_Gmoribund_unused	|5|	尚未使用，但是在 gdb 中进行了硬编码。|
_Gdead	|6|	尚未使用，这个状态可能是刚退出或是刚被初始化，此时它并没有执行用户代码，有可能有也有可能没有分配堆栈。|
_Genqueue_unused	|7|	尚未使用。|
_Gcopystack	|8|	正在复制堆栈，并没有执行用户代码，也不在运行队列中。|

p状态值枚举

|状态	|值|	含义|
|---|---|---|
_Pidle	|0|	刚刚被分配，还没有进行进行初始化。|
_Prunning	|1|	当 M 与 P 绑定调用 acquirep 时，P 的状态会改变为 _Prunning。|
_Psyscall	|2|	正在执行系统调用。|
_Pgcstop	|3|	暂停运行，此时系统正在进行 GC，直至 GC 结束后才会转变到下一个状态阶段。|
_Pdead	|4|	废弃，不再使用。|


可以看得出来很多G处于waiting状态，包括负责GC的G。有一些G处于syscall状态，在指定的m上执行，没有绑定p。
有些G处于running状态，会有绑定的p和m。有些G处于runnable状态，在队列中，此时不会绑定m

    G3: status=4(GC sweep wait) m=-1 lockedm=-1 // GC等待，没有绑定m
    G55: status=3() m=10 lockedm=-1 // 正在进行syscall，绑定在M10执行
    P0: status=1 schedtick=91 syscalltick=0 m=0 runqsize=7 gfreecnt=0 timerslen=0 // 绑定在M0上，本地队列长度为7
    G104: status=1() m=-1 lockedm=-1 // 在队列中，不会绑定m，等待p执行