// Package proxy implements a proxy that forward HTTP requests.
package main

import (
	"io"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
)

// PrintHTTP prints debug request info
func PrintHTTP(req *http.Request, res *http.Response) {
	fmt.Printf("%v %v\n", req.Method, req.RequestURI)
	for k, v := range req.Header {
		fmt.Println(k, ":", v)
	}
	fmt.Println("==============================")
	fmt.Printf("HTTP/1.1 %v\n", res.Status)
	for k, v := range res.Header {
		fmt.Println(k, ":", v)
	}
	fmt.Println(res.Body)
	fmt.Println("==============================")
}

// GetPathMapping returns a key/value map of the PATH_MAPPING env var.
func GetPathMapping() map[string]string {
	pathMappingValue := os.Getenv("PATH_MAPPING")

	var pathMapping map[string]string
	json.Unmarshal([]byte(pathMappingValue), &pathMapping)
	return pathMapping
}

// GetURLPathPrefix returns the first directory if it exists for a given URL.
func GetURLPathPrefix(requestedURL string) string {
	urlParts := strings.Split(requestedURL, "/")
	if len(urlParts) > 0 {
		return urlParts[0]
	}
	return ""
}

// NormalizeURL adds a scheme to a given URL without scheme.
func NormalizeURL(rawURL string) string {
	if !strings.Contains(rawURL, "http") {
		return "http://" + rawURL
	}
	return rawURL
}

// IsValidURL validates URL format and ensure host contains a dot.
func IsValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if !strings.Contains(parsedURL.Host, ".") {
		return false
	}
	return true
}

// Proxy struct
type Proxy struct {
}

// NewProxy factory
func NewProxy() *Proxy { return &Proxy{} }

// Handler handler process all the incoming HTTP requests.
func (p *Proxy) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	// Forward to Status page when homepage is requested
	if r.URL.Path == "/" {
		return
	}

	// Get the requested URL path and trim the / prefix
	requestedURL := strings.TrimPrefix(r.URL.String(), "/")

	urlPathPrefix := GetURLPathPrefix(requestedURL)
	pathMapping := GetPathMapping()

	if host, ok := pathMapping[urlPathPrefix]; ok {
		requestedURL = strings.Replace(requestedURL, urlPathPrefix+"/", "", 1)

		// Forward to 404 page when URL path prefix matches the requested URL
		if urlPathPrefix == requestedURL {
			http.Error(wr, err.Error(), http.StatusNotFound)
			return
		}

		requestedURL = host + "/" + requestedURL
	}

	// Add default HTTP scheme if not provided
	requestedURL = NormalizeURL(requestedURL)

	// Ensure we forward valid URLs
	if !IsValidURL(requestedURL) {
		http.Error(wr, "Invalid URL", http.StatusNotFound)
		return
	}

	log.Printf("%v %v", r.Method, requestedURL)
	req, err = http.NewRequest(r.Method, requestedURL, r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	defer r.Body.Close()

	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
	defer resp.Body.Close()

	PrintHTTP(r, resp)
}

func main() {
	godotenv.Load()

	proxy := NewProxy()
	fmt.Println("==============================")
	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), proxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
