package config

import (
    "flag"
    "log"
    "os"
    "strconv"
    "sync"
)


// Constants for routes
var ROUTES = struct {
    INDEX  string
    HEALTH string
}{
    INDEX:  "/",
    HEALTH: "/healthz",
}

var (
    Port            int
    ApiFile         string
    ProxyTarget     string
    AuthTokenHeader string
    mutex           sync.Mutex
)

func init() {
    flag.IntVar(&Port, "port", 3000, "Port to run the proxy server on")
    flag.StringVar(&ApiFile, "api-file", "file_to_watch.txt", "Path to the file containing the API key")
    flag.StringVar(&ProxyTarget, "proxy-target", "http://example.com", "Target URL for proxying requests")
    flag.StringVar(&AuthTokenHeader, "auth-token-header", "authorization", "Header name for authentication token")
}

func LoadConfig() {
    flag.Parse()
    setFlagFromEnv("PORT", &Port)
    setFlagFromEnv("API_FILE", &ApiFile)
    setFlagFromEnv("PROXY_TARGET", &ProxyTarget)
    setFlagFromEnv("AUTH_TOKEN_HEADER", &AuthTokenHeader)
}

func setFlagFromEnv(envVar string, flagValue interface{}) {
    if value := os.Getenv(envVar); value != "" {
        switch v := flagValue.(type) {
        case *int:
            val, err := strconv.Atoi(value)
            if err != nil {
                log.Fatalf("Error parsing %s value: %v", envVar, err)
            }
            *v = val
        case *string:
            *v = value
        default:
            log.Fatalf("Unsupported flag type: %T", v)
        }
    }
}
