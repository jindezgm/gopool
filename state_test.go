/*
 * @Author: jinde.zgm
 * @Date: 2020-05-17 22:24:27
 * @Description:
 */

package gopool

import "testing"

func TestState(t *testing.T) {
	var s state

	if !s.set(stateClosed) {
		t.Fatal("set failed")
	}

	if s.is(stateRunning) {
		t.Fatal("closed state but running?")
	}
}
