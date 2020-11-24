package main

import (
	"fmt"
	"time"
)

/*
start send chanTmp2
end send chanTmp2
start send chanTmp3
start send chanTmp1
end send chanTmp3
chanTmp2 receive 0
chanTmp2 receive 1
*/
func main() {
	chanTmp1 := make(chan int)
	chanTmp2 := make(chan int, 2)
	chanTmp3 := make(chan int)
	go func() {
		fmt.Println("start send chanTmp1")
		chanTmp1 <- 0
		fmt.Println("end send chanTmp1")
	}()
	go func() {
		fmt.Println("start send chanTmp2")
		chanTmp2 <- 0
		chanTmp2 <- 1
		close(chanTmp2)
		fmt.Println("end send chanTmp2")
	}()
	go func() {
		fmt.Println("start send chanTmp3")
		chanTmp3 <- 0
		fmt.Println("end send chanTmp3")
	}()
	<-chanTmp3
	time.Sleep(time.Second * 3)
	for v := range chanTmp2 {
		fmt.Println("chanTmp2 receive", v)
	}
}
