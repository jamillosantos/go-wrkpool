package wrkpool_test

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jamillosantos/go-wrkpool"
)

func workWithN() {
	time.Sleep(time.Second / 10)
}

func BenchmarkWorkWithNGoroutines(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go workWithN()
	}
}

func BenchmarkWorkWithNWithPool(b *testing.B) {
	pool := wrkpool.Builder().
		Workers(int(math.Max(float64(b.N/10), 1))). // Get 1/10th of workers
		JobsCapacity(int(math.Max(float64(b.N/10), 1))).Build() // Get 1/10th of jobs capacity
	for i := 0; i < b.N; i++ {
		pool.Do(workWithN)
	}
	pool.Close()
}

func BenchmarkSumOfNumbersGoroutines(b *testing.B) {
	var n int32
	var wg sync.WaitGroup
	sumN := func() {
		defer wg.Done()
		atomic.AddInt32(&n, 1)
	}

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go sumN()
	}
	wg.Wait()

	assert.Equal(b, n, int32(b.N))
}

func BenchmarkSumOfNumbersWithPool(b *testing.B) {
	var n int32
	sumN := func() {
		atomic.AddInt32(&n, 1)
	}

	// Get 1/10th of workers
	pool := wrkpool.Builder().
		Workers(int(math.Max(float64(b.N/10), 1))).
		JobsCapacity(int(math.Max(float64(b.N/10), 1))).Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Do(sumN)
	}
	pool.Close()

	assert.Equal(b, n, int32(b.N))
}

func TestWrkPoolDoWithPanic(t *testing.T) {
	var (
		n, panics int32
	)
	job := func() {
		if atomic.AddInt32(&n, 1)%4 == 0 {
			panic("controlled panic")
		}
		time.Sleep(time.Second / 3)
	}

	pool := wrkpool.Builder().
		Workers(3).
		JobsCapacity(10).
		Recover(func(data interface{}) {
			if data != "controlled panic" {
				assert.Equal(t, "controlled panic", data, "the panic captured by the recover function was not the expected one")
				t.Fail()
			}
			atomic.AddInt32(&panics, 1)
		}).
		Build()
	for i := 0; i < 10; i++ {
		pool.Do(job)
	}
	time.Sleep(time.Second / 10)
	assert.Equal(t, int32(3), pool.WorkersCount)
	pool.Close()

	assert.Equal(t, int32(2), panics, "the number of panics is not right")
	assert.Equal(t, int32(10), n, "the result of the entire job was not computed right")
}

func TestWrkPoolDoWithNoMinimumWorkers(t *testing.T) {
	var n int32
	job := func() {
		atomic.AddInt32(&n, 1)
		time.Sleep(time.Second / 3)
	}

	pool := wrkpool.Builder().
		Workers(3).
		JobsCapacity(10).Build()
	for i := 0; i < 10; i++ {
		pool.Do(job)
	}
	time.Sleep(time.Second / 10)
	assert.Equal(t, int32(3), pool.WorkersCount)
	pool.Close()

	assert.Equal(t, int32(10), n, "the result of the entire job was not computed right")
}

func TestWrkPoolWithMinWorkers(t *testing.T) {
	var n int32
	jobShort := func() {
		atomic.AddInt32(&n, 1)
		time.Sleep(time.Second / 3)
	}
	jobLong := func() {
		atomic.AddInt32(&n, 1)
		time.Sleep(time.Second)
	}

	pool := wrkpool.Builder().
		Workers(10).
		JobsCapacity(10).
		WorkersMin(3, time.Second).
		Build()
	for i := 0; i < 10; i++ {
		pool.Do(jobShort)
	}
	for i := 0; i < 2; i++ {
		pool.Do(jobLong)
	}
	time.Sleep(time.Second)
	assert.Equal(t, int32(10), pool.WorkersCount)
	time.Sleep(time.Second)
	assert.Equal(t, int32(3), pool.WorkersCount)
	pool.Close()

	assert.Equal(t, int32(12), n, "the result of the entire job was not computed right")
}
