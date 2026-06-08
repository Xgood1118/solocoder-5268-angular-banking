package main

import (
	"banking/config"
	"banking/internal/account"
	"banking/internal/audit"
	"banking/internal/auth"
	"banking/internal/interest"
	"banking/internal/limit"
	"banking/internal/recon"
	"banking/internal/report"
	"banking/internal/transfer"
	"banking/pkg/database"
	"banking/pkg/redis"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	rdb, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	database.AutoMigrate(db)

	auditRepo := audit.NewRepository(db)
	auditSvc := audit.NewService(auditRepo)
	auditMiddleware := audit.NewMiddleware(auditSvc)

	limitRepo := limit.NewRepository(db, rdb)
	limitSvc := limit.NewService(limitRepo)

	accountRepo := account.NewRepository(db)
	accountSvc := account.NewService(accountRepo, auditSvc)

	transferRepo := transfer.NewRepository(db)
	transferSvc := transfer.NewService(transferRepo, accountSvc, auditSvc, limitSvc)

	authRepo := auth.NewRepository(db, rdb)
	authSvc := auth.NewService(authRepo, auditSvc)

	reconRepo := recon.NewRepository(db)
	reconSvc := recon.NewService(reconRepo, accountRepo, rdb)
	reconSvc.StartScheduler()

	reportSvc := report.NewService(db)

	interestSvc := interest.NewService(db, accountSvc)
	interestSvc.StartScheduler()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		auth.RegisterRoutes(api, authSvc)

		authorized := api.Group("/")
		authorized.Use(auth.AuthMiddleware(authSvc))
		authorized.Use(auditMiddleware.Audit())
		{
			auth.RegisterProtectedRoutes(authorized, authSvc)
			account.RegisterRoutes(authorized, accountSvc)
			transfer.RegisterRoutes(authorized, transferSvc)
			limit.RegisterRoutes(authorized, limitSvc)
			audit.RegisterRoutes(authorized, auditSvc)
			recon.RegisterRoutes(authorized, reconSvc)
			report.RegisterRoutes(authorized, reportSvc)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
