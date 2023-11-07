package api

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/miloszizic/der/db"
	"github.com/miloszizic/der/vault"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"

	"github.com/miloszizic/der/service"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// The Application holds the application state.
type Application struct {
	logger     *zap.SugaredLogger
	tokenMaker service.Maker
	config     service.Config
	stores     service.Stores
	wg         sync.WaitGroup
}

// Run starts the application.
func Run() error {
	output, err := initializeApplication(os.Args[1:])
	if err != nil {
		return fmt.Errorf("initializing application: got error: %w, output: %v", err, output)
	}

	app, err := prepareApplication(output)
	if err != nil {
		return errors.Wrap(err, "failed to prepare application")
	}

	// Once app has been declared, defer the resource cleanup
	defer func() {
		if closeErr := app.CloseResources(); closeErr != nil {
			app.logger.Errorw("Error in closing resources: %v", closeErr)
		}

		if syncErr := app.logger.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", syncErr)
		}
	}()

	err = startServer(app)
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}

	return nil
}

func startServer(app *Application) error {
	err := app.Serve()
	if err != nil {
		app.logger.Errorw("Failed to serve", "error", err)
		return err
	}

	return nil
}

// CloseResources closes the application resources.
func (app *Application) CloseResources() error {
	if app.stores.DB != nil {
		if err := app.stores.DB.Close(); err != nil {
			app.logger.Errorw("Error in closing database connection", "error", err)
		}
	}
	// TODO: if minio client has close method add it here
	return nil
}

func initializeApplication(args []string) (service.Config, error) {
	conf, _, err := service.ParseFlags(os.Args[0], args)
	if errors.Is(err, flag.ErrHelp) {
		return service.Config{}, fmt.Errorf("help: %w", err)
	}

	var settings service.Config
	// Check if a configuration file path was provided.

	if conf.Path == "" {
		settings = service.LoadDefaultConfig()
	} else {
		settings, err = service.LoadProductionConfig(conf.Path)
		if err != nil {
			return service.Config{}, fmt.Errorf("loading configuration: %w", err)
		}
	}

	return settings, nil
}

func prepareApplication(config service.Config) (*Application, error) {
	logger := initLogger()

	tokenMaker, err := initTokenMaker(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token maker: %w", err)
	}

	dbService, err := initDBService(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DB service: %w", err)
	}

	minioClient, err := initMinioClient(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	app := &Application{
		logger:     logger,
		tokenMaker: tokenMaker,
		config:     config,
		stores:     service.NewStores(dbService, minioClient),
	}

	err = addUser(app)
	if err != nil {
		return nil, fmt.Errorf("failed to add user: %w", err)
	}

	return app, nil
}

func initDBService(config service.Config, logger *zap.SugaredLogger) (*sql.DB, error) {
	dbs, err := service.FromPostgresDB(config.Database.ConnectionInfo(), config.Database.Automigrate, config.Env)
	if err != nil {
		logger.Error("calling db failed", zap.Error(err))
		return nil, err
	}

	return dbs, nil
}

func initMinioClient(config service.Config, logger *zap.SugaredLogger) (*minio.Client, error) {
	minioConfig := config.Minio
	minioClient, err := service.FromMinio(
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
	)
	if err != nil {
		logger.Error("calling minio failed", zap.Error(err))
		return nil, err
	}

	return minioClient, nil
}

func initTokenMaker(config service.Config, logger *zap.SugaredLogger) (service.Maker, error) {
	tokenMaker, err := service.NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		logger.Error("failed to create paseto maker for tokens: ", zap.Error(err))
		return nil, err
	}

	return tokenMaker, nil
}

func initLogger() *zap.SugaredLogger {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
	)
	logger := zap.New(core, zap.AddCaller())
	sugarLogger := logger.Sugar()

	return sugarLogger
}

func addUser(app *Application) error {
	pass, err := vault.Hash("password")
	if err != nil {
		return err
	}
	id, err := app.stores.DBStore.GetRoleID(context.Background(), "admin")
	if err != nil {
		fmt.Printf("getting role id: %v , role: %q ", err, "admin")
		return err
	}

	// add default user
	user := db.CreateUserParams{
		Username: "Simba",
		Password: pass,
		RoleID:   uuid.NullUUID{UUID: id, Valid: true},
	}

	exists, err := app.stores.DBStore.UserExists(context.Background(), user.Username)
	if err != nil {
		fmt.Printf("checking user in DB store: %v , username: %q ", err, user.Username)
		return err
	}

	if exists {
		return nil
	}

	_, err = app.stores.DBStore.CreateUser(context.Background(), user)
	if err != nil {
		fmt.Printf("creating user in DB: %v , username: %q ", err, user.Username)
		return err
	}

	return nil
}

// Serve starts the HTTP server.
func (app *Application) Serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 80 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Infof("caught signal: %s", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Infof("completing background tasksat addr: %s with env: %s", srv.Addr, app.config.Env)

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Infof("starting background tasks at addr: %s with env: %s", srv.Addr, app.config.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Infof("stopped server at address: %s", srv.Addr)

	return nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./test.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}
