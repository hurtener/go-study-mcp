// Command go-study-mcp is an MCP server for generating study audio content.
//
// It registers four tools (generate_podcast, generate_study_guide,
// generate_flashcards, synthesize_speech) and serves the MCP protocol
// with an inline UI.
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/hurtener/dockyard/runtime/apps"
	"github.com/hurtener/dockyard/runtime/server"
)

//go:embed all:web/dist
var uiBundle embed.FS

const (
	httpAddr = "127.0.0.1:8080"
	appURI   = "ui://go-study-mcp/studio/index.html"
	appName  = "studio"
	appTitle = "Study Audio Studio"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv, err := server.New(server.Info{
		Name:    "go-study-mcp",
		Title:   "Go Study MCP",
		Version: "0.4.0",
	}, &server.Options{Logger: logger})
	if err != nil {
		logger.Error("create server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := registerApp(srv); err != nil {
		logger.Error("register app", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := registerTools(srv); err != nil {
		logger.Error("register tools", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := serve(ctx, srv, logger); err != nil {
		logger.Error("serve", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// registerApp embeds the Svelte UI bundle and registers it with the server.
func registerApp(srv *server.Server) error {
	html, err := fs.ReadFile(uiBundle, "web/dist/index.html")
	if err != nil {
		return err
	}
	return apps.Register(srv, apps.App{
		URI:   appURI,
		Name:  appName,
		Title: appTitle,
		HTML:  html,
	})
}

func serve(ctx context.Context, srv *server.Server, logger *slog.Logger) error {
	switch transport := os.Getenv("DOCKYARD_TRANSPORT"); transport {
	case "", "stdio":
		return srv.ServeStdio(ctx)
	case "http":
		return serveHTTP(ctx, srv, logger)
	default:
		return errors.New("unsupported DOCKYARD_TRANSPORT " + transport + " (want \"stdio\" or \"http\")")
	}
}

func serveHTTP(ctx context.Context, srv *server.Server, logger *slog.Logger) error {
	handler, err := srv.HTTPHandler(nil)
	if err != nil {
		return err
	}
	addr := httpAddr
	if override := os.Getenv("DOCKYARD_HTTP_ADDR"); override != "" {
		addr = override
	}
	httpSrv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		_ = httpSrv.Close()
	}()
	logger.Info("serving streamable-HTTP transport", slog.String("addr", addr))
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
