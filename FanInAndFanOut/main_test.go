package main

import (
	"sort"
	"sync"
	"testing"
)

// -------------------- Unit tests for isPrime --------------------

func TestIsPrime_Primes(t *testing.T) {
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
	for _, p := range primes {
		if !isPrime(p) {
			t.Errorf("isPrime(%d) = false, want true", p)
		}
	}
}

func TestIsPrime_NonPrimes(t *testing.T) {
	nonPrimes := []int{1, 4, 6, 8, 9, 10, 12, 15, 20, 25, 50, 100}
	for _, n := range nonPrimes {
		if isPrime(n) {
			t.Errorf("isPrime(%d) = true, want false", n)
		}
	}
}

func TestIsPrime_EdgeCases(t *testing.T) {
	tests := []struct {
		input int
		want  bool
	}{
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
	}
	for _, tc := range tests {
		got := isPrime(tc.input)
		if got != tc.want {
			t.Errorf("isPrime(%d) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// -------------------- Unit test for worker --------------------

func TestWorker_FiltersPrimes(t *testing.T) {
	ch := make(chan int, 10)
	ch1 := make(chan int, 10)

	// Feed some numbers
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	go func() {
		for _, v := range inputs {
			ch <- v
		}
		close(ch)
	}()

	// Run a single worker
	var wg sync.WaitGroup
	wg.Add(1)
	go worker(ch, ch1, 0, &wg)

	// Close ch1 after worker finishes
	go func() {
		wg.Wait()
		close(ch1)
	}()

	var got []int
	for v := range ch1 {
		got = append(got, v)
	}
	sort.Ints(got)

	expected := []int{2, 3, 5, 7}
	if len(got) != len(expected) {
		t.Fatalf("worker produced %v, want %v", got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], expected[i])
		}
	}
}

// -------------------- Fan-out / Fan-in integration test --------------------

func TestFanOutFanIn_AllPrimesUpTo100(t *testing.T) {
	n := 100
	ch := make(chan int)
	ch1 := make(chan int)

	// Fan-out: 3 workers
	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go worker(ch, ch1, i, &wg)
	}

	// Close ch1 once all workers are done
	go func() {
		wg.Wait()
		close(ch1)
	}()

	// Stage 1: generate integers
	go func() {
		for i := 1; i <= n; i++ {
			ch <- i
		}
		close(ch)
	}()

	// Fan-in: collect results
	var results []int
	for v := range ch1 {
		results = append(results, v)
	}
	sort.Ints(results)

	expected := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
	if len(results) != len(expected) {
		t.Fatalf("got %d primes, want %d\ngot:  %v\nwant: %v", len(results), len(expected), results, expected)
	}
	for i := range expected {
		if results[i] != expected[i] {
			t.Errorf("results[%d] = %d, want %d", i, results[i], expected[i])
		}
	}
}

// -------------------- Fan-out with different N values --------------------

func TestFanOutFanIn_SmallRange(t *testing.T) {
	n := 10
	ch := make(chan int)
	ch1 := make(chan int)

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go worker(ch, ch1, i, &wg)
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	go func() {
		for i := 1; i <= n; i++ {
			ch <- i
		}
		close(ch)
	}()

	var results []int
	for v := range ch1 {
		results = append(results, v)
	}
	sort.Ints(results)

	expected := []int{2, 3, 5, 7}
	if len(results) != len(expected) {
		t.Fatalf("got %d primes, want %d\ngot:  %v\nwant: %v", len(results), len(expected), results, expected)
	}
	for i := range expected {
		if results[i] != expected[i] {
			t.Errorf("results[%d] = %d, want %d", i, results[i], expected[i])
		}
	}
}

func TestFanOutFanIn_SingleValue(t *testing.T) {
	ch := make(chan int)
	ch1 := make(chan int)

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go worker(ch, ch1, i, &wg)
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	go func() {
		ch <- 7 // prime
		close(ch)
	}()

	var results []int
	for v := range ch1 {
		results = append(results, v)
	}

	if len(results) != 1 || results[0] != 7 {
		t.Errorf("got %v, want [7]", results)
	}
}

func TestFanOutFanIn_NoPrimes(t *testing.T) {
	ch := make(chan int)
	ch1 := make(chan int)

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go worker(ch, ch1, i, &wg)
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	// Send only non-primes
	go func() {
		nonPrimes := []int{1, 4, 6, 8, 9, 10}
		for _, v := range nonPrimes {
			ch <- v
		}
		close(ch)
	}()

	var results []int
	for v := range ch1 {
		results = append(results, v)
	}

	if len(results) != 0 {
		t.Errorf("got %v, want empty slice", results)
	}
}

// -------------------- Benchmark --------------------

func BenchmarkIsPrime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isPrime(97)
	}
}

func BenchmarkFanOutFanIn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ch := make(chan int)
		ch1 := make(chan int)

		var wg sync.WaitGroup
		wg.Add(3)
		for j := 0; j < 3; j++ {
			go worker(ch, ch1, j, &wg)
		}

		go func() {
			wg.Wait()
			close(ch1)
		}()

		go func() {
			for k := 1; k <= 100; k++ {
				ch <- k
			}
			close(ch)
		}()

		for range ch1 {
		}
	}
}
