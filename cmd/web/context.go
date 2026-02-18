package main

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	authenticatedUserIDKey    = contextKey("authenticatedUserID")
	userRoleContextKey        = contextKey("userRole")
)
