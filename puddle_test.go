package puddle_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/manzanit0/puddle"
)

func TestWorkerPool(t *testing.T) {
	t.Run("when tasks are queued, they are run", func(t *testing.T) {
		m := &sync.Mutex{}
		var count int

		// Create a pool with a max of 3 workers.
		pool := puddle.NewPool(3)

		// Run 10 successful tasks.
		for i := 0; i < 10; i++ {
			pool.Do(func() error {
				m.Lock()
				defer m.Unlock()

				count++
				return nil
			})
		}

		err := pool.Wait()
		if err != nil {
			t.Fatal(err)
		}

		if count != 10 {
			t.Fatalf("expected count to be 10, got %d", count)
		}

		got := pool.SuccessfulTasks()
		if got != uint64(10) {
			t.Fatalf("expected successful tasks to be 10, got %d", got)

		}

		got = pool.FailedTasks()
		if got != uint64(0) {
			t.Fatalf("expected failed tasks to be 0, got %d", got)
		}
	})

	t.Run("when tasks fail, the failures are counted", func(t *testing.T) {
		m := &sync.Mutex{}
		var count int

		pool := puddle.NewPool(3)

		// Run 9 successful tasks.
		for i := 0; i < 9; i++ {
			pool.Do(func() error {
				m.Lock()
				defer m.Unlock()

				count++
				return nil
			})
		}

		// And a failed task.
		pool.Do(func() error {
			return fmt.Errorf("failed")
		})

		err := pool.Wait()
		if err != nil {
			t.Fatal(err)
		}

		if count != 9 {
			t.Fatalf("expected count to be 9, got %d", count)
		}

		got := pool.SuccessfulTasks()
		if got != uint64(9) {
			t.Fatalf("expected successful tasks to be 10, got %d", got)

		}

		got = pool.FailedTasks()
		if got != uint64(1) {
			t.Fatalf("expected failed tasks to be 0, got %d", got)
		}
	})

	t.Run("when a task is in progress, the pool provides feedback", func(t *testing.T) {
		pool := puddle.NewPool(3)

		pool.Do(func() error {
			time.Sleep(5 * time.Millisecond)
			return fmt.Errorf("failed")
		})

		pool.Do(func() error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})

		done := pool.IsDone()
		if done {
			t.Fatal("expected pool to not be done")
		}

		running := pool.RunningWorkers()
		if running != 2 {
			t.Fatalf("expected running workers to be 2, got %d", running)
		}

		err := pool.Wait()
		if err != nil {
			t.Fatal(err)
		}

		done = pool.IsDone()
		if !done {
			t.Fatal("expected pool to be done")
		}
	})

	t.Run("when more tasks than the max pool size are submitted, the concurrency is throttled", func(t *testing.T) {
		pool := puddle.NewPool(3)

		for i := 0; i < 4; i++ {
			pool.Do(func() error {
				time.Sleep(5 * time.Millisecond)
				return nil
			})
		}

		done := pool.IsDone()
		if done {
			t.Fatal("expected pool to not be done")
		}

		running := pool.RunningWorkers()
		if running != 3 {
			t.Fatalf("expected running workers to be 3, got %d", running)
		}

		err := pool.Wait()
		if err != nil {
			t.Fatal(err)
		}

		running = pool.RunningWorkers()
		if running != 0 {
			t.Fatalf("expected running workers to be 0, got %d", running)
		}

		success := pool.SuccessfulTasks()
		if success != 4 {
			t.Fatalf("expected successful tasks to be 4, got %d", success)
		}
	})
}
