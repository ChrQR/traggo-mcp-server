package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/ChrQR/traggo-mcp-server/internal/auth"
	"github.com/ChrQR/traggo-mcp-server/internal/mcp"
	"github.com/ChrQR/traggo-mcp-server/views/pages"
	"github.com/a-h/templ"
)

func main() {
	ctx := context.Background()

	authHandler := auth.NewAuthHandler()

	mcpServer := mcp.NewMcpServer("1.0.0")

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	mux.Handle("/", http.RedirectHandler("/login", http.StatusPermanentRedirect))
	mux.Handle("GET /login", templ.Handler(pages.Login()))
	mux.Handle("POST /login", authHandler.Login())
	mux.Handle("/mcp", mcp.AuthMiddleware(mcpServer.Handler))

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// db, err := shared.NewDB(ctx, "tmp/db.sqlite")
	// if err != nil {
	// 	slog.Error("unable to connect to db", "error", err.Error())
	// 	return
	// }

	// err = db.Ping()
	// if err != nil {
	// 	slog.Error("error pinging db", "error", err.Error())
	// }

	slog.Info("Starting server...")

	log.Fatal(srv.ListenAndServe())
}
