package server

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/user/crypto-invoice-generator/backend/internal/config"
	"github.com/user/crypto-invoice-generator/backend/internal/handler"
	"github.com/user/crypto-invoice-generator/backend/internal/repository"
	"github.com/user/crypto-invoice-generator/backend/internal/service"
	"github.com/user/crypto-invoice-generator/backend/internal/watcher"
	"gorm.io/gorm"
)

type Server struct {
	Cfg     *config.Config
	Gin     *gin.Engine
	DB      *gorm.DB
	Watcher *watcher.Watcher
}

func NewServer(cfg *config.Config, router *gin.Engine, db *gorm.DB) *Server {
	return &Server{
		Cfg: cfg,
		Gin: router,
		DB:  db,
	}
}

func (s *Server) Shutdown(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}

func ConfigRoutesAndSchedulers(s *Server) {
	s.Gin.Use(HandleOption)

	// Init Eth Client
	client, err := ethclient.Dial(s.Cfg.Ethereum.RPCURL)
	if err != nil {
		// Log fatal since we need blockchain for everything now
		panic("Failed to connect to Ethereum RPC: " + err.Error())
	}

	// Setup Layers
	repo := repository.NewInvoiceRepository(s.DB)
	svc := service.NewInvoiceService(repo, s.Cfg, client)
	h := handler.NewInvoiceHandler(svc)

	// Start Watcher (Background)
	w := watcher.NewWatcher(s.DB, repo, s.Cfg, client)
	s.Watcher = w
	w.Start()

	// Setup Router
	api := s.Gin.Group("/api")
	{
		api.POST("/invoices", h.CreateInvoice)
		api.GET("/invoices/:id", h.GetInvoice)
	}
}

// HandleOption sets security headers and CORS options
func HandleOption(c *gin.Context) {
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	requestOrigin := c.Request.Header.Get("Origin")

	// By default, block
	allowOrigin := ""

	// Special case: allow any origin by echoing back
	if allowedOriginsStr == "*" {
		allowOrigin = requestOrigin
	} else {
		allowedOrigins := make(map[string]bool)
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			allowedOrigins[strings.TrimSpace(origin)] = true
		}
		if allowedOrigins[requestOrigin] {
			allowOrigin = requestOrigin
		}
	}

	if allowOrigin != "" {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, withcredentials, X-CSRF-Token")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

	// Prevent caching
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Handle preflight
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}
