package main

import (
	"context"
	"github.com/pistolricks/models/cmd/models"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

func (app *application) contextClearUser(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, nil)
	return r.WithContext(ctx)
}
