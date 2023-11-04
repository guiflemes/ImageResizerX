package resizer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcquireWorker(t *testing.T) {
	assert := assert.New(t)
	maxWorks := 5
	pool := NewImagePool(maxWorks)

	for i := 0; i < maxWorks; i++ {
		pool.AcquireWorker()
	}

	assert.Equal(maxWorks, len(pool.worker))
	assert.NotPanics(func() {
		<-pool.worker
	})

}

func TestReleaseWorker(t *testing.T) {
	assert := assert.New(t)
	maxWorks := 3
	pool := NewImagePool(maxWorks)

	for i := 0; i < maxWorks; i++ {
		pool.AcquireWorker()
	}

	for i := 0; i < maxWorks; i++ {
		pool.ReleaseWorker()
	}

	assert.Equal(0, len(pool.worker))
}

func TestRunTask(t *testing.T) {
	assert := assert.New(t)
	maxWorks := 3
	sucessCount := 0
	pool := NewImagePool(maxWorks)

	var wg sync.WaitGroup

	for i := 0; i < maxWorks; i++ {
		wg.Add(1)
		pool.RunTask(func() {
			sucessCount++
			wg.Done()
		})
	}

	wg.Wait()

	assert.Equal(maxWorks, sucessCount)
	assert.Equal(0, len(pool.worker))
}
