/*
 * @Author: jinde.zgm
 * @Date: 2020-05-14 16:31:35
 * @Description: coroutine implement.
 */

package gopool

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// coroutine define routine which can run.
type coroutine struct {
	rc     chan Routine // Routine channel.
	pool   *pool        // Pool pointer.
	active time.Time    // Last active time.
	next   *coroutine
}

// run is coroutine run routine.
func (c *coroutine) run() {
	// Coroutine is exist.
	defer c.pool.wg.Done()

	// Wait for routine.
	for r := range c.rc {
		// Exit signal?
		if r == nil {
			return
		}

		// Call routine.
		r(c.pool.ctx)

		// Put coroutine to idles.
		c.pool.pushIdle(c)
	}
}

// loadNext load next coroutine as unsafe.Pointer.
func (c *coroutine) loadNext() unsafe.Pointer {
	return atomicLoadCoroutine(&c.next)
}

// storeNext store unsafe.Pointer to next.
func (c *coroutine) storeNext(next unsafe.Pointer) {
	atomicStoreCoroutine(&c.next, next)
}

// atomicLoadCoroutine atomic load coroutine pointer as unsafe.Pointer
func atomicLoadCoroutine(p **coroutine) unsafe.Pointer {
	return atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(p)))
}

//  atomicStoreCoroutine atomic store unsafe.Pointer to coroutine pointer
func atomicStoreCoroutine(p **coroutine, v unsafe.Pointer) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(p)), v)
}
