package core

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/danielboakye/filechangestracker/internal/commandexecutor"
	"github.com/danielboakye/filechangestracker/internal/config"
	"github.com/danielboakye/filechangestracker/internal/filechangestracker"
	"github.com/danielboakye/filechangestracker/internal/httpserver"
	"github.com/danielboakye/filechangestracker/internal/mongolog"
	"github.com/danielboakye/filechangestracker/pkg/osquerymanager"
	"github.com/osquery/osquery-go"
)

type App struct {
	ctx       context.Context
	cancel    context.CancelFunc
	apiServer *httpserver.Server
	executor  commandexecutor.CommandExecutor
	tracker   filechangestracker.FileChangesTracker
	logStore  mongolog.LogStore
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	a.ctx = ctx
	a.cancel = cancel

	cfg, err := config.LoadConfig(config.ConfigName, config.ConfigPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logStore, err := mongolog.NewMongoLogStore(a.ctx, cfg.MongoURI, config.LogsDBName, config.LogsCollectionName)
	if err != nil {
		log.Fatalf("failed to start mongo: %v", err)
	}

	appLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	executor := commandexecutor.New(appLogger, cfg)
	if err := executor.Start(a.ctx); err != nil {
		log.Fatalf("failed to start command executor: %v", err)
	}

	osqueryClient, err := osquery.NewClient(cfg.SocketPath, 10*time.Second)
	if err != nil {
		log.Fatalf("Error creating osquery client: %v", err)
	}
	osqueryManager := osquerymanager.New(osqueryClient)

	tracker := filechangestracker.New(appLogger, cfg, osqueryManager, logStore)
	if err := tracker.Start(a.ctx); err != nil {
		log.Fatalf("failed to start tracker: %v", err)
	}

	appLogger.Info("started-tracker-on-directory", slog.String("directory", cfg.Directory))

	handler := httpserver.NewHandler(tracker, executor)
	router := handler.RegisterRoutes()

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	apiServer := httpserver.NewServer(addr, appLogger, router)
	if err := apiServer.Start(); err != nil {
		log.Fatal("failed to start http server on: ", addr)
	}

	a.executor = executor
	a.tracker = tracker
	a.apiServer = apiServer
	a.logStore = logStore
}

func (a *App) Stop() {
	a.apiServer.Stop(a.ctx)
	a.executor.Stop(a.ctx)
	a.tracker.Stop(a.ctx)
	a.logStore.Close(a.ctx)
	a.cancel()
	fmt.Println("app stopped!")
}

func New() *App {
	return &App{}
}
