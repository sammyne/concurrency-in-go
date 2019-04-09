# Chapter06. Goroutines and the Go Runtime

## WORK STEALING

- **Working stealing**: The algorithm employed by Go to multiplex goroutines onto OS threads
- 2 main concerns

  - Processor utilization
  - Cache locality

> Fork-Join model
>
> - Forks are when goroutines are started
> - Join points are when two or more goroutines are synchronized through `channel` or types in the `sync` package.

### Implementation Strategies

#### Fair scheduling

- A naive strategy for sharing work across many processors
- Problems
  - In a fork-join paradigm, **tasks are likely dependent on one another**, and it turns out naively splitting them among processors will likely cause one of the processors to be underutilized
  - Lead to **poor cache locality** as tasks that require the same data are scheduled on other processors

#### A FIFO priority queue

- Strategy
  - Work tasks get scheduled into the queue
  - Processors dequeue tasks as they have capacity, or block on joins
- Better than fair scheduling
- Introduce a centralized data structure shared across all the processors, as well as **problems**
  - Synchronization cost induced by entering and exiting critical sections
  - Cache locality problems have only been exacerbated
- Goroutines usually aren't coarse-grained, making a centralized queue probably not a great choice

#### Decentralized work queues

- Strategy: each processor is given its own thread and a double-ended queue (a.k.a. _deque_)
- Basic rules
  - At a fork point, add tasks to the **tail** of the deque associated with the thread
  - If the thread is idle, steal work from the **head** of deque associated with some other random thread
  - At a join point that cannot be realized yet (i.e., the goroutine it is synchronized with has not completed yet), pop work off the **tail** of the thread's own deque.
  - If the thread's deque is empty, either:
    - Stall at a join
    - Steal work from the **head** of a random thread's associated deque.
- The work sitting on the tail of its deque has a couple of interesting properties (as showed later)
  - It's the work most likely needed to complete the parent's join
  - It's the work most likely to still be in our processor's cache
- This is Go's implementation of work stealing

### Stealing Tasks or Continuations?

- Under a fork-join paradigm, there are two options to enqueue and steal

  - Tasks
  - Continuations

    ```go
    var fib func(n int) <-chan int

    fib = func(n int) <-chan int {
      result := make(chan int)

      go func() { // <1> goroutines are tasks
        defer close(result)
        if n <= 2 {
          result <- 1
          return
        }
        result <- <-fib(n-1) + <-fib(n-2)
      }()

      return result // <2> everything after a goroutine is called is the continuation
    }

    fmt.Printf("fib(4) = %d", <-fib(4))
    ```

- **Go's work-stealing algorithm enqueues and steals continuations**
- **Stalling join**: An unrealized join point requires the thread to pause execution and go fishing for a task to steal
- Axioms

  - When creating a goroutine, it is very likely that your program will want the function in that goroutine to execute
  - It's also reasonably likely that the continuation from that goroutine will at some point want to join with that goroutine
  - It's common for the continuation to attempt a join before the goroutine has finished completing
  - If we push the continuation onto the tail of the deque

    - It's least likely to get stolen by another thread that is popping things from the head of the deque
    - It becomes very likely that we'll be able to just pick it back up when we're finished executing our goroutine thus avoiding a stall

#### Demo

##### Conventions

- Continuations to enqueue are denoted as `cont. of x`
- The dequeued continuations are converted implicitly to the next invocation of `fib()`
- Jobs are taken from work queue

##### Assumptions

- The program is executing on a hypothetical machine with two single-core processors
- One OS thread is spawned on each processor, `T1` for processor one, and `T2` for processor two
- As we walk through this example, `T1` will be flipped to T2 in an effort to provide some structure. In reality, none of this is deterministic

##### A Step-by-Step Workflow

