# http-throttle

[![GoDoc](https://godoc.org/github.com/jybp/http-throttle?status.svg)](https://godoc.org/github.com/jybp/http-throttle)

Package http-throttle provides a http.RoundTripper to rate limit HTTP requests.

## Usage

```go
package example

import (
    "net/http"
    "github.com/jybp/http-throttle"
    "golang.org/x/time/rate"
)

func Example() {
    client := &http.Client{
        Transport: throttle.Default(
            // Returns ErrQuotaExceeded if more than 36000 requests occured within an hour.
            throttle.NewQuota(time.Hour, 36000), 
            // Blocks to never exceed 99 requests per second.
            rate.NewLimiter(99, 1), 
        ),
    }
    resp, err := client.Get("https://golang.org/")
    if err == throttle.ErrQuotaExceeded {
        // Handle err.
    }
    if err != nil {
        // Handle err.
    }
    _ = resp // Do something with resp.
}
```