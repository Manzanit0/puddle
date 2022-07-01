package puddle

import "sync"

type Pool interface {
	Do(func() error)
	Wait() error
	IsDone() bool
	RunningWorkers() uint64
	SuccessfulTasks() uint64
	FailedTasks() uint64
}

type pool struct {
	maxWorkers      int
	mt              *sync.Mutex
	wg              *sync.WaitGroup
	tasksInProgress uint64
	tasksFinished   uint64
	tasksSuccessful uint64
	tasksFailed     uint64
}

func NewPool(maxWorkers int) Pool {
	return &pool{
		maxWorkers:      maxWorkers,
		wg:              &sync.WaitGroup{},
		mt:              &sync.Mutex{},
		tasksInProgress: 0,
		tasksFinished:   0,
	}
}

func (p *pool) Do(f func() error) {
	p.mt.Lock()
	p.tasksInProgress++
	p.mt.Unlock()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		err := f()

		p.mt.Lock()
		defer p.mt.Unlock()

		if err != nil {
			p.tasksFailed++
		} else {
			p.tasksSuccessful++
		}

		p.tasksFinished++
		p.tasksInProgress--
	}()
}

func (p *pool) Wait() error {
	p.wg.Wait()
	return nil
}

func (p *pool) IsDone() bool {
	return p.tasksInProgress == 0
}

func (p *pool) SuccessfulTasks() uint64 {
	return p.tasksSuccessful
}

func (p *pool) FailedTasks() uint64 {
	return p.tasksFailed
}

func (p *pool) RunningWorkers() uint64 {
	return p.tasksInProgress
}
