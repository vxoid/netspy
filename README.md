# spynet v0.1.0
## Changelog
- Init
## Libraries
- <a href="https://golang.org/x/net/html">HTML parser</a>
## Examples
```go
package main

import (
	"fmt"
	"net/http"
	
	"github.com/CURVoid/netspy"
)

func handler(r *http.Response) string {
	return r.Request.URL.String()
}

func main() {
	start := "https://books.toscrape.com/index.html"
	hosts := []string{"books.toscrape.com"}
	rules := []netspy.Rule{
		netspy.NewRule([]string{"category"}, []string{}),
		netspy.NewRule([]string{"catalogue"}, []string{"category"}),
	}

	crawler := netspy.NewCrawler[string](start, hosts, rules, handler)

	outputs, err := crawler.Crawl()
	if err != nil {
		panic(err)
	}

	fmt.Printf("passed - %d\n", len(outputs))
	for _, output := range outputs {
		println(output)
	}
}
```
## Roadmap
- add proxies support