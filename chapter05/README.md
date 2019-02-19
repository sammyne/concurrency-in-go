# Concurrency at Scale

## Error Propagation

- **Motivation**: it's both easy for something to go wrong in your system, and difficult to understand why it happened

### Errors

- Errors indicate that your system has entered a state in which it cannot fulfill an operation that a user either explicitly or implicitly requested. The critical information it relays (should be added manually) include

  - What happened
  - When and where it occurred

    > Errors should contain
    >
    > - **always** a complete **stack trace**
    >   - starting with how the call was initiated
    >   - ending with where the error was instantiated
    >   - not contained in the error message (more on this in a bit)
    > - information regarding its running context
    > - the time on the machine the error was instantiated on, in UTC

  - A about-one-line friendly user-facing message
  - How the user can get more information
    > Errors that are presented to users should provide an **ID** that can be cross-referenced to a corresponding log that displays the full information of the error
    >
    > - time the error occurred (not the time the error was logged)
    > - the stack trace

- 2 Categories
  - Bugs: errors that you have not customized to your system, or "raw" errors -- your known edge cases
  - Known edge cases (e.g., broken network connections, failed disk writes, etc.)
- At the boundaries of each component, all incoming errors must be wrapped in a well-formed error for the component our code is within
  > - It is only necessary to wrap errors in this fashion at your own module boundaries—public functions/methods—or when your code can add valuable context
  > - Well-formed incoming errors helps to control how errors escape our module.
- All errors should be logged with as much information as is available
- A well-formed error received by user-facing code can be simply logged and displayed to user
- When malformed errors, or bugs, are propagated up to the user, we should **also log the error, but then display a friendly message to the user stating something unexpected has happened**
  > - If we support automatic error reporting in our system, the error should be reported back as a bug
  > - Otherwise, we might suggest the user file a bug report

## Timeouts and Cancellation

### Why Timeouts

- System saturation (i.e., its ability to process requests is at capacity)
  - general guidelines for when to time out
    - If the request is unlikely to be repeated when it is timed out
    - If you don't have the resources to store the requests
    - If the need for the request, or the data it's sending, will go stale
- Stale data
  - data has a window within which it must be processed before more relevant data is available, or the need to process the data has expired.
  - managed by `context.Context` created with
    - `context.WithDeadline`, or `context.WithTimeout` for the timing window known beforehand
    - `context.WithCancel`, otherwise
- Attempting to prevent deadlocks
  - The timeout period's purpose is only to prevent deadlock, and so it only needs to be short enough that a deadlocked system will unblock in a reasonable amount of time for your use case
  - It is preferable to chance a livelock and fix that as time permits, than for a deadlock to occur and have a system recoverable only by restart

### Why Cancellation

- Timeouts
- User intervention
- Parent cancellation
- Replicated requests
  - Scenario: data are sent to multiple concurrent processes in an attempt to get a faster response from one of them

### Job after Cacellation

#### The preemptability of a concurrent process

##### non-preemptable job 1

```go
var value interface{}
select {
case <-done:
  return
case value = <-valueStream:
}

// next line is non-preemptable
// which will block the cancellation for a long time during execution
result := reallyLongCalculation(value)

select {
case <-done:
  return
case resultStream <- result:
}
```

##### non-preemptable job 2

```go
reallyLongCalculation := func(
  done <-chan interface{},
  value interface{},
) interface{} {
  intermediateResult := longCalculation(value)
  select {
  case <-done:
    return nil
  default:
  }
}

return longCaluclation(intermediateResult)
```

> `longCalculation()` may still block cancellation during execution

##### ok job as `preemptability/main.go`

##### 2 steps to achieve preemptability

- define the period within which our concurrent process is preemptable
- ensure that any functionality that takes more time than this period is itself preemptable

#### Problems

##### quit from shared state

**solution**: If possible, build up intermediate results in-memory and then modify state as quickly as possible

- wrong way

```go
result := add(1, 2, 3)
writeTallyToState(result)
result = add(result, 4, 5, 6)
writeTallyToState(result)
result = add(result, 7, 8, 9)
writeTallyToState(result)
```

- good way

```go
result := add(1, 2, 3, 4, 5, 6, 7, 8, 9)
writeTallyToState(result)
```

##### dulplicated messages

- **use case**
