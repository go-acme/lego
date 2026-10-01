package internal

import "context"

type token string

const tokenKey token = "token"

func WithContext(ctx context.Context, credential string) context.Context {
	return context.WithValue(ctx, tokenKey, credential)
}

func getToken(ctx context.Context) string {
	credential, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return credential
}
