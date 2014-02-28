package compact

import (
    "regexp"
    "strings"
    "errors"

    "io/ioutil"
    "path/filepath"

    "encoding/base64"

    "net/url"
    "net/http"

    "mime"

    "log"
)

// URL types(local file is also considered as url)
const (
    U_URL = iota
    U_FILE
    U_DATA
    U_UNKNOWN
)

var httpcache = make(map[url.URL][]byte)


type Converter struct {
    IgnoreFunc func(*url.URL) bool
    IgnoreContentImageFunc func(*url.URL) bool
    IgnoreCssFunc func(*url.URL) bool
    IgnoreCssImageFunc func(*url.URL) bool
    IgnoreJsFunc func(*url.URL) bool
}

func (this *Converter) Convert(u *url.URL) (string, error) {

    fixRelativePath(u)

    c, err := fetch(u)
    content := string(c)

    if err != nil {
        return "", err
    }

    // convert content image
    content = ConvertContentImage(content, u, func(u *url.URL) bool {
        if this.IgnoreFunc != nil && this.IgnoreFunc(u) {
            return true
        }

        if this.IgnoreContentImageFunc != nil && this.IgnoreContentImageFunc(u) {
            return true
        }

        return false
    })

    // convert js
    content = ConvertExternalJs(content, u, func(u *url.URL) bool {
        if this.IgnoreFunc != nil && this.IgnoreFunc(u) {
            return true
        }

        if this.IgnoreJsFunc != nil && this.IgnoreJsFunc(u) {
            return true
        }

        return false
    })

    // convert css
    content = ConvertExternalCss(content, u, func(u *url.URL) bool {
        if this.IgnoreFunc != nil && this.IgnoreFunc(u) {
            return true
        }

        if this.IgnoreCssFunc != nil && this.IgnoreCssFunc(u) {
            return true
        }

        return false
    }, func(u *url.URL) bool {
        if this.IgnoreFunc != nil && this.IgnoreFunc(u) {
            return true
        }

        if this.IgnoreCssImageFunc != nil && this.IgnoreCssImageFunc(u) {
            return true
        }

        return false
    })

    return content, nil
}

func Convert(u *url.URL, ignore func(*url.URL) bool) (string, error) {
    c := &Converter {
        IgnoreFunc: ignore,
    }

    return c.Convert(u)
}

// Convert images in html into data-uri format
func ConvertContentImage(content string, u *url.URL, ignoreImageUrlFunc func(*url.URL) bool) string {
    fixRelativePath(u)

    // replace image
    re := regexp.MustCompile(`(?i)(<img\s*[^>]*\s*src=")([^">]+)("[^>]*>)`)

    content = re.ReplaceAllStringFunc(content, func(m string) string {
        parts := re.FindStringSubmatch(m)

        src := parts[2]

        imgUrl, err := url.Parse(src)

        if err != nil {
            log.Println(err)
            return m
        }

        imgUrl = u.ResolveReference(imgUrl)

        if ignoreImageUrlFunc != nil && ignoreImageUrlFunc(imgUrl) {
            return m
        }

        dataUri, err := ConvertDataURI(imgUrl)

        if err != nil {
            log.Println(err)
            return m
        }

        return parts[1] + dataUri + parts[3]
    })

    return content
}

// Embed external css files
func ConvertExternalCss(content string, u *url.URL, ignoreCssUrlFunc func(*url.URL) bool, ignoreImageUrlFunc func(*url.URL) bool) string {
    fixRelativePath(u)

    re := regexp.MustCompile(`(?i)(<link)(\s*[^>]*\s*)(href=")([^"]+)(")([^>]*>)`)

    reRel := regexp.MustCompile(`\s+rel\s*=\s*"stylesheet"`)

    content = re.ReplaceAllStringFunc(content, func(m string) string {
        parts := re.FindStringSubmatch(m)

        // rel should be stylesheet
        if !reRel.MatchString(parts[0]) {
            return m
        }

        cssUrl, err := url.Parse(parts[4])

        if err != nil {
            log.Println(err)
            return m
        }

        cssUrl = u.ResolveReference(cssUrl)

        if ignoreCssUrlFunc != nil && ignoreCssUrlFunc(cssUrl) {
            return m
        }

        cssContentBytes, err := fetch(cssUrl)

        if err != nil {
            log.Println(err)
            return m
        }

        cssContent := ConvertCssImage(string(cssContentBytes), cssUrl, ignoreImageUrlFunc)

        return "<style" + parts[2] + strings.Replace(parts[6], "/>", ">", -1) + cssContent + "</style>"
    })

    return content
}

