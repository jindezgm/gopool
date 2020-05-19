/*
 * @Author: jinde.zgm
 * @Date: 2020-05-16 15:57:08
 * @Description: coroutine pool state.
 */

package gopool

import "sync/atomic"

const (
	stateRunning state = iota // Running state
	stateClosed               // Closed state
)

// Define state.
type state int32

// set state.
func (s *state) set(v state) bool {
	return atomic.CompareAndSwapInt32((*int32)(s), atomic.LoadInt32((*int32)(s)), int32(v))
}

// is the specific state?
func (s *state) is(v state) bool {
	return atomic.LoadInt32((*int32)(s)) == int32(v)
}
