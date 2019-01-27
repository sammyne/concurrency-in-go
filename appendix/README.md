# Appendix

## Overview

This appendix will discuss some of these tools and how they can aid you before, during, and after developmen

## Anatomy of a Goroutine Error

- As of Go 1.6 and greater, only the stack trace of the panicking goroutine is printed

Given program as

```go
package main

func main() {
	waitForever := make(chan interface{})

	go func() {
		panic("test panic")
	}()

	<-waitForever
}
```

Stack trace goes as (where the angle bracket `<>` and the number surrounded by it is added mannually)

```bash
panic: test panic

goroutine 4 [running]:
main.main.func1() <3>
        /home/loccs/Workspaces/golang/src/github.com/sammyne/concurrency-in-go/appendix/panicking.go:7 +0x39 <1>
created by main.main
        /home/loccs/Workspaces/golang/src/github.com/sammyne/concurrency-in-go/appendix/panicking.go:6 +0x58 <2>
exit status 2
```

Comments

- `<1>` refers to where the panic occurred
- `<2>` refers to where the goroutine was started
- `<3>` indicates the name of goroutine function, and an anonymous function will assigned an automatic and unique identifier

- To see stack traces of all goroutines during panicking, set the `GOTRACEBACK` environment variable to `all`

## Race Detection

- Build/Install/Run/Test the program with `-race` flag
- One caveat: algorithm will only find races that are contained in code that is exercised
- The Go team recommends running a build of your application built with the race flag under real-world load
- some options specified via environmental variables, although generally the defaults are sufficient
  - `LOG_PATH`: tells the race detector to write reports to the `LOG_PATH.pid` file
  - `STRIP_PATH_PREFIX`: strip the beginnings of file paths in reports to make them more concise
  - `HISTORY_SIZE`
    - sets the per-goroutine history size, which controls how many previous memory accesses are remembered per goroutine
    - the valid range of values is [0, 7]
    - begins at 32 KB when `HISTORY_SIZE` is 0, and doubles with each subsequent value for a maximum of 4 MB at a `HISTORY_SIZE` of 7
    - can significantly increase memory consumption

### Demo

Given program as

```go
package main

import "fmt"

func main() {
	var data int

	go func() {
		data++
	}()

	if 0 == data {
		fmt.Printf("the value is %v.\n", data)
	}
}
```

Running with `-race` flag triggers an error as

```bash
the value is 0.
==================
WARNING: DATA RACE
Write at 0x00c00009a010 by goroutine 6:
  main.main.func1()
      /home/loccs/Workspaces/golang/src/github.com/sammyne/concurrency-in-go/appendix/race.go:11 +0x4e <1>

Previous read at 0x00c00009a010 by main goroutine:
  main.main()
      /home/loccs/Workspaces/golang/src/github.com/sammyne/concurrency-in-go/appendix/race.go:14 +0x88

Goroutine 6 (running) created at:
  main.main()
      /home/loccs/Workspaces/golang/src/github.com/sammyne/concurrency-in-go/appendix/race.go:10 +0x7a <2>
==================
Found 1 data race(s)
exit status 66
```

Comments

- `<1>` signals a unsynchronized write attempt
- `<2>` signals a unsynchronized read attempt

## pprof

- Motivations: monitoring for jobs such as
  - the number of running goroutines
  - utilization efficiency of CPUs
  - memory usage
- Tool: the standard `pprof` package
- `pprof` is a tool capable of displaying profile data either while a program is running, or by consuming saved runtime statistics
- Predefined profiles to hook into and display

| profile        | description                                                     |
| -------------- | --------------------------------------------------------------- |
| `goroutine`    | stack traces of all current goroutines                          |
| `heap`         | a sampling of all heap allocations                              |
| `threadcreate` | stack traces that led to the creation of new OS threads         |
| `block`        | stack traces that led to blocking on synchronization primitives |
| `mutex`        | stack traces of holders of contended mutexes                    |

- An example of detecting goroutine leaks by `pprof` as `goroutine_leaks.go`
- An exmaple of customized profile as `custom_profile.go`
