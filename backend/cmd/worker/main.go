package main

import (
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/log"
)

func main() {
	cfg := config.FromEnv()
	logger := log.New(cfg.Env)
	defer func() { _ = logger.Sync() }()

	logger.Info("worker starting")

	// TODO Phase 3: attribution post-processing.
	// Читает raw events из ClickHouse, классифицирует hunks
	// (typed | pasted-AI | pasted-other | AI-inline | AI-agent | unknown),
	// пишет в attribution_events.

	select {} // блок до получения сигнала завершения
}
