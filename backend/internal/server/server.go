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
	"github.com/lavianrose/flowforge/internal/scheduler"
)

type Server struct {
	app          *fiber.App
	cfg          *config.Config
	jwtMgr       *auth.JWTManager
	authMW       *middleware.AuthMiddleware
	rateLimit    *middleware.RateLimiter
	authHdl      *handlers.AuthHandler
	workflowHdl  *handlers.WorkflowHandler
	runHdl       *handlers.RunHandler
	scheduleHdl  *handlers.ScheduleHandler
	webhookHdl   *handlers.WebhookHandler
	scheduler    *scheduler.Scheduler
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
	rateLimit.AddConfig("auth", 10, time.Minute, middleware.ByIP)           // 10 req/min for auth
	rateLimit.AddConfig("read", 100, time.Minute, middleware.ByUserID)      // 100 req/min for reads
	rateLimit.AddConfig("write", 30, time.Minute, middleware.ByUserID)      // 30 req/min for writes
	rateLimit.AddConfig("trigger", 10, time.Minute, middleware.ByUserID)    // 10 req/min for triggers

	userRepo := repository.NewUserRepository(db.Pool)
	authHdl := handlers.NewAuthHandler(userRepo, jwtMgr)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	runRepo := repository.NewRunRepository(db.Pool)
	workflowHdl := handlers.NewWorkflowHandler(workflowRepo, runRepo)
	runHdl := handlers.NewRunHandler(runRepo)

	scheduleRepo := repository.NewScheduleRepository(db.Pool)
	scheduleHdl := handlers.NewScheduleHandler(scheduleRepo, workflowRepo, runRepo)

	webhookRepo := repository.NewWebhookRepository(db.Pool)
	webhookHdl := handlers.NewWebhookHandler(webhookRepo, workflowRepo, runRepo)

	sched := scheduler.NewScheduler(scheduleRepo, workflowRepo, runRepo)

	return &Server{
		app: app,
		cfg: cfg,
		jwtMgr: jwtMgr,
		authMW: authMW,
		rateLimit: rateLimit,
		authHdl: authHdl,
		workflowHdl: workflowHdl,
		runHdl: runHdl,
		scheduleHdl: scheduleHdl,
		webhookHdl: webhookHdl,
		scheduler: sched,
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

	// Start HTTP server
	return s.app.Listen(":" + s.cfg.Port)
}
