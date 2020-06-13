/*
 * @Author: jinde.zgm
 * @Date: 2020-05-19 16:06:31
 * @Description: gopool example.
 */

package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jindezgm/gopool"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	p, err := gopool.New(gopool.WithCapacity(10), gopool.WithIdleTimeout(time.Second))
	if nil != err {
		panic(err)
	}

	c := make(chan [26]int)
	data := make([]bytes.Buffer, 10)
	for i := 0; i < 10; i++ {
		count := 512 + rand.Intn(512)
		for j := 0; j < count; j++ {
			data[i].WriteByte(byte('a' + rand.Int31n(26)))
		}
	}

	for i := 0; i < 10; i++ {
		p.Go(wordCount(data[i].String(), c))
	}

	var result [26]int
	for i := 0; i < 10; i++ {
		r := <-c
		for i := 0; i < 26; i++ {
			result[i] += r[i]
		}
	}

	fmt.Println("word count:")
	for i := 0; i < 26; i++ {
		fmt.Printf("%c:%d\n", byte('a'+i), result[i])
	}

	p.Close()
}

func wordCount(data string, resultc chan [26]int) gopool.Routine {
	return func(context.Context) {
		var result [26]int
		for i := 0; i < len(data); i++ {
			result[data[i]-'a']++
		}
		resultc <- result
	}
}
