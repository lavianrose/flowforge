package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/execution"
	"github.com/lavianrose/flowforge/internal/handlers"
	"github.com/lavianrose/flowforge/internal/middleware"
	"github.com/lavianrose/flowforge/internal/repository"
	"github.com/lavianrose/flowforge/internal/scheduler"
)

type Server struct {
	app         *fiber.App
	cfg         *config.Config
	jwtMgr      *auth.JWTManager
	authMW      *middleware.AuthMiddleware
	rateLimit   *middleware.RateLimiter
	authHdl     *handlers.AuthHandler
	workflowHdl *handlers.WorkflowHandler
	runHdl      *handlers.RunHandler
	scheduleHdl *handlers.ScheduleHandler
	webhookHdl  *handlers.WebhookHandler
	statsHdl    *handlers.StatsHandler
	scheduler   *scheduler.Scheduler
	engine      *execution.Engine
}

func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret)
	authMW := middleware.NewAuthMiddleware(jwtMgr)
	rateLimit := middleware.NewRateLimiter(db.RDB)

	// Configure rate limits
	rateLimit.AddConfig("auth", 10, time.Minute, middleware.ByIP)        // 10 req/min for auth
	rateLimit.AddConfig("read", 100, time.Minute, middleware.ByUserID)   // 100 req/min for reads
	rateLimit.AddConfig("write", 30, time.Minute, middleware.ByUserID)   // 30 req/min for writes
	rateLimit.AddConfig("trigger", 10, time.Minute, middleware.ByUserID) // 10 req/min for triggers

	userRepo := repository.NewUserRepository(db.Pool)
	authHdl := handlers.NewAuthHandler(userRepo, jwtMgr)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	runRepo := repository.NewRunRepository(db.Pool)

	engine := execution.NewEngine(runRepo, workflowRepo)

	workflowHdl := handlers.NewWorkflowHandler(workflowRepo, runRepo, engine)
	runHdl := handlers.NewRunHandler(runRepo)

	scheduleRepo := repository.NewScheduleRepository(db.Pool)
	scheduleHdl := handlers.NewScheduleHandler(scheduleRepo, workflowRepo, runRepo)

	webhookRepo := repository.NewWebhookRepository(db.Pool)
	webhookHdl := handlers.NewWebhookHandler(webhookRepo, workflowRepo, runRepo, engine)

	statsHdl := handlers.NewStatsHandler(runRepo)

	sched := scheduler.NewScheduler(scheduleRepo, workflowRepo, runRepo, engine)

	return &Server{
		app:         app,
		cfg:         cfg,
		jwtMgr:      jwtMgr,
		authMW:      authMW,
		rateLimit:   rateLimit,
		authHdl:     authHdl,
		workflowHdl: workflowHdl,
		runHdl:      runHdl,
		scheduleHdl: scheduleHdl,
		webhookHdl:  webhookHdl,
		statsHdl:    statsHdl,
		scheduler:   sched,
		engine:      engine,
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
	auth.Post("/login", s.rateLimit.Middleware("auth"), s.authHdl.Login)
	auth.Get("/me", s.authMW.Auth(), s.authHdl.Me)

	// Protected routes
	api.Use(s.authMW.Auth())

	// Stats routes
	api.Get("/stats/health", s.rateLimit.Middleware("read"), s.statsHdl.GetHealthStats)

	// Workflow routes - All authenticated users can view
	api.Get("/workflows", s.rateLimit.Middleware("read"), s.workflowHdl.ListWorkflows)
	api.Get("/workflows/:id", s.rateLimit.Middleware("read"), s.workflowHdl.GetWorkflow)
	api.Get("/workflows/:id/versions", s.rateLimit.Middleware("read"), s.workflowHdl.GetWorkflowVersions)
	api.Get("/runs", s.rateLimit.Middleware("read"), s.runHdl.ListRuns)
	api.Get("/runs/:id", s.rateLimit.Middleware("read"), s.runHdl.GetRun)
	api.Get("/runs/:id/stream", s.rateLimit.Middleware("read"), s.runHdl.StreamRun)

	// Editor and Admin can create, update, trigger
	editorOrAdmin := s.authMW.Role("editor", "admin")
	api.Post("/workflows", s.rateLimit.Middleware("write"), editorOrAdmin, s.workflowHdl.CreateWorkflow)
	api.Put("/workflows/:id", s.rateLimit.Middleware("write"), editorOrAdmin, s.workflowHdl.UpdateWorkflow)
	api.Post("/workflows/:id/trigger", s.rateLimit.Middleware("trigger"), editorOrAdmin, s.workflowHdl.TriggerWorkflow)
	api.Post("/workflows/:id/rollback/:version", s.rateLimit.Middleware("write"), editorOrAdmin, s.workflowHdl.RollbackWorkflow)

	// Only Admin can delete
	api.Delete("/workflows/:id", s.rateLimit.Middleware("write"), s.authMW.Role("admin"), s.workflowHdl.DeleteWorkflow)

	// Schedule routes
	api.Get("/schedules", s.rateLimit.Middleware("read"), s.scheduleHdl.ListSchedules)
	api.Post("/workflows/:id/schedule", s.rateLimit.Middleware("write"), editorOrAdmin, s.scheduleHdl.CreateSchedule)
	api.Delete("/schedules/:id", s.rateLimit.Middleware("write"), editorOrAdmin, s.scheduleHdl.DeleteSchedule)

	// Webhook routes
	api.Get("/webhooks", s.rateLimit.Middleware("read"), s.webhookHdl.ListWebhooks)
	api.Post("/workflows/:id/webhook", s.rateLimit.Middleware("write"), editorOrAdmin, s.webhookHdl.CreateWebhook)
	api.Delete("/webhooks/:id", s.rateLimit.Middleware("write"), editorOrAdmin, s.webhookHdl.DeleteWebhook)

	// Public webhook trigger (no authentication)
	s.app.Post("/webhooks/:path", s.webhookHdl.TriggerWebhook)
}

func (s *Server) Start() error {
	// Start scheduler
	s.scheduler.Start()

	// Channel to listen for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server starting on port %s\n", s.cfg.Port)
		serverErrors <- s.app.Listen(":" + s.cfg.Port)
	}()

	// Wait for interrupt or server error
	select {
	case sig := <-c:
		fmt.Printf("Received signal: %v. Shutting down gracefully...\n", sig)
	case err := <-serverErrors:
		fmt.Printf("Server error: %v\n", err)
		return err
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
		return err
	}

	// Stop scheduler
	s.scheduler.Stop()

	// Wait for in-flight workflow executions to finish
	fmt.Println("Waiting for in-flight workflow executions to complete...")
	s.engine.Wait()

	fmt.Println("Server shutdown complete")
	return nil
}

// GetApp returns the Fiber app for testing purposes
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// WaitExecutions blocks until all in-flight workflow executions complete.
func (s *Server) WaitExecutions() {
	s.engine.Wait()
}
