package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Relaxed CSP for trade pages that need TradingView widget
		if strings.HasPrefix(r.URL.Path, "/trade/") {
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://s3.tradingview.com https://*.tradingview.com; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src fonts.gstatic.com; frame-src https://*.tradingview.com; connect-src 'self' https://*.tradingview.com wss://*.tradingview.com; img-src 'self' data: https://*.tradingview.com")
		} else {
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src fonts.gstatic.com")
		}
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			ctx = context.WithValue(ctx, authenticatedUserIDKey, id) // Add userID too
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
