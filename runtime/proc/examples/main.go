package main

import "sync"

func main() {
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(wg *sync.WaitGroup) {
			var counter int
			for i := 0; i < 1e10; i++ {
				counter++
			}
			wg.Done()
		}(&wg)
	}

	wg.Wait()
}
