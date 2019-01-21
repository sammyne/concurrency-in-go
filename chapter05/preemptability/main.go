package main

import "time"

func doWork(done <-chan interface{},
	valueStream <-chan interface{}) <-chan interface{} {

	resultStream := make(chan interface{})

	go func() {

		var value interface{}

		select {
		case <-done:
			return
		case value = <-valueStream:
		}

		// the next line is non-preemptable
		result := reallyLongCalculation(done, value)

		select {
		case <-done:
			return
		case resultStream <- result:
		}
	}()

	return resultStream
}

func main() {
}

func longCalculation(done <-chan interface{}, v interface{}) interface{} {
	select {
	case <-time.After(time.Hour):
	case <-done:
	}

	return nil
}

func reallyLongCalculation(done <-chan interface{}, v interface{}) interface{} {
	intermediateResult := longCalculation(done, v)
	return longCalculation(done, intermediateResult)
}
