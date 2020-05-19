/*
 * @Author: jinde.zgm
 * @Date: 2020-05-17 22:08:20
 * @Description:
 */

package gopool

import (
	"context"
	"testing"
	"unsafe"
)

func TestAtomicLoadCoroutine(t *testing.T) {
	c := new(coroutine)
	c1 := atomicLoadCoroutine(&c)
	if c != (*coroutine)(c1) {
		t.Fatal("atomicLoadCoroutine failed")
	}
}

func TestAtomicStoreCoroutine(t *testing.T) {
	c := new(coroutine)
	var c1 *coroutine
	if atomicStoreCoroutine(&c1, unsafe.Pointer(c)); c1 != c {
		t.Fatal("atomicStoreCoroutine failed")
	}
}

func TestCoroutine(t *testing.T) {
	p := &pool{}
	c := &coroutine{
		pool: p,
		rc:   make(chan Routine, 1),
	}

	c1 := new(coroutine)
	c.storeNext(unsafe.Pointer(c1))
	if c.loadNext() != unsafe.Pointer(c1) {
		t.Fatal("storeNext or loadNext failed")
	}

	p.wg.Add(1)
	go c.run()

	x := make(chan struct{})
	r := func(context.Context) { x <- struct{}{} }
	c.rc <- r

	<-x

	c.rc <- nil
	p.wg.Wait()
}
