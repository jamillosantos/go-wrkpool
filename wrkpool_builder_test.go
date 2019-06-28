package wrkpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrkPoolBuilder_WrkPoolBuilder(t *testing.T) {
	builder := Builder()
	assert.NotNil(t, builder)
}

func TestWrkPoolBuilder_Workers(t *testing.T) {
	builder := Builder()
	assert.Equal(t, builder.workers, 0)
	builder.Workers(10)
	assert.Equal(t, builder.workers, 10)
	builder.Workers(50)
	assert.Equal(t, builder.workers, 50)
}

func TestWrkPoolBuilder_WorkersMin(t *testing.T) {
	builder := Builder()
	assert.Equal(t, builder.workersMin, 0)
	assert.Equal(t, builder.workerTimeout, time.Duration(0))
	builder.WorkersMin(10, time.Second*2)
	assert.Equal(t, builder.workersMin, 10)
	builder.WorkersMin(50, time.Second*3)
	assert.Equal(t, builder.workersMin, 50)
}

func TestWrkPoolBuilder_JobsCapacity(t *testing.T) {
	builder := Builder()
	assert.Equal(t, builder.jobsCapacity, 0)
	builder.JobsCapacity(10)
	assert.Equal(t, builder.jobsCapacity, 10)
	builder.JobsCapacity(50)
	assert.Equal(t, builder.jobsCapacity, 50)
}

func TestWrkPoolBuilder_Build(t *testing.T) {
	called := false
	r := func(interface{}) {
		called = true
	}
	pool := Builder().
		Workers(1).
		WorkersMin(2, time.Second*4).
		JobsCapacity(3).
		Recover(r).
		Build()
	assert.Equal(t, 1, cap(pool.pool))
	assert.Equal(t, int32(2), pool.workersMin)
	assert.Equal(t, time.Second*4, pool.workersMinTimeout)
	assert.Equal(t, 3, cap(pool.jobs))
	assert.Equal(t, 3, cap(pool.jobs))
	// Call the recovery to ensure it is `r`.
	pool.Recover(nil)
	assert.Equal(t, true, called, "the recovery method is not working properly")
}
