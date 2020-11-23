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

*   1.初始m0,g0,调度器
*   2. CALL	runtime·newproc(SB) 启动g0