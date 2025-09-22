package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	gommonlog "github.com/labstack/gommon/log"
	"github.com/sony/gobreaker"
)

var (
	ErrHttpGenericMessage    = echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	ErrWrongCredentials      = echo.NewHTTPError(http.StatusUnauthorized, "username or password is invalid")
	ErrDependencyUnavailable = echo.NewHTTPError(http.StatusServiceUnavailable, "users-api unavailable, try again later")
	jwtSecret                = "myfancysecret"
)

func main() {
	hostport := ":" + os.Getenv("AUTH_API_PORT")
	userAPIAddress := os.Getenv("USERS_API_ADDRESS")

	if v := os.Getenv("JWT_SECRET"); v != "" {
		jwtSecret = v
	}

	httpClient := &http.Client{Timeout: 800 * time.Millisecond}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "users-api",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.Requests >= 10 && float64(c.TotalFailures)/float64(c.Requests) >= 0.5
		},
	})

	userService := UserService{
		Client:         httpClient,
		UserAPIAddress: userAPIAddress,
		AllowedUserHashes: map[string]interface{}{
			"admin_admin": nil,
			"johnd_foo":   nil,
			"janed_ddd":   nil,
		},
		cb: cb,
	}

	e := echo.New()
	e.Logger.SetLevel(gommonlog.INFO)

	if zipkinURL := os.Getenv("ZIPKIN_URL"); zipkinURL != "" {
		e.Logger.Infof("init tracing to Zipkin at %s", zipkinURL)
		if tracedMiddleware, tracedClient, err := initTracing(zipkinURL); err == nil {
			e.Use(echo.WrapMiddleware(tracedMiddleware))
			userService.Client = tracedClient // TracedClient implementa HTTPDoer
		} else {
			e.Logger.Infof("Zipkin tracer init failed: %s", err.Error())
		}
	} else {
		e.Logger.Infof("Zipkin URL was not provided, tracing is not initialised")
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/version", func(c echo.Context) error {
		return c.String(http.StatusOK, "Auth API, written in Go\n")
	})

	e.POST("/login", getLoginHandler(userService))

	e.Logger.Fatal(e.Start(hostport))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getLoginHandler(userService UserService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var requestData LoginRequest
		if err := json.NewDecoder(c.Request().Body).Decode(&requestData); err != nil {
			log.Printf("could not read credentials from POST body: %s", err.Error())
			return ErrHttpGenericMessage
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), 900*time.Millisecond)
		defer cancel()

		user, err := userService.Login(ctx, requestData.Username, requestData.Password)
		if err != nil {
			switch err {
			case ErrWrongCredentials:
				return ErrWrongCredentials
			case ErrDependencyUnavailable:
				return ErrDependencyUnavailable
			default:
				log.Printf("could not authorize user '%s': %s", requestData.Username, err.Error())
				return ErrHttpGenericMessage
			}
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["username"] = user.Username
		claims["firstname"] = user.FirstName
		claims["lastname"] = user.LastName
		claims["role"] = user.Role
		claims["exp"] = time.Now().Add(72 * time.Hour).Unix()

		t, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			log.Printf("could not generate a JWT token: %s", err.Error())
			return ErrHttpGenericMessage
		}

		return c.JSON(http.StatusOK, map[string]string{
			"accessToken": t,
		})
	}
}
