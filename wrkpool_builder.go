package wrkpool

import (
	"time"
)

type wrkPoolBuilder struct {
	workers       int
	workersMin    int
	workerTimeout time.Duration
	jobsCapacity  int
	recover       func(interface{})
}

// Builder will return a new instance of a worker pool builder.
func Builder() *wrkPoolBuilder {
	return &wrkPoolBuilder{}
}

func (b *wrkPoolBuilder) Workers(value int) *wrkPoolBuilder {
	b.workers = value
	return b
}

// WorkersMin define the amount of workers that will be kept up even with no job
// to be performed.
//
// The second param, timeout, is used to timeout a job that
func (b *wrkPoolBuilder) WorkersMin(value int, timeout time.Duration) *wrkPoolBuilder {
	b.workersMin = value
	b.workerTimeout = timeout
	return b
}

func (b *wrkPoolBuilder) JobsCapacity(value int) *wrkPoolBuilder {
	b.jobsCapacity = value
	return b
}

func (b *wrkPoolBuilder) Recover(value func(interface{})) *wrkPoolBuilder {
	b.recover = value
	return b
}

func (b *wrkPoolBuilder) Build() *wrkPool {
	r := &wrkPool{
		pool:              make(chan worker, b.workers),
		jobs:              make(chan Job, b.jobsCapacity),
		workersMin:        int32(b.workersMin),
		workersMinTimeout: b.workerTimeout,
		Recover:           b.recover,
	}
	for i := 0; i < b.workersMin; i++ {
		r.startWorker()
	}
	return r
}
