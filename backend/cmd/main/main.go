// @title           Design Portfolio API
// @version         1.0
// @description     Backend API for Olga's design portfolio.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"design-profile/backend/config"
	_ "design-profile/backend/docs"
	"design-profile/backend/internal/email"
	"design-profile/backend/internal/handler"
	"design-profile/backend/internal/middleware"
	"design-profile/backend/internal/repository"
	"design-profile/backend/internal/service"
	"design-profile/backend/migrations"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfgPath := flag.String("config", "./config/config.yaml", "path to config file")
	flag.Parse()

	// Структурированный логгер: время | уровень | сообщение | поля.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	// Connect to PostgreSQL.
	pool, err := repository.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations.
	if err := repository.Migrate(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// Build dependency tree.
	otpRepo := repository.NewOTPRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	requestRepo := repository.NewRequestRepository(pool)
	contactRepo := repository.NewContactRepository(pool)

	emailClient := email.NewClient(
		cfg.Email.SMTPHost, cfg.Email.SMTPPort,
		cfg.Email.Username, cfg.Email.Password, cfg.Email.Sender,
	)

	authSvc := service.NewAuthService(otpRepo, emailClient, cfg.Email.AdminEmail, cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	projectSvc := service.NewProjectService(projectRepo)
	requestSvc := service.NewRequestService(requestRepo)
	contactSvc := service.NewContactService(contactRepo)

	authH := handler.NewAuthHandler(authSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	requestH := handler.NewRequestHandler(requestSvc)
	contactH := handler.NewContactHandler(contactSvc)

	r := gin.Default()

	// Global CORS middleware.
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Swagger UI.
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes.
	v1 := r.Group("/api/v1")

	// Public auth routes.
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/request-otp", authH.RequestOTP)
		authGroup.POST("/verify-otp", authH.VerifyOTP)
	}

	// Public project routes.
	projectsGroup := v1.Group("/projects")
	{
		projectsGroup.GET("", projectH.ListProjects)
		projectsGroup.GET("/:id", projectH.GetProject)
		projectsGroup.GET("/:id/media/:mediaId", projectH.ServeMedia)
	}

	// Public contact routes.
	v1.GET("/contacts", contactH.GetContacts)

	// Public request submission.
	v1.POST("/requests", requestH.CreateRequest)

	// Admin routes (require JWT).
	admin := v1.Group("/admin")
	admin.Use(middleware.RequireAuth(cfg.JWT.Secret))
	{
		// Project management.
		adminProjects := admin.Group("/projects")
		{
			adminProjects.POST("", projectH.CreateProject)
			adminProjects.PUT("/:id", projectH.UpdateProject)
			adminProjects.DELETE("/:id", projectH.DeleteProject)
			adminProjects.POST("/:id/media", projectH.AddMedia)
			adminProjects.DELETE("/:id/media/:mediaId", projectH.DeleteMedia)
		}

		// Requests management.
		adminRequests := admin.Group("/requests")
		{
			adminRequests.GET("", requestH.ListRequests)
			adminRequests.GET("/:id", requestH.GetRequest)
			adminRequests.GET("/:id/attachments/:attachmentId", requestH.ServeAttachment)
		}

		// Contacts management.
		admin.PUT("/contacts", contactH.UpdateContacts)
	}

	addr := cfg.Server.Addr()
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
