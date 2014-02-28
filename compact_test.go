package compact

import (
    "testing"
    "net/url"
    "io/ioutil"
)

func TestFile(t *testing.T) {
    c := &Converter{}

    u, _ := url.Parse("tests/index.html")

    content, err := c.Convert(u)

    err = ioutil.WriteFile("tests/file.html", []byte(content), 0666)
    if err != nil {
        t.Error(err)
    }
}

func TestUrl(t *testing.T) {
    c := &Converter{}

    u, _ := url.Parse("http://revel.github.io/index.html")

    content, err := c.Convert(u)

    err = ioutil.WriteFile("tests/url.html", []byte(content), 0666)
    if err != nil {
        t.Error(err)
    }
}