| step | `T1` call stack | `T1` work deque   | `T2` call stack                           | `T2` work deque   |
| ---- | --------------- | ----------------- | ----------------------------------------- | ----------------- |
| 1    | main            |                   |                                           |                   |
| 2    | `fib(4)`        | cont. of main     |                                           |                   |
| 3    | `fib(4)`        |                   | cont. of main                             |                   |
| 4    | `fib(3)`        | cont. of `fib(4)` | cont. of main                             |                   |
| 5    | `fib(3)`        |                   | cont. of main (unrealized join point)     |                   |
| 5    |                 |                   | cont. of `fib(4)`                         |                   |
| 6    | `fib(2)`        | cont. of `fib(3)` | cont. of main (unrealized join point)     |                   |
| 6    |                 |                   | cont. of `fib(4)`                         |                   |
| 7    | `fib(2)`        | cont. of `fib(3)` | cont. of main (unrealized join point)     | cont. of `fib(4)` |
| 7    |                 |                   | `fib(2)`                                  |                   |
| 8    | (returns 1)     | cont. of `fib(3)` | cont. of main (unrealized join point)     | cont. of `fib(4)` |
| 8    |                 |                   | `fib(2)`                                  |                   |
| 9    | (returns 1)     | cont. of `fib(3)` | cont. of main (unrealized join point)     | cont. of `fib(4)` |
| 9    |                 |                   | (returns 1)                               |                   |
| 10   | `fib(1)`        |                   | cont. of main (unrealized join point)     | cont. of `fib(4)` |
| 10   |                 |                   | (returns 1)                               |                   |
| 11   | `fib(1)`        |                   | cont. of main (unrealized join point)     |                   |
| 11   |                 |                   | cont. of `fib(4)` (unrealized join point) |                   |
| 12   | (returns 2)     |                   | cont. of main (unrealized join point)     |                   |
| 12   |                 |                   | cont. of `fib(4)` (unrealized join point) |                   |
| 13   |                 |                   | cont. of main (unrealized join point)     |                   |
| 13   |                 |                   | (returns 3)                               |                   |
| 14   |                 |                   | (prints 3)                                |                   |

> comments
> step 11.2 `fib(4)` is unrealized due to `fib(3)` is being worked by `T1`

> **The runtime on a single thread using goroutines is the same as if we had just used functions**

- **Conclusion**: Stealing continuations are considered to be theoretically superior to stealing tasks, and therefore it is best to queue the continuation and not the goroutines

- **Benefits**

  |                    | Continuation | Tasks        |
  | ------------------ | ------------ | ------------ |
  | Queue size         | Bounded      | Unbounded    |
  | Order of Execution | Serial       | Out of Order |
  | Join Point         | Nonstalling  | Stalling     |

#### Go's Scheduler

- **3 main concepts**

  | Concept | Description                                        |
  | ------- | -------------------------------------------------- |
  | `G`     | A goroutine                                        |
  | `M`     | An OS thread (a.k.a. a machine in the source code) |
  | `P`     | A context (a.k.a. a processor in the source code)  |

- **Relationship**: In Go's runtime, `M`s are started, which then host `P`s, which then schedule and host `G`s

  ![relationsip of `M`, `P` and `G`](./images/relationship-of-MPG.png)

- The `GOMAXPROCS` setting controls how many contexts are available for use by the runtime
  - The default setting is for there to be one context per logical CPU on the host machine
- There may be more or less OS threads than cores to help Go's runtime manage things like garbage collection and goroutines
- **One very important guarantee** in the runtime: there will always be at least enough OS threads available to handle hosting every context

- Blocking handling

  - For a blocked OS thread, Go would dissociate the context from it so that the context can be handed off to another unblocked OS thread
    > Normally, if any of the goroutines were blocked either by input/output or by making a system call outside of Go's runtime. The OS thread that hosts the goroutine would also be blocked and would be unable to make progress or host any other goroutines.
  - Once the hosting goroutines is unblocked(?? done), the host OS thread attempts to steal back a context from one of the other OS threads so that it can continue executing the previously blocked goroutine
    > In case of no contexts for stealing, the unblocked goroutine will be placed in _global context_, which will be handled by extra steps added to the work-stealing algorithm

- Global context
  - Periodically checked by native contexts to see if there are any goroutines there
  - Will be first option (w.r.t other OS threads' contexts) to steal work for a context with empty work queue

## Presenting All of This to the Developer

Developers just needs known the `go` keyword, and the rest is handled by the smart Go runtime
