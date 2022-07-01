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
	maxWorkers      uint64
	mt              *sync.Mutex
	wg              *sync.WaitGroup
	tasksInProgress uint64
	tasksFinished   uint64
	tasksSuccessful uint64
	tasksFailed     uint64

	queuedTasks []func() error
}

func NewPool(maxWorkers uint64) Pool {
	return &pool{
		maxWorkers:      maxWorkers,
		wg:              &sync.WaitGroup{},
		mt:              &sync.Mutex{},
		tasksInProgress: 0,
		tasksFinished:   0,
	}
}

func (p *pool) Do(f func() error) {
	if p.tasksInProgress >= p.maxWorkers {
		p.pushTask(f)
		return
	}

	p.trackInProgress()
	go p.runTask(f)
}

func (p *pool) trackInProgress() {
	p.mt.Lock()
	p.tasksInProgress++
	p.mt.Unlock()

	p.wg.Add(1)
}

func (p *pool) popTask() func() error {
	var f func() error

	p.mt.Lock()
	f, p.queuedTasks = p.queuedTasks[0], p.queuedTasks[1:]
	p.mt.Unlock()

	return f
}

func (p *pool) pushTask(f func() error) {
	p.mt.Lock()
	p.queuedTasks = append(p.queuedTasks, f)
	p.mt.Unlock()
}

func (p *pool) trackResult(err error) {
	p.mt.Lock()
	if err != nil {
		p.tasksFailed++
	} else {
		p.tasksSuccessful++
	}

	p.tasksFinished++
	p.tasksInProgress--
	p.mt.Unlock()
}

func (p *pool) runTask(f func() error) {
	defer p.wg.Done()

	err := f()
	p.trackResult(err)

	if len(p.queuedTasks) > 0 {
		f := p.popTask()

		p.trackInProgress()
		go p.runTask(f)
	}
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
