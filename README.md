# wrkpool [![CircleCI](https://circleci.com/gh/jamillosantos/go-wrkpool.svg?style=shield)](https://circleci.com/gh/lab259/argo) [![Go Report Card](https://goreportcard.com/badge/github.com/jamillosantos/go-wrkpool)](https://goreportcard.com/report/github.com/jamillosantos/go-wrkpool) [![codecov](https://codecov.io/gh/jamillosantos/go-wrkpool/branch/master/graph/badge.svg)](https://codecov.io/gh/jamillosantos/go-wrkpool)

## What is `wrkpool`

`wrkpool` is a library that provides a way of setting a pool with a bunch of
pre-allocated goroutines that will execute jobs. Each `Job` is a `func` with no
parameter.

The library provides a way to limit the number of ongoing workers. So, we can
throttle the job preventing that peak of "jobs" requests can be eased and don't
overwhelm your service. 

## Getting started

First things first, you must import the library by adding:

```go
import "github.com/jamillosantos/go-wrkpool"
```

Now, we need 2 things:

1. The pool of workers;
2. Start the jobs;

```go
package main

import (
	"fmt"

	"github.com/jamillosantos/go-wrkpool"
)

func main() {
	// 1st step
	// --------
	pool := wrkpool.Builder().
		Workers(10).         // We will have 10 workers at maximum.
		JobsCapacity(1000).  // The number of Jobs that can be received in our queue (a go channel).
		Build()

	// 2nd step
	// --------
	for i := 0; i < 1000; i++ {
		pool.Do(func () {
			// The work that will be run in parallel.
		})
	}
	pool.Close() // Finishes the pool blocking until all jobs be finished.
	fmt.Println("The end.")
}
```

Note that, in the example above we will only process 10 jobs concurrently. But,
we can securely enqueue 1000 jobs without locking the "main thread".

#### What if my goroutine panics?

For that, a `Recover` method was introduced to address those cases. Even if you
do not set one, the `panic` will be gracefully dismissed and the goroutine will
continue ready for the next job.

```go
pool := wrkpool.Builder().
	Recover(func(data interface{}) {
		log.Println("PANIC: ", data)
	}).
	Build()
```

## License

MIT.