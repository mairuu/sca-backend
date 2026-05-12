package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	buckethandler "github.com/mairuu/mp-api/internal/features/bucket/handler"
	bucket "github.com/mairuu/mp-api/internal/features/bucket/model"
	bucketservice "github.com/mairuu/mp-api/internal/features/bucket/service"
	historyhandler "github.com/mairuu/mp-api/internal/features/history/handler"
	historyservice "github.com/mairuu/mp-api/internal/features/history/service"
	libraryhandler "github.com/mairuu/mp-api/internal/features/library/handler"
	library "github.com/mairuu/mp-api/internal/features/library/model"
	libraryservice "github.com/mairuu/mp-api/internal/features/library/service"
	mangahandler "github.com/mairuu/mp-api/internal/features/manga/handler"
	manga "github.com/mairuu/mp-api/internal/features/manga/model"
	mangaservice "github.com/mairuu/mp-api/internal/features/manga/service"
	userhandler "github.com/mairuu/mp-api/internal/features/user/handler"
	user "github.com/mairuu/mp-api/internal/features/user/model"
	userservice "github.com/mairuu/mp-api/internal/features/user/service"
	"github.com/mairuu/mp-api/internal/persistence/repositories"
	"github.com/mairuu/mp-api/internal/platform/authentication"
	"github.com/mairuu/mp-api/internal/platform/authorization"
	"github.com/mairuu/mp-api/internal/platform/config"
	"github.com/mairuu/mp-api/internal/platform/database"
	"github.com/mairuu/mp-api/internal/platform/logging"
	"github.com/mairuu/mp-api/internal/platform/scheduler"
	"github.com/mairuu/mp-api/internal/platform/storage"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/handler"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	log := logging.New(cfg.App.LogLevel)

	db, err := database.NewClient(&cfg.DB, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		panic(err)
	}
	log.Info("connected to database")

	publicBucket, err := storage.NewBucket(&cfg.Storage.PublicBucket)
	if err != nil {
		log.Error("failed to create public bucket", "error", err)
		panic(err)
	}
	log.Info("public storage backend", "type", cfg.Storage.PublicBucket.StorageType)

	temporaryBucket, err := storage.NewBucket(&cfg.Storage.TemporaryBucket)
	if err != nil {
		log.Error("failed to create temporary bucket", "error", err)
		panic(err)
	}
	log.Info("temporary storage backend", "type", cfg.Storage.TemporaryBucket.StorageType)

	enforcer, err := authorization.NewEnforcer()
	if err != nil {
		log.Error("failed to initialize authorization enforcer", "error", err)
		panic(err)
	}
	log.Info("authorization enforcer initialized")

	err = enforcer.AddPolicies(
		bucket.AllPolicies(),
		user.AllPolicies(),
		manga.AllPolicies(),
		library.AllPolicies(),
	)
	if err != nil {
		log.Error("failed to add policies to enforcer", "error", err)
		panic(err)
	}

	tokenService := authentication.NewTokenService(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	userRepo := repositories.NewUserRepository(db)
	mangaRepo := repositories.NewMangaRepository(db)
	libraryRepo := repositories.NewLibraryRepository(db)
	historyRepo := repositories.NewHistoryRepository(db)

	bucketService := bucketservice.NewService(enforcer, temporaryBucket)
	userService := userservice.NewService(userRepo, tokenService, enforcer)
	mangaService := mangaservice.NewService(log, mangaRepo, enforcer, publicBucket, temporaryBucket)
	libraryService := libraryservice.NewService(libraryRepo)
	historyService := historyservice.NewService(log, historyRepo)

	r := gin.New()
	r.SetTrustedProxies(nil)

	r.Use(gin.Recovery())
	r.Use(middleware.CORS("*")) // todo: make configurable

	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	r.Use(middleware.TraceID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Auth(tokenService))

	router := httptransport.NewRouter(r, []httptransport.Handler{
		handler.NewHealthHandler(log),
		buckethandler.NewBucketHandler(log, bucketService),
		userhandler.NewUserHandler(log, userService),
		mangahandler.NewHandler(log, mangaService),
		libraryhandler.NewHandler(log, libraryService),
		historyhandler.NewHandler(log, historyService),
	})
	router.RegisterRoutes()

	// start background services
	ttl := cfg.Cleanup.TTL
	scheduler.Schedule(ctx, cfg.Cleanup.Interval, func(ctx context.Context) {
		bucketService.CleanupExpiredFiles(ctx, ttl)
	})
	scheduler.Schedule(ctx, cfg.Cleanup.Interval, func(ctx context.Context) {
		userService.CleanupExpiredTokens(ctx)
	})

	srv := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: r,
	}

	var wg sync.WaitGroup
	// graceful shutdown
	wg.Go(func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("failed to gracefully shutdown server", "error", err)
		}
		log.Info("server shutdown complete")
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Error("failed to close database connection", "error", err)
			}
		}
		log.Info("database connection closed")
	})

	log.Info("starting server", "addr", cfg.HTTP.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("failed to run server", "error", err)
	}
	wg.Wait()
}
