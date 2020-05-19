<!--
 * @Author: jinde.zgm
 * @Date: 2020-05-19 16:19:53
 * @Description:  gopool readme
--> 


## Introduction

Gopool using lock-free queue to implement coroutine pool  with fixed capacity, managing and recycling a large number of goroutines, allowing developers to limit the number of goroutines in your concurrent programs.

## Features:

- High-performance managing and recycling of a large number of coroutines
- Periodically clean idle timeout coroutine
- Support blocking and nonblocking
