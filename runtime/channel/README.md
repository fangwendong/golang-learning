# channel

## 关键数据结构

    type hchan struct {
        qcount   uint           // total data in the queue
        dataqsiz uint           // size of the circular queue
        buf      unsafe.Pointer // points to an array of dataqsiz elements
        elemsize uint16
        closed   uint32
        elemtype *_type // element type
        sendx    uint   // send index
        recvx    uint   // receive index
        recvq    waitq  // list of recv waiters
        sendq    waitq  // list of send waiters
    
        // lock protects all fields in hchan, as well as several
        // fields in sudogs blocked on this channel.
        //
        // Do not change another G's status while holding this lock
        // (in particular, do not ready a G), as this can deadlock
        // with stack shrinking.
        lock mutex
    }
    // 双向链表保存g队列
    type waitq struct {
    	first *sudog
    	last  *sudog
    }
    
![avatar](https://github.com/fangwendong/golang-learning/blob/master/runtime/channel/images/channel.png)

## makechan
* 1.计算出申请地址空间长度校验是否超过maxAlloc
* 2.buf申请地址空间
* 3.通过传参chantype和size对hchan进行赋值


    func makechan(t *chantype, size int) *hchan {
        elem := t.elem
        // 申请的地址长度不能超过maxAlloc
        mem, overflow := math.MulUintptr(elem.size, uintptr(size))
        if overflow || mem > maxAlloc-hchanSize || size < 0 {
            panic(plainError("makechan: size out of range"))
        }
    
        var c *hchan
        switch {
        case mem == 0:
            // Queue or element size is zero.
            c = (*hchan)(mallocgc(hchanSize, nil, true))
            // Race detector uses this location for synchronization.
            c.buf = c.raceaddr()
        case elem.ptrdata == 0:
            // Elements do not contain pointers.
            // Allocate hchan and buf in one call.
            c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
            c.buf = add(unsafe.Pointer(c), hchanSize)
        default:
            // Elements contain pointers.
            c = new(hchan)
            c.buf = mallocgc(mem, elem, true)
        }
    
        c.elemsize = uint16(elem.size)
        c.elemtype = elem
        c.dataqsiz = uint(size)
    
        return c
    }

## chansend

    func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
        // 如果c==nil，gopark当前的g,让出p
        if c == nil {
            if !block {
                return false
            }
            gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
            throw("unreachable")
        }
    
        if !block && c.closed == 0 && ((c.dataqsiz == 0 && c.recvq.first == nil) ||
            (c.dataqsiz > 0 && c.qcount == c.dataqsiz)) {
            return false
        }
    
        var t0 int64
        if blockprofilerate > 0 {
            t0 = cputicks()
        }
    
        lock(&c.lock)
    
        // 如果channel被关闭了，调用send函数会直接panic
        if c.closed != 0 {
            unlock(&c.lock)
            panic(plainError("send on closed channel"))
        }
        // 检查recvq队列中是否有g，如果有g则出队一个sg，直接将数据ep给sg
        if sg := c.recvq.dequeue(); sg != nil {
            // Found a waiting receiver. We pass the value we want to send
            // directly to the receiver, bypassing the channel buffer (if any).
            send(c, sg, ep, func() { unlock(&c.lock) }, 3)
            return true
        }
        
        // buf缓冲队列未满 
        if c.qcount < c.dataqsiz {
            // 入队，获取队尾位置的指针qp
            qp := chanbuf(c, c.sendx)
            // 将数据ep赋值到qp
            typedmemmove(c.elemtype, qp, ep)
            // 入队成功，队尾下标后移一位
            c.sendx++
            // 环形队列，如果到数组末尾了，就回到0的位置
            if c.sendx == c.dataqsiz {
                c.sendx = 0
            }
            c.qcount++
            unlock(&c.lock)
            return true
        }
    
        // buf缓冲队列满了
        if !block {
            unlock(&c.lock)
            return false
        }
    
        // Block on the channel. Some receiver will complete our operation for us.
        gp := getg()
        mysg := acquireSudog()
        mysg.releasetime = 0
        if t0 != 0 {
            mysg.releasetime = -1
        }
        // No stack splits between assigning elem and enqueuing mysg
        // on gp.waiting where copystack can find it.
        mysg.elem = ep
        mysg.waitlink = nil
        mysg.g = gp
        mysg.isSelect = false
        mysg.c = c
        gp.waiting = mysg
        gp.param = nil
        // 当前的g加入sendq队列
        c.sendq.enqueue(mysg)
        // gopark让出当前g占用的p和m,等待被唤醒
        gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
        KeepAlive(ep)
    
        // g被再次执行时从这里开始
        if mysg != gp.waiting {
            throw("G waiting list is corrupted")
        }
        gp.waiting = nil
        gp.activeStackChans = false
        if gp.param == nil {
            if c.closed == 0 {
                throw("chansend: spurious wakeup")
            }
            panic(plainError("send on closed channel"))
        }
        gp.param = nil
        if mysg.releasetime > 0 {
            blockevent(mysg.releasetime-t0, 2)
        }
        mysg.c = nil
        releaseSudog(mysg)
        return true
    }