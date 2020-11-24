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
