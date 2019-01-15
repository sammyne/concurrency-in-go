package main

import (
	"fmt"
	"time"
)

func orDone(done <-chan struct{}, c <-chan interface{}) <-chan interface{} {
	valStream := make(chan interface{})
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-done:
				}
			}
		}
	}()
	return valStream
}

func main() {
	done := make(chan struct{})

	go func() {
		time.Sleep(time.Second * 2)
		close(done)
	}()

	c := make(chan interface{}, 8)
	for i := 0; i < 8; i++ {
		c <- i
	}

	for v := range orDone(done, c) {
		fmt.Print(v, " ")
	}

	<-done
	fmt.Println("Done")
}
