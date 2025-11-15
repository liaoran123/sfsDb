package db

import (
	"fmt"
	"testing"
	"time"
)

func producer(ch chan<- int) {
	for i := 1; i <= 5; i++ {
		fmt.Printf("Producing: %d\n", i)
		ch <- i
		time.Sleep(time.Millisecond * 500)
	}
	close(ch)
}

func consumer(ch <-chan int) {
	// 从通道接收数据，直到通道关闭close(ch)
	// 这个循环实质就是一个排队等待操作
	for num := range ch {
		fmt.Printf("Consuming: %d\n", num)
		time.Sleep(time.Millisecond * 1000)
	}
}

func TestChannel(t *testing.T) {
	ch := make(chan int)
	go producer(ch)
	consumer(ch)
}

// channel实现一个消息队列
func TestChannel1(t *testing.T) {
	queue := make(chan string, 5) // 创建一个带缓冲区大小为5的字符串通道
	queue <- "one"                // 入队
	queue <- "two"
	queue <- "three"
	fmt.Println(<-queue) // 出队，输出"one"
	fmt.Println(<-queue) // 出队，输出"two"
	queue <- "four"
	queue <- "five"
	close(queue) // 关闭通道，表明队列已经结束
	for elem := range queue {
		fmt.Println(elem) // 遍历队列中剩余的元素，输出"three", "four", "five"
	}
}
