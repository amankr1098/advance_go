package main

import (
	"fmt"
	"slices"
	"sync"
)

func main() {

	//stage 1:Genrate integer 1 - 100
	n := 100
	ch := make(chan int)
	ch1 := make(chan int)
	done := make(chan bool)
	result := []int{}

	//stage 2 fan-out 3 worker
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go worker(ch, ch1, 0, wg) // worker 1
	go worker(ch, ch1, 1, wg) // worker 2
	go worker(ch, ch1, 2, wg) // worker 3

	go func() {
		for v := range ch1 {
			// fmt.Println("prime no ", v)
			result = append(result, v)
		}
		done <- true
	}()

	for i := 1; i <= n; i++ {
		ch <- i
	}

	close(ch)
	wg.Wait()

	close(ch1)
	<-done

	slices.Sort(result)

	for _, v := range result {
		fmt.Println(v)
	}

	fmt.Println("program exited")

}

func isPrime(num int) bool {
	count := 0
	for i := 1; i <= num; i++ {
		if num%i == 0 {
			count++
		}
	}
	if count == 2 {
		return true
	}

	return false

}

func worker(ch chan int, ch1 chan int, job int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("worker started", job)
	for v := range ch {
		if isPrime(v) {
			ch1 <- v
		}
	}
}
