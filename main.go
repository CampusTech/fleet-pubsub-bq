package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/campus-it/fleet-pubsub-bq/internal/ingest"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	bqProjectID := requireEnv("BQ_PROJECT_ID")
	datasetID   := requireEnv("BQ_DATASET_ID")
	resultSub   := requireEnv("RESULT_SUBSCRIPTION")
	statusSub   := requireEnv("STATUS_SUBSCRIPTION")
	auditSub    := requireEnv("AUDIT_SUBSCRIPTION")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	h, err := ingest.NewHandler(ctx, bqProjectID, datasetID, resultSub, statusSub, auditSub, log)
	if err != nil {
		log.WithError(err).Fatal("failed to create ingest handler")
	}
	defer h.Close()

	srv := ingest.NewServer(h, resultSub, statusSub, auditSub, log)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	log.WithField("port", port).Info("starting fleet-pubsub-bq")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.WithError(err).Fatal("server exited")
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		logrus.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
