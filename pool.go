/*
 * @Author: jinde.zgm
 * @Date: 2020-05-14 16:04:44
 * @Description: coroutine pool implement.
 */

package gopool

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Pool define coroutine pool interface.
type Pool interface {
	Name() string             // Get coroutine pool name.
	Capacity() int32          // Get coroutine pool capacity.
	Tune(size int32)          // Change coroutine pool capacity.
	Status() Status           // Get coroutine pool status.
	Go(Routine) error         // Go the specific routine.
	GoNonblock(Routine) error // Nonblock go the specific routine.
	Close()                   // Close coroutine pool.
}

// Routine is function called by the coroutine pool
type Routine func(context.Context)

// Status is cotoutine pool status
type Status struct {
	Runnings int32 // Running coroutines
	Idles    int32 // Idle coroutines.
}

// pool implement Pool interface.
type pool struct {
	*Options                    // Coroutine options.
	ctx      context.Context    // Coroutine context.
	cancel   context.CancelFunc // Coroutine context cancel function.
	state    state              // Coroutine pool state.
	done     chan struct{}      // Coroutine pool closed flag.
	cache    sync.Pool          // Used to speeds up the obtainment of the an usable coroutine.
	wg       sync.WaitGroup     // Used to wait for coroutine exiting.
	count    int64              // The upper 32 bits are running coroutines count and the lower 32 bits are the total coroutines count
	idles    *coroutine         // Idle coroutines.
}

var _ Pool = &pool{}
var (
	// Used for cleaning timeout idle coroutine.
	cleaning = &coroutine{}

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("coroutine pool has been closed")

	// ErrPoolFull will be returned when the pool is full and no coroutine available.
	ErrPoolFull = errors.New("too many coroutines running")
)

// New create Pool.
func New(options ...Option) (Pool, error) {
	// Create options.
	opts, err := NewOptions(options...)
	if nil != err {
		return nil, err
	}

	// Create pool.
	p := &pool{
		Options: opts,
		done:    make(chan struct{}),
	}

	// Set cache new function.
	p.cache.New = func() interface{} {
		c := &coroutine{
			rc:   make(chan Routine, 1),
			pool: p,
		}
		return c
	}

	// Create pool context.
	if nil == opts.Context {
		p.ctx, p.cancel = context.WithCancel(context.Background())
	} else {
		p.ctx, p.cancel = context.WithCancel(opts.Context)
	}

	go p.run()

	return p, nil
}

// Name implement Pool.Name()
func (p *pool) Name() string {
	return p.Options.Name
}

// Capacity implement Pool.Capacity()
func (p *pool) Capacity() int32 {
	return atomic.LoadInt32(&p.Options.Capacity)
}

// Tune implement Pool.Tune()
func (p *pool) Tune(size int32) {
	if size < 0 || p.Capacity() == size {
		return
	}

	atomic.StoreInt32(&p.Options.Capacity, size)
}

// Status implement Pool.Status()
func (p *pool) Status() Status {
	count := atomic.LoadInt64(&p.count)
	runninngs := int32(count >> 32)
	idles := int32(count) - runninngs
	return Status{Runnings: runninngs, Idles: idles}
}

// Close implement Pool.Close()
func (p *pool) Close() {
	if p.state.set(stateClosed) {
		p.cancel()
		<-p.done
	}
}

// Go implement Pool.Go()
func (p *pool) Go(r Routine) error {
	return p.goRoutine(r, false)
}

// GoNonblock implement Pool.GoNonblock()
func (p *pool) GoNonblock(r Routine) error {
	return p.goRoutine(r, true)
}

// goContext go routine with specific context.
func (p *pool) goRoutine(r Routine, nonblocking bool) error {
	if !p.state.is(stateRunning) {
		return ErrPoolClosed
	}

	var c *coroutine
	for c = p.popIdle(); nil == c; c = p.popIdle() {
		if count := atomic.LoadInt64(&p.count); int32(count) >= p.Capacity() {
			if nonblocking {
				return ErrPoolFull
			}

			runtime.Gosched()
		} else if atomic.CompareAndSwapInt64(&p.count, count, count+1) {
			c = p.cache.Get().(*coroutine)
			p.wg.Add(1)
			go c.run()
			break
		}
	}

	atomic.AddInt64(&p.count, int64(1)<<32)
	c.rc <- r
	return nil
}

// run is pool core routine to schedule coroutine.
func (p *pool) run() {
	// Close pool if exit run().
	defer p.close()

	// Create ticker to period clean timeout coroutine.
	ticker := time.NewTicker(p.IdleTimeout)
	defer ticker.Stop()

	for {
		select {
		// Period clean timeout coroutine.
		case now := <-ticker.C:
			p.clean(now)

		// Coroutine closed?
		case <-p.ctx.Done():
			return
		}
	}
}

// close all coroutine and set done flag.
func (p *pool) close() {
	for int32(atomic.LoadInt64(&p.count)) > 0 {
		if c := p.popIdle(); nil == c {
			runtime.Gosched()
		} else {
			c.rc <- nil
			atomic.AddInt64(&p.count, -1)
		}
	}

	p.wg.Wait()
	close(p.done)
}

// clean timeout coroutines.
func (p *pool) clean(now time.Time) {
	// set a special coroutine ot the queue header indicate that the pool is in a cleaning state
	p.pushIdle(cleaning)

	for from, c := cleaning, cleaning.next; nil != c; from, c = c, c.next {
		// Find the first timeout coroutine.
		if now.Sub(c.active) >= p.IdleTimeout {
			from.storeNext(nil)

			// Release all timeout coroutines.
			var count int32
			for c != nil {
				c.rc <- nil
				c, c.next = c.next, nil
				count++
			}

			// Reduce coroutine count.
			atomic.AddInt64(&p.count, -int64(count))
			break
		}
	}

	// Recover state.
	atomicStoreCoroutine(&p.idles, unsafe.Pointer(cleaning.next))
}

// pushIdle coroutine to idle queue.
func (p *pool) pushIdle(c *coroutine) {
	// Set the last active time of the coroutine.
	c.active = time.Now()
	for {
		// Get the first idle coroutine.
		head, clean := p.idleHead()
		if clean {
			// If the coroutine pool is in the cleaning state, spin until the state is recovered.
			// Because the cleaning is very fast, spin is a better method
			runtime.Gosched()
			continue
		}

		// p.idles->c.next->head
		if c.storeNext(head); p.casIdleHead(head, unsafe.Pointer(c)) {
			if c != cleaning {
				atomic.AddInt64(&p.count, int64(-1)<<32)
			}
			break
		}
	}
}

// popIdle pop the head coroutine of the idle queue.
func (p *pool) popIdle() *coroutine {
	for {
		// Get the first idle coroutine.
		head, cleaning := p.idleHead()
		if nil == head {
			return nil
		} else if cleaning {
			runtime.Gosched()
			continue
		}

		// p.idles->head.next
		c := (*coroutine)(head)
		if next := c.loadNext(); p.casIdleHead(head, next) {
			c.storeNext(nil)
			return c
		}
	}
}

// idleHead get head of the idle coroutine queue.
func (p *pool) idleHead() (unsafe.Pointer, bool) {
	head := atomicLoadCoroutine(&p.idles)
	return head, cleaning == (*coroutine)(head)
}

// casIdleHead compare head of the idle coroutine queuue and swap it.
func (p *pool) casIdleHead(o, n unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&p.idles)), o, n)
}
