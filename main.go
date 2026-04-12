package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort    = 8080
	defaultHost    = "127.0.0.1" // personal use only — bind to localhost by default
	appName        = "sub2api"
	appVersion     = "dev"
)

// Config holds the application configuration
type Config struct {
	Host    string
	Port    int
	Debug   bool
	Token   string
}

func main() {
	cfg := parseConfig()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRouter(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("%s %s starting on %s", appName, appVersion, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}

// parseConfig reads configuration from flags and environment variables.
// Environment variables take precedence over defaults; flags override env vars.
func parseConfig() Config {
	cfg := Config{
		Host:  getEnv("SUB2API_HOST", defaultHost),
		Port:  getEnvInt("SUB2API_PORT", defaultPort),
		Debug: getEnvBool("SUB2API_DEBUG", false),
		Token: getEnv("SUB2API_TOKEN", ""),
	}

	flag.StringVar(&cfg.Host, "host", cfg.Host, "host address to listen on")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "port to listen on")
	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "enable debug mode")
	flag.StringVar(&cfg.Token, "token", cfg.Token, "optional bearer token for API authentication")
	flag.Parse()

	return cfg
}

// setupRouter initialises the Gin engine and registers all routes.
func setupRouter(cfg Config) *gin.Engine {
	r := gin.Default()

	// Health / readiness probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": appVersion})
	})

	// API v1 group — optionally protected by a static bearer token
	v1 := r.Group("/api/v1")
	if cfg.Token != "" {
		v1.Use(bearerAuth(cfg.Token))
	}

	v1.GET("/convert", handleConvert)

	return r
}

// bearerAuth returns a middleware that validates a static bearer token.
func bearerAuth(token string) gin.HandlerFunc {
	expected := "Bearer " + token
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// handleConvert is a placeholder for the subscription-to-API conversion logic.
func handleConvert(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}
	// TODO: implement conversion logic
	c.JSON(http.StatusNotImplemented, gin.H{"message": "conversion not yet implemented", "url": url})
}

// --- helpers ---

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
