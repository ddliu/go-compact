# compact

Merge external js/css/images into a single HTML file.

## Why

A beautifuy HTML page contains CSS, images and even javascript. 
But it's painful to distribute. You have to compress all files, send, uncompress, and then figure out which file to open in your browser.

It's much easier if all these go into a single HTML file.

## Features

- Include external CSS file
- Include external JS file
- Turn `<img>` into data-uri format
- Turn css background-image into data-uri format
- Custom ignore function

## Command Line Usage

```
compact http://github.com github.html
compact /path/to/local/file.html result.html
```

## Package Usage

```go
package main

import (
    "github.com/ddliu/go-compact"
    "net/url"
    "fmt"
)

func main() {
    // u, _ := url.Parse("/home/dong/project/example.html")
    u, _ := url.Parse("http://github.com")
    content, err := compact.Convert(u, func(u *url.Url) bool {
        if u.Host == "google.com" {
            return true
        }
    })

    fmt.Println(content)
}
```

## Limit

- Webfont not supported
- The convertion is done with the help of regexp, so it might not work for some HTML pages

## Changelog

### v0.1.0 (2014-02-28)

Initial release