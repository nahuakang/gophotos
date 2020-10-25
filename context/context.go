package context

import (
	"context"

	"github.com/nahuakang/gophotos/models"
)

type privateKey string

const userKey privateKey = "user"

// WithUser returns a context.Context with the user information
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// User looks up a user from a given context.Context
func User(ctx context.Context) *models.User {
	// Verify that a user was previously stored in the context
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}
