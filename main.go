package main

import (
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/kv"
)

func main() {
	// session
	http.HandleFunc("POST /dev/session/new", handJson(Session{}, SessionNew))
	http.HandleFunc("POST /dev/session/verify", handJson(Session{}, SessionVerify))

	workers.Serve(nil)
}

func handJson[reqT any, resT any](session Session, handler func(*http.Request, reqT, Session) (resT, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req reqT
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		res, err := handler(r, req, session)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func handAuth[reqT any, resT any](handler func(*http.Request, reqT, Session) (resT, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		var claims SessionClaims
		token, err := jwt.ParseWithClaims(r.Header.Get("Authorization"), &claims, func(t *jwt.Token) (any, error) { return GetJwtKey("session") })
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if !token.Valid {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		sessions, err := kv.NewNamespace(SessionNamespace)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		sessionReader, err := sessions.GetReader(claims.Subject, nil)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		var session Session
		err = gob.NewDecoder(sessionReader).Decode(&session)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		var req reqT
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		handJson(session, handler)
	}
}
