package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/handlers"
	"github.com/lavianrose/flowforge/internal/middleware"
	"github.com/lavianrose/flowforge/internal/repository"
)

type Server struct {
	app          *fiber.App
	cfg          *config.Config
	jwtMgr       *auth.JWTManager
	authMW       *middleware.AuthMiddleware
	authHdl      *handlers.AuthHandler
	workflowHdl  *handlers.WorkflowHandler
	runHdl       *handlers.RunHandler
}

func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret)
	authMW := middleware.NewAuthMiddleware(jwtMgr)
	userRepo := repository.NewUserRepository(db.Pool)
	authHdl := handlers.NewAuthHandler(userRepo, jwtMgr)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	runRepo := repository.NewRunRepository(db.Pool)
	workflowHdl := handlers.NewWorkflowHandler(workflowRepo, runRepo)
	runHdl := handlers.NewRunHandler(runRepo)

	return &Server{
		app: app,
		cfg: cfg,
		jwtMgr: jwtMgr,
		authMW: authMW,
		authHdl: authHdl,
		workflowHdl: workflowHdl,
		runHdl: runHdl,
	}
}

func (s *Server) Setup() {
	// Middleware
	s.app.Use(logger.New())
	s.app.Use(recover.New())
	s.app.Use(cors.New())

	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// API routes
	api := s.app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", s.authHdl.Login)
	auth.Get("/me", s.authMW.Auth(), s.authHdl.Me)

	// Protected routes
	api.Use(s.authMW.Auth())

	// Workflow routes
	api.Get("/workflows", s.workflowHdl.ListWorkflows)
	api.Post("/workflows", s.workflowHdl.CreateWorkflow)
	api.Get("/workflows/:id", s.workflowHdl.GetWorkflow)
	api.Put("/workflows/:id", s.workflowHdl.UpdateWorkflow)
	api.Delete("/workflows/:id", s.workflowHdl.DeleteWorkflow)
	api.Post("/workflows/:id/trigger", s.workflowHdl.TriggerWorkflow)

	// Run routes
	api.Get("/runs", s.runHdl.ListRuns)
	api.Get("/runs/:id", s.runHdl.GetRun)
	api.Get("/runs/:id/stream", s.runHdl.StreamRun)
}

func (s *Server) Start() error {
	return s.app.Listen(":" + s.cfg.Port)
}
