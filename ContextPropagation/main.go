package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Result struct {
	URL    string
	Result string
	Err    error
}

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	urls := []string{
		"https://www.google.com",
		"https://www.example.com",
		"https://www.github.com",
		"https://www.golang.org",
		"https://httpbin.org/delay/2",
		"https://httpbin.org/status/404",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := fetchAll(ctx, urls)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Results:", len(result))
	fmt.Println("process ends")
}

func fetchAll(ctx context.Context, urls []string) ([]string, error) {
	fmt.Printf("fetchAll called ....\n ")
	// ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan Result, len(urls))
	r := make([]string, 0)
	// wg := &sync.WaitGroup{}
	for _, url := range urls {
		go func(url string) {
			res, err := fetch(ctx, url)

			result <- Result{
				URL:    url,
				Result: res,
				Err:    err,
			}
		}(url)
	}
	for i := 0; i < len(urls); i++ {
		v := <-result
		if v.Err != nil {
			fmt.Printf("Error fetching URL %v: %v\n", v.URL, v.Err)
			return r, v.Err
		}
		fmt.Printf("Fetched URL %v: %d bytes\n", v.URL, len(v.Result))
		r = append(r, v.Result)
	}

	fmt.Printf("fetchAll completed ....\n ")

	// wg.Wait()

	return r, nil
}

func fetch(ctx context.Context, requrl string) (string, error) {
	fmt.Println("fetching the info for URL ", requrl)

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.Proxy = http.ProxyFromEnvironment
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.Client{
		Transport: customTransport,
		Timeout:   time.Duration(3 * time.Second),
	}
	parseUrl, _ := url.Parse(requrl)
	req := &http.Request{
		URL: parseUrl,
	}
	req = req.WithContext(ctx)

	// for d, ok := ctx.Deadline() ; ok {}

	response, err := client.Do(req)
	if err != nil {
		// fmt.Println("error ", err.Error())
		return "", err
	}
	defer response.Body.Close()
	fmt.Printf("Status %v for URL %v \n", response.Status, requrl)
	if response.StatusCode != 200 && response.StatusCode != 201 {
		// fmt.Println("invalid response")
		return "", fmt.Errorf("invalid response, status code: %d received", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		// fmt.Println("Read error:", err.Error())
		return "", err
	}

	// fmt.Println("Length:", len(data))
	// fmt.Println("Success URL ", requrl)
	return string(data), nil
}
