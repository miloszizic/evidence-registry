package api

import (
	"context"
	"errors"
	"evidence/internal/data"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/natefinch/lumberjack"
)

type Application struct {
	logger     *zap.SugaredLogger
	tokenMaker Maker
	config     data.Config
	stores     data.Stores
	wg         sync.WaitGroup
}

func Run() {
	conf, output, err := data.ParseFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println("got error:", err)
		fmt.Println("output:\n", output)
		os.Exit(1)
	}
	settings, err := data.LoadProductionConfig(conf.Path)
	if err != nil {
		fmt.Println("got error:", err)
	}
	app, err := NewApplication(settings)
	if err != nil {
		app.logger.Fatal("failed to create application", zap.Error(err))
	}
	err = app.Serve()
	if err != nil {
		app.logger.Fatal("failed to serve", zap.Error(err))
	}

}

func NewApplication(config data.Config) (*Application, error) {
	logger := InitLogger()
	defer logger.Sync()
	tokenMaker, err := NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create paseto maker for tokens: %w", err)
	}
	db, err := data.FromPostgresDB(config.Database.ConnectionInfo())
	if err != nil {
		logger.Fatal("calling database failed", zap.Error(err))
	}
	minioConfig := config.Minio
	minioClient, err := data.FromMinio(
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
	)
	if err != nil {
		logger.Fatal("calling minio failed", zap.Error(err))
	}
	app := &Application{
		logger:     logger,
		tokenMaker: tokenMaker,
		config:     config,
		stores:     data.NewStores(db, minioClient),
		//minioConfig:      minioClient,
	}
	//add default user
	user := &data.User{
		Username: "Simba",
	}
	err = user.Password.Set("opsAdmin")
	if err != nil {
		return nil, err
	}
	err = app.stores.User.Add(user)
	if err != nil {
		return nil, err
	}
	return app, nil

}

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

		app.logger.Info("caught signal", zap.String("signal", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", zap.String("addr", srv.Addr))

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting background tasks", zap.String("addr", srv.Addr), zap.String("env", app.config.Env))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", zap.String("addr", srv.Addr))

	return nil
}

func InitLogger() *zap.SugaredLogger {
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