// Embed external js files
func ConvertExternalJs(content string, u *url.URL, ignore func(*url.URL) bool) string {
    fixRelativePath(u)

    re := regexp.MustCompile(`(?i)(<script\s*[^>]*\s*)(src=")([^">]+)(")([^>]*>)(\s*</script>)`)

    content = re.ReplaceAllStringFunc(content, func(m string) string {
        parts := re.FindStringSubmatch(m)

        src := parts[3]

        jsUrl, err := url.Parse(src)

        if err != nil {
            log.Println(err)
            return m
        }

        jsUrl = u.ResolveReference(jsUrl)

        if ignore != nil && ignore(jsUrl) {
            return m
        }

        jsContent, err := fetch(jsUrl)
        if err != nil {
            log.Println(err)
            return m
        }

        return parts[1] + parts[5] + string(jsContent) + parts[6]
    })

    return content
}

// Convert background images into data-uri format
func ConvertCssImage(content string, u *url.URL, ignoreImageUrlFunc func(*url.URL) bool) string {
    fixRelativePath(u)

    re := regexp.MustCompile(`(?Ui)(background(\-image)?\s*:\s*[^;]*url\()(.+)(\)[^;]*;)`)

    content = re.ReplaceAllStringFunc(content, func(m string) string {
        parts := re.FindStringSubmatch(m)

        src := parts[3]
        src = strings.Trim(src, `"' `)
        imgUrl, err := url.Parse(src)

        if err != nil {
            log.Println(err)
            return m
        }

        imgUrl = u.ResolveReference(imgUrl)

        if checkUrl(imgUrl) == U_DATA {
            return m
        }

        if ignoreImageUrlFunc != nil && ignoreImageUrlFunc(imgUrl) {
            return m
        }

        dataUri, err := ConvertDataURI(imgUrl)

        if err != nil {
            log.Println(err)
            return m
        }

        return parts[1] + "\"" + dataUri + "\"" + parts[4]
    })

    return content
}

func checkUrl(u *url.URL) int {
    if strings.ToLower(u.Scheme) == "http" || strings.ToLower(u.Scheme) == "https" {
        return U_URL
    } else if strings.ToLower(u.Scheme) == "file" || u.Scheme == "" {
        return U_FILE
    } else if strings.ToLower(u.Scheme) == "data" {
        return U_DATA
    } else {
        return U_UNKNOWN
    }
}

func fixRelativePath(u *url.URL) {
    if checkUrl(u) == U_FILE {
        if !filepath.IsAbs(u.Path) {
            if abs, err := filepath.Abs(u.Path); err == nil {
                u.Path = abs
            }
        }
    }
}

// Fetch url content
func fetch(u *url.URL) ([]byte, error) {
    utype := checkUrl(u)

    if utype == U_URL {
        content, ok := httpcache[*u]
        if !ok {
            res, err := http.Get(u.String())
            if err != nil {
                return nil, err
            }

            if res.StatusCode != 200 {
                return nil, errors.New("Response code of " + u.String() + " is not 200")
            }

            defer res.Body.Close()
            content, err = ioutil.ReadAll(res.Body)
            if err != nil {
                return nil, err
            }

            httpcache[*u] = content
        }

        return content, nil
    } else if utype == U_FILE {
        return ioutil.ReadFile(u.Path)
    } else {
       return nil, errors.New("Unrecognized url " + u.String())
    }
}

// Convert url into data-uri format
func ConvertDataURI(u *url.URL) (string, error) {
    content, err := fetch(u)
    if err != nil {
        return "", err
    }

    ext := filepath.Ext(u.Path)
    m := mime.TypeByExtension(ext)

    if m == "" {
        m = "image/png"
    }

    return "data: " + m + ";base64," + base64.StdEncoding.EncodeToString(content), nil
}