package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ChrQR/traggo-mcp-server/views/components"
)

type authHandler struct {
	authService authService
}

func NewAuthHandler() *authHandler {
	traggoURL := os.Getenv("TRAGGO_URL")
	return &authHandler{
		authService: authService{
			TraggoURL: traggoURL,
		},
	}
}

func (h *authHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to parse form: %s", err.Error()), http.StatusUnprocessableEntity)
		}
		userName := r.Form.Get("username")
		password := r.Form.Get("password")

		token, err := h.authService.Authenticate(userName, password)
		if err != nil {
			slog.Error("error fetching token", "error", err.Error())
			components.LoginForm(true, new("Invalid username or password")).Render(r.Context(), w)
			return
		}

		components.TokenCard(token).Render(r.Context(), w)

	}
}
