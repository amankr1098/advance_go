package main

import "concurrent_map/cache"

func main() {
	cacheInstance := cache.NewCache[string, string]()

	cacheInstance.Set("aman", "kumar")

}
