package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type authService struct {
	TraggoURL string
}

type loginRequest struct {
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
	Query         string         `json:"query"`
}

type loginResponse struct {
	Data struct {
		Login struct {
			Token string `json:"token"`
		} `json:"login"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (s *authService) Authenticate(username, password string) (string, error) {
	body := loginRequest{
		OperationName: "Login",
		Variables: map[string]any{
			"name": username,
			"pass": password,
		},
		Query: `mutation Login($name: String!, $pass: String!) {
			login(username: $name, pass: $pass, deviceName: "test", type: NoExpiry, cookie: false) {
				token
			}
		}`,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(s.TraggoURL+"/graphql", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	var result loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	if result.Data.Login.Token == "" {
		return "", fmt.Errorf("empty token in response")
	}

	return result.Data.Login.Token, nil
}
