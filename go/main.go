package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"github.com/fsnotify/fsnotify"
)

var (
	portFlag         = flag.Int("port", 3000, "Port to run the proxy server on")
	apiFileFlag      = flag.String("api-file", "file_to_watch.txt", "Path to the file containing the API key")
	proxyTargetFlag  = flag.String("proxy-target", "http://example.com", "Target URL for proxying requests")
	authTokenHeader  = flag.String("auth-token-header", "authorization", "Header name for authentication token")
	mutex            sync.Mutex
)

func main() {
	flag.Parse()

	go watchFile(*apiFileFlag)

	http.HandleFunc("/", proxyHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := *proxyTargetFlag + r.URL.String()
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	authTokenValue, err := readAuthToken()
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	req.Header.Set(*authTokenHeader, authTokenValue)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w, resp)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
}

func readAuthToken() (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	content, err := ioutil.ReadFile(*apiFileFlag)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func watchFile(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatalf("Error adding file to watcher: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("File modified, updating authentication token...")
				if err := handleFileChange(filePath); err != nil {
					log.Printf("Error updating authentication token: %v", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error watching file:", err)
		}
	}
}

func handleFileChange(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		log.Println("Warning: Empty file content.")
		return nil
	}

	log.Println("Authentication token updated successfully.")
	return nil
}

func copyHeaders(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

func handleError(w http.ResponseWriter, err error, code int) {
	log.Printf("Error: %v", err)
	http.Error(w, err.Error(), code)
}
