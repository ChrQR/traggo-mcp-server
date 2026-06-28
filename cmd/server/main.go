package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/ChrQR/traggo-mcp-server/internal/auth"
	"github.com/ChrQR/traggo-mcp-server/internal/mcp"
	"github.com/ChrQR/traggo-mcp-server/views/pages"
	"github.com/a-h/templ"
)

func main() {
	authHandler := auth.NewAuthHandler()

	traggoURL := os.Getenv("TRAGGO_URL")
	mcpServer := mcp.NewMcpServer("1.0.0", traggoURL)

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

	slog.Info("Starting server...")

	log.Fatal(srv.ListenAndServe())
}
