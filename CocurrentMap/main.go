package main

import (
	"concurrent_map/cache"
	"fmt"
	"sync"
)

func main() {
	c := cache.NewCache[string, int]()

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key%d", i), i*10)
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, ok := c.Get(fmt.Sprintf("key%d", i))
			fmt.Printf("Get key%d: %d, found: %v\n", i, val, ok)
		}(i)
	}

	wg.Wait()
	fmt.Println("All keys:", c.GetKeys())

	c.Delete("key2")
	fmt.Println("After delete:", c.GetKeys())

}
