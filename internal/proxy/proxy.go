package proxy

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"

    "my-proxy-service/internal/config"
    "my-proxy-service/internal/utils"
)

func HandleProxy(w http.ResponseWriter, r *http.Request) {
    log.Printf("Incoming request: %s %s", r.Method, r.URL)

    targetURL := config.ProxyTarget
    if r.URL.String() != "/" {
        targetURL += r.URL.String()
    }
    log.Printf("Target URL: %s", targetURL)

    req, err := http.NewRequest(r.Method, targetURL, r.Body)
    if err != nil {
        utils.HandleError(w, err, http.StatusInternalServerError)
        return
    }

    authTokenValue, err := utils.ReadAuthToken()
    if err != nil {
        utils.HandleError(w, err, http.StatusInternalServerError)
        return
    }
    authTokenValue = strings.TrimSpace(authTokenValue)

    if len(authTokenValue) == 0 {
        log.Println("Authentication token is empty")
        utils.HandleError(w, fmt.Errorf("authentication token is empty"), http.StatusUnauthorized)
        return
    }

    req.Header.Set(config.AuthTokenHeader, authTokenValue)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        utils.HandleError(w, err, http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    log.Printf("Response status: %s", resp.Status)

    utils.CopyHeaders(w, resp)
    w.WriteHeader(resp.StatusCode)
    _, err = io.Copy(w, resp.Body)
    if err != nil {
        utils.HandleError(w, err, http.StatusInternalServerError)
        return
    }
}
