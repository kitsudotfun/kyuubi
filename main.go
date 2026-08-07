package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/cron"
)

func main() {
	// clean up database nightly
	cron.ScheduleTaskNonBlock(func(ctx context.Context) error { return CleanServers() })

	// session
	http.HandleFunc("POST /dev/session/new", handJson(SessionNew, Session{}))
	http.HandleFunc("POST /dev/session/verify", handJson(SessionVerify, Session{}))

	// server
	http.HandleFunc("POST /dev/server/heartbeat", handAuth(ServerHeartbeat))
	http.HandleFunc("POST /dev/server/delete", handAuth(ServerDelete))

	http.HandleFunc("POST /dev/server/list", handAuth(ServerList))
	http.HandleFunc("POST /dev/server/join", handAuth(ServerJoin))

	workers.Serve(nil)
}

func handJson[reqT any, resT any](handler func(reqT, Session) (resT, error), session Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req reqT
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := handler(req, session)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handAuth[reqT any, resT any](handler func(reqT, Session) (resT, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var claims SessionClaims
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), &claims, func(t *jwt.Token) (any, error) { return MustGetJwtKey("session"), nil })
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var session Session
		session.ID.FromString(claims.Subject)
		session.GameID = claims.GameID

		handJson(handler, session).ServeHTTP(w, r)
	}
}
