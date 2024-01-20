package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type SafeURL struct {
	mu sync.Mutex
	v  map[string]string
}

func (c *SafeURL) Value(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.v[key]
	return val, ok
}

func (c *SafeURL) append(key string) {
	c.mu.Lock()
	c.v[key] = key
	c.mu.Unlock()
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, urlHash *SafeURL, wg *sync.WaitGroup) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	defer wg.Done()
	_, ok := urlHash.Value(url)
	if ok {
		fmt.Printf("alread exists %s\n", url)
		return
	}

	fmt.Printf("appending %s to hash\n", url)
	urlHash.append(url)

	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, urlHash, wg)
	}
}

func main() {
	hash := SafeURL{v: make(map[string]string)}
	var wg sync.WaitGroup
	wg.Add(1)
	go Crawl("https://golang.org/", 4, fetcher, &hash, &wg)
	wg.Wait()

	fmt.Printf("%v", hash.v)

}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
