package resizer

type ImagePool struct {
	worker     chan struct{}
	maxWorkers int
}

func NewImagePool(maxWorkers int) *ImagePool {
	return &ImagePool{
		worker:     make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

func (pool *ImagePool) AcquireWorker() {
	pool.worker <- struct{}{}
}

func (pool *ImagePool) ReleaseWorker() {
	<-pool.worker
}

func (pool *ImagePool) RunTask(task func()) {
	go func() {
		pool.AcquireWorker()
		defer pool.ReleaseWorker()
		task()
	}()
}
