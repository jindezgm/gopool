/*
 * @Author: jinde.zgm
 * @Date: 2020-05-15 11:21:47
 * @Description:
 */

package gopool

import (
	"testing"
	"time"

	"context"
)

func TestOption(t *testing.T) {
	opts := &Options{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if WithName("test")(opts); opts.Name != "test" {
		t.Fatal("WithName failed")
	}

	if WithCapacity(12)(opts); opts.Capacity != 12 {
		t.Fatal("WithCapacity failed")
	}

	if WithContext(ctx)(opts); opts.Context != ctx {
		t.Fatal("WithContext failed")
	}

	if WithIdleTimeout(1234)(opts); opts.IdleTimeout != 1234 {
		t.Fatal("WithIdleTimeout failed", opts.IdleTimeout, 1234)
	}
}

func TestCheckOptions(t *testing.T) {
	opts := &Options{}
	WithCapacity(-1)(opts)
	if err := opts.Check(); ErrInvalidPoolSize != err {
		t.Fatal(err)
	}

	opts.Capacity = 0
	WithIdleTimeout(-1)(opts)
	if err := opts.Check(); ErrInvalidIdleTimeout != err {
		t.Fatal(err)
	}

	opts.IdleTimeout = 0
	if err := opts.Check(); nil != err {
		t.Fatal(err)
	} else if DefaultIdleTimeout != opts.IdleTimeout || DefaultPoolSize != opts.Capacity {
		t.Fatal("not default options", opts)
	}
}

func TestNewOptions(t *testing.T) {
	if _, err := NewOptions(WithCapacity(1), WithIdleTimeout(time.Second)); nil != err {
		t.Fatal(err)
	}

	if _, err := NewOptions(WithCapacity(-1)); ErrInvalidPoolSize != err {
		t.Fatal(err)
	}
}
