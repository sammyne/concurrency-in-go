package main

func main() {
	waitForever := make(chan interface{})

	go func() {
		panic("test panic")
	}()

	<-waitForever
}
