package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reindexer-service/config"
	"reindexer-service/internal/cache"
	"reindexer-service/internal/httpapi"
	"reindexer-service/internal/model"
	"reindexer-service/internal/repo"
	"reindexer-service/internal/service"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/restream/reindexer"
	_ "github.com/restream/reindexer/bindings/cproto"
)

func main() {
	cfg, infcfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	log.Printf("CONFIG SOURCES:")
	log.Printf("HTTP_ADDR: %s", infcfg.HTTPAddr) //для проверки в консоли какой конфиг взяли default, yaml, env
	log.Printf("REINDEXER_DSN: %s", infcfg.ReindexerDSN)

	log.Printf("config: http=%s dsn=%s ns=%s cache_ttl=%s",
		cfg.HTTP.Addr, cfg.Reindexer.DSN, cfg.Reindexer.Namespace, cfg.Cache.TTL)

	db := reindexer.NewReindex(
		cfg.Reindexer.DSN,
		reindexer.WithCreateDBIfMissing(),
		reindexer.WithConnPoolSize(8),
	)
	defer db.Close()

	if err := waitForReindexer(db, 10*time.Second); err != nil {
		log.Fatalf("reindexer not connected: %v", err)
	}
	log.Printf("reindexer OK: %s", cfg.Reindexer.DSN)

	if err := ensureNamespaces(db, cfg); err != nil {
		log.Fatalf("no necessary collections: %v", err)
	}
	log.Printf("namespace made: %s", cfg.Reindexer.Namespace)

	if err := db.Ping(); err != nil {
		log.Fatalf("reindexer ping failed: %v (dsn=%s)", err, cfg.Reindexer.DSN)
	}
	log.Printf("connected to reindexer: %s", cfg.Reindexer.DSN)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	docRepo := repo.NewDocumentRepo(db, cfg.Reindexer.Namespace)
	docCache := cache.NewRedisDocCache(rdb, cfg.Cache.TTL, cfg.Redis.Prefix)
	docSvc := service.NewDocumentService(docRepo, docCache)

	h := httpapi.NewHandler(docSvc)

	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func waitForReindexer(db *reindexer.Reindexer, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("ping timeout after %s: %w", timeout, lastErr)
}

func ensureNamespaces(db *reindexer.Reindexer, cfg config.Config) error {

	return db.OpenNamespace(cfg.Reindexer.Namespace, reindexer.DefaultNamespaceOptions(), model.Document{})
}
