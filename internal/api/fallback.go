package api

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterFallback() {
	staticDir := r.cfg.StaticDir
	r.engine.NoRoute(func(c *gin.Context) {
		// static assets (vite)
		if fileExists(staticDir, c.Request.URL.Path) {
			c.File(filepath.Join(staticDir, c.Request.URL.Path))
			return
		}

		// SPA fallback
		r.serveIndex(c)
	})
}

func (r *Router) serveIndex(c *gin.Context) {
	c.File(filepath.Join(r.cfg.StaticDir, "index.html"))
}

func fileExists(publicDir, reqPath string) bool {
	clean := filepath.Clean(reqPath)

	// prevent path traversal
	if clean == "." || clean == "/" || clean == ".." {
		return false
	}

	fullPath := filepath.Join(publicDir, clean)

	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
