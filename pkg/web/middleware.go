package web

import (
	"auth/pkg/jwt"
	"auth/pkg/response"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const BASE_RATE_LIMIT = 60

var limiter = rate.NewLimiter(rate.Every(time.Minute), rateLimit())

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := strings.Split(r.RequestURI, "?")[0]

		if !server.NeedsAuth(uri) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		if server.GuestOnly(uri) {
			if authHeader != "" {
				log.Println("This route is for guests only.")
				res := response.New(w, r, "This route is for guests only.", http.StatusUnauthorized)
				res.Process()
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		if len(authHeader) < 1 {
			log.Println("Authentication header not provided.")
			res := response.New(w, r, "Authentication header not provided.", http.StatusUnauthorized)
			res.Process()
			return
		}

		auth := strings.TrimPrefix(authHeader, "Bearer ")

		if len(auth) < 1 {
			log.Println("Authentication header not provided.")
			res := response.New(w, r, "Authentication header not provided.", http.StatusUnauthorized)
			res.Process()
			return
		}

		token, err := jwt.New(auth)

		if err != nil {
			log.Println("Invalid token provided.")
			res := response.New(w, r, "Invalid token provided.", http.StatusUnauthorized)
			res.Process()
			return
		}

		if !token.Valid {
			log.Println("Invalid token provided.")
			res := response.New(w, r, "Invalid token provided.", http.StatusUnauthorized)
			res.Process()
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func rateLimit() int {
	if val := os.Getenv("RATE_LIMIT"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			return v
		}
	}

	return BASE_RATE_LIMIT
}
