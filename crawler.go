package netspy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
)

type Crawler[T any] struct {
	rules         []Rule
	startURL      string
	allowed_hosts []string
	handler       func(*http.Response) T
}

func NewCrawler[T any](startURL string, allowed_hosts []string, rules []Rule, handler func(*http.Response) T) Crawler[T] {
	if len(rules) == 0 {
		rules = append(rules, NewRule([]string{}, []string{}))
	}

	return Crawler[T]{
		startURL:      startURL,
		allowed_hosts: allowed_hosts,
		rules:         rules,
		handler:       handler,
	}
}

func (crawler *Crawler[T]) Crawl() ([]T, error) {
	baseURL, err := baseURLof(crawler.startURL)
	if err != nil {
		return []T{}, err
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}

	response, err := client.Get(crawler.startURL)
	if err != nil {
		return []T{}, err
	}

	defer response.Body.Close()

	if response.StatusCode > 399 {
		return []T{}, fmt.Errorf("got %s code", response.Status)
	}

	var passed []map[string]bool
	for range crawler.rules {
		passed = append(passed, make(map[string]bool))
	}

	var mutex sync.Mutex
	outputs, err := crawl[T](response.Body, baseURL, crawler.allowed_hosts, crawler.rules, 0, passed, &mutex, crawler.handler, &client)

	return outputs, err
}

func crawl[T any](
	htmlReader io.Reader,
	startURL string,
	allowed_hosts []string,
	rules []Rule,
	rule_n int,
	passed []map[string]bool,
	mutex *sync.Mutex,
	handler func(*http.Response) T,
	client *http.Client,
) ([]T, error) {
	node, err := html.Parse(htmlReader)
	if err != nil {
		return []T{}, err
	}

	links := searchForLinks(node, startURL+"/..")

	var wg sync.WaitGroup
	used := make(map[string]bool)
	childsOutputs := make(chan []T)
	var outputs []T

	for _, link := range links {
		if used[link] {
			continue
		}
		used[link] = true
		mutex.Lock()
		if passed[rule_n][link] {
			mutex.Unlock()
			continue
		}
		mutex.Unlock()

		parsedURL, err := url.Parse(link)
		if err != nil {
			continue
		}

		if len(allowed_hosts) > 0 && !includes(allowed_hosts, parsedURL.Host) {
			continue
		}

		var new_rule_n int
		if rule_n == 0 {
			if rules[rule_n].Pass(parsedURL) {
				new_rule_n = rule_n + 1
			} else {
				continue
			}
		} else {
			mutex.Lock()
			if rules[rule_n].Pass(parsedURL) {
				new_rule_n = rule_n + 1
			} else if rules[0].Pass(parsedURL) && !passed[0][link] {
				new_rule_n = 1
			} else {
				mutex.Unlock()
				continue
			}
			mutex.Unlock()
		}
		mutex.Lock()
		passed[new_rule_n-1][link] = true
		mutex.Unlock()

		response, err := client.Get(link)
		if err != nil {
			fmt.Printf("request error %e\n", err)
			continue
		}

		defer response.Body.Close()

		if response.StatusCode > 399 {
			continue
		}

		if new_rule_n == len(rules) {
			outputs = append(outputs, handler(response))
			new_rule_n = 0
		}

		wg.Add(1)
		go func(htmlReader io.Reader,
			startURL string,
			allowed_hosts []string,
			rules []Rule,
			rule_n int,
			passed []map[string]bool,
			mutex *sync.Mutex,
			handler func(*http.Response) T,
			client *http.Client,
			channel chan []T,
			wg *sync.WaitGroup,
		) {
			defer wg.Done()

			outputs, err := crawl[T](htmlReader, startURL, allowed_hosts, rules, rule_n, passed, mutex, handler, client)
			if err != nil {
				return
			}

			channel <- outputs
		}(response.Body, link, allowed_hosts, rules, new_rule_n, passed, mutex, handler, client, childsOutputs, &wg)
	}
	go func() {
		wg.Wait()
		close(childsOutputs)
	}()

	for outputs_batch := range childsOutputs {
		outputs = append(outputs, outputs_batch...)
	}

	return outputs, nil
}
