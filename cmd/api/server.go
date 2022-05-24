package api

import (
	"context"
	"errors"
	"evidence/internal/data"
	"evidence/jsonlog"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type Application struct {
	logger     *jsonlog.Logger
	tokenMaker Maker
	config     data.Config
	stores     data.Stores
	wg         sync.WaitGroup
}

func NewApplication(config data.Config) (*Application, error) {
	logger := jsonlog.New(jsonlog.LevelInfo)
	tokenMaker, err := NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenMaker maker: %w", err)
	}
	db, err := data.FromPostgresDB(config.Database.ConnectionInfo())
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	minioConfig := config.Minio
	minioClient, err := data.FromMinio(
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
	)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	app := &Application{
		logger:     logger,
		tokenMaker: tokenMaker,
		config:     config,
		stores:     data.NewStores(db, minioClient),
		//minioConfig:      minioClient,
	}
	//TODO: remove after testing the minio UI
	//add default user
	user := &data.User{
		Username: "Simba",
	}
	user.Password.Set("opsAdmin")
	app.stores.UserDB.Add(user)
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

		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.Env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}

func Run() {
	boolPtr := flag.Bool("prod", false, "Provide this flag in production. This ensures that a .config.json file is provided before the Application starts.")
	flag.Parse()
	config, err := data.LoadProductionConfig(*boolPtr)
	if err != nil {
		log.Fatal(err)
	}
	app, err := NewApplication(config)
	if err != nil {
		log.Fatal(err)
	}
	err = app.Serve()
	if err != nil {
		log.Fatal(err)
	}
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
