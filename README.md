# Puddle

Puddle is a minimalistic worker pool library. It's not intended to be performant
or api-rich, just as a learning experiment.

## Architecture

The architecture taken was the simplest possible:
- A `sync.WaitGroup` to keep track of running workers
- An `[]func() error` in-memory queue of tasks to throttle from from when the
max pool size has been reached.

In order to provide some observability into the status of the pool some
functions such as `RunningWorkers()` or `FailedTasks()` have been added.

## Digging deeper

Some things that could be interesting to look into:

- How could this be implemented with channels?
- Is there a performance difference between using channels and waitgroups/mutexes?
- How could we benchmark the library?
- How could different algorithms be implemented; eager vs lazy worker initialisation?