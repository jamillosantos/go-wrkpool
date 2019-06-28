package wrkpool

import (
	"sync"
	"sync/atomic"
	"time"
)

// Job represents the function that will be executed when the pool
type Job func()

type worker func(*wrkPool)

// Pool
type wrkPool struct {
	pool              chan worker
	poolMutex         sync.Mutex
	jobs              chan Job
	JobsCount         int64
	WorkersWorking    int32
	WorkersCount      int32
	workersMin        int32
	workersMinTimeout time.Duration
	Recover           func(data interface{})
	wg                sync.WaitGroup
}

func (p *wrkPool) reportPanic(data interface{}) {
	// If a recover method was found.
	if p.Recover != nil {
		p.Recover(data)
	}
}

func (p *wrkPool) incWorkerWorking() {
	atomic.AddInt32(&p.WorkersWorking, 1)
}

func (p *wrkPool) decWorkerWorking() {
	atomic.AddInt32(&p.WorkersWorking, -1)
}

// startWorker initializes a worker, starting a new goroutine with a
// workerMethod.
func (p *wrkPool) startWorker() {
	// Increment the `p.WorkersCount`
	p.wg.Add(1)
	atomic.AddInt32(&p.WorkersCount, 1)
	go p.worker()
}

// Do enqueues a job to be processed by the worker.
//
// As the jobs are being added, the pool will allocate
func (p *wrkPool) Do(job Job) {
	func() {
		p.poolMutex.Lock()
		defer p.poolMutex.Unlock()

		poolCapacity := int32(cap(p.pool))
		if p.WorkersCount < poolCapacity && len(p.pool) == 0 {
			p.startWorker()
		}
	}()

	p.jobs <- job
}

func (p *wrkPool) Close() {
	close(p.jobs)
	p.wg.Wait()
}

// workerMethod is the method that will be started as a goroutine. Designed to
// never end until the pool is closed, this method is blocking.
//
// It protects the goroutine against a possible crash through panic defering a
// recover function that will call the `pool.reportPanic` if anything bad
// happens.
func (p *wrkPool) worker() {
	defer func() {
		p.wg.Add(-1) // `wg.Done` without one more pc jump.
		atomic.AddInt32(&p.WorkersCount, -1)
	}()

	var (
		job Job
		ok  bool
	)
	for {
		// If there is a minimum workers allocation needed...
		if p.workersMin > 0 {
			select {
			case job, ok = <-p.jobs:
				// This case will be treated further
			case <-time.After(p.workersMinTimeout):
				// If there is more workers than needed... Good bye worker.
				if p.WorkersCount > p.workersMin {
					return
				}
				// The worker must be kept alive to meet the `workersMin`
				// configuration.
				continue
			}
		} else {
			// No workers min needed. So, no need to add more complexity than it.
			job, ok = <-p.jobs
		}
		if ok {
			atomic.AddInt64(&p.JobsCount, 1)
		} else {
			return
		}
		func() {
			defer func() {
				// Decrement working
				p.decWorkerWorking()
				if r := recover(); r != nil {
					p.reportPanic(r)
				}
			}()

			// Increment working
			p.incWorkerWorking()
			job()
		}()
	}
}
