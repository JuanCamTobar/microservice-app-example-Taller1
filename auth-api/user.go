package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sony/gobreaker"
)

// Modelos
type User struct {
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Role      string `json:"role"`
}

// Cliente HTTP genérico para facilitar tests (mocks)
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Servicio
type UserService struct {
	Client            HTTPDoer
	UserAPIAddress    string
	AllowedUserHashes map[string]interface{}
	cb                *gobreaker.CircuitBreaker
}

func NewUserService(userAPI string, allowed map[string]interface{}) *UserService {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "users-api",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.Requests >= 10 && float64(c.TotalFailures)/float64(c.Requests) >= 0.5
		},
	})
	httpClient := &http.Client{Timeout: 2 * time.Second}

	return &UserService{
		Client:            httpClient,
		UserAPIAddress:    userAPI,
		AllowedUserHashes: allowed,
		cb:                cb,
	}
}

func (h *UserService) Login(ctx context.Context, username, password string) (User, error) {
	user, err := h.getUser(ctx, username)
	if err != nil {
		return user, err
	}

	userKey := fmt.Sprintf("%s_%s", username, password)
	if _, ok := h.AllowedUserHashes[userKey]; !ok {
		// Usamos ErrWrongCredentials de main.go (mismo paquete)
		return user, ErrWrongCredentials
	}

	return user, nil
}

func (h *UserService) getUser(ctx context.Context, username string) (User, error) {
	var user User

	token, err := h.getUserAPIToken(username)
	if err != nil {
		return user, err
	}

	url := fmt.Sprintf("%s/users/%s", h.UserAPIAddress, username)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	// Circuit Breaker alrededor de la llamada remota
	exec := func() (interface{}, error) {
		resp, err := h.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("users-api %d: %s", resp.StatusCode, string(body))
		}
		return body, nil
	}

	var result interface{}
	if h.cb != nil {
		result, err = h.cb.Execute(exec)
	} else {
		// Por si acaso no se inyectó cb
		result, err = exec()
	}
	if err != nil {
		// Devuelve la misma variable que usa main.go (503)
		return user, ErrDependencyUnavailable
	}

	if err := json.Unmarshal(result.([]byte), &user); err != nil {
		return user, err
	}
	return user, nil
}

func (h *UserService) getUserAPIToken(username string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)
	claims := t.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["scope"] = "read"
	return t.SignedString([]byte(jwtSecret))
}
