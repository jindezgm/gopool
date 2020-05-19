/*
 * @Author: jinde.zgm
 * @Date: 2020-05-15 10:43:33
 * @Description:
 */

package gopool

import (
	"context"
	"testing"
	"time"
)

// func TestPool(t *testing.T) {
// 	p, err := New(WithCapacity(4))
// 	if nil != err {
// 		t.Fatal(err)
// 	}

// 	rand.Seed(time.Now().UnixNano())

// 	wg := sync.WaitGroup{}
// 	wg.Add(10)
// 	for i := 0; i < 10; i++ {
// 		p.Go(func(context.Context) {
// 			time.Sleep(time.Duration(rand.Int63n(time.Millisecond.Nanoseconds() * 100)))
// 			fmt.Println("hello world")
// 			wg.Done()
// 		})
// 	}

// 	wg.Wait()
// 	p.Close()
// }

func TestNew(t *testing.T) {
	cc := new(coroutine)
	dd := new(coroutine)
	cc.next = dd
	c1 := cc
	c1, c1.next = c1.next, nil
	if cc.next != nil || c1 != dd {
		t.Fatal("!!!!!")
	}

	if _, err := New(WithCapacity(-1)); ErrInvalidPoolSize != err {
		t.Fatal(err)
	}

	p, err := New(WithName("test"))
	if nil != err {
		t.Fatal(err)
	}

	if p.Name() != "test" {
		t.Fatal("pool name", p.Name(), "is not test")
	}

	if p, err := New(WithContext(context.Background()), WithCapacity(10)); nil != err {
		t.Fatal(err)
	} else if p.Capacity() != 10 {
		t.Fatal(p.Capacity(), "capacity is not 10")
	}
}

func TestPushIdle(t *testing.T) {
	pp, _ := New()
	p := pp.(*pool)
	c := new(coroutine)
	p.pushIdle(c)
	if p.popIdle() != c {
		t.Fatal("push or pop idle failed")
	}
}

func TestClean(t *testing.T) {
	p := &pool{Options: &Options{IdleTimeout: 500 * time.Millisecond}}

	p.wg.Add(10)
	p.count = int64(10) << 32
	for i := 0; i < 10; i++ {
		p.pushIdle(&coroutine{pool: p, rc: make(chan Routine, 1)})
		time.Sleep(100 * time.Millisecond)
	}
	if p.count>>32 != 0 {
		t.Fatal("invalid running count")
	}

	p.clean(time.Now())
	var idles int
	for c := p.idles; c != nil; c = c.next {
		idles++
	}
	if idles != 4 {
		t.Fatal("clean failed", idles)
	}
}

func TestGoNonblock(t *testing.T) {
	p, _ := New(WithCapacity(2))
	c := make(chan struct{})
	for i := 0; i < 2; i++ {
		if err := p.Go(func(context.Context) {
			<-c
		}); nil != err {
			t.Fatal(err)
		}
	}

	if err := p.GoNonblock(func(context.Context) {
		t.Log("go nonblock")
	}); ErrPoolFull != err {
		t.Fatal(err)
	}
}
