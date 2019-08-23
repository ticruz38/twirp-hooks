package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/twitchtv/twirp"
	jwtDecode "gopkg.in/square/go-jose.v2/jwt"
)

var jwtCtxKey = new(int)

// UserKey to be used to retriev the current user from the context
var UserKey = new(int)

// WithJWT Extract JWT from request headers
func WithJWT(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt := r.Header.Get("Authorization")
		// jwt := getJWT(r) // Pulls the jwt out of the Authorization header, or something
		ctx := r.Context()
		if jwt != "" {
			ctx = context.WithValue(ctx, jwtCtxKey, jwt)
		}
		r = r.WithContext(ctx)
		base.ServeHTTP(w, r)
	})
}

func authorized(jwt string, svc string, method string) bool {
	fmt.Println("Implement here a function that check which svc method should be restricted or not")
	return true
}

// RoleResolver Mapping between role code number and actual role meaning
var RoleResolver = map[string]string{
	"0":    "System",
	"35":   "ManCo",
	"36":   "User",
	"3647": "UserSupervisor",
	"37":   "TA",
	"3747": "TASupervisor",
	"38":   "SubTA",
	"39":   "OrderRouter",
	"47":   "Supervisor",
}

func getUserFromJWT(jwt string) (*AuthenticatedUser, error) {
	claims := make(map[string]interface{}) // generic map to store parsed token
	// decode JWT token without verifying the signature
	token, err := jwtDecode.ParseSigned(jwt)
	if err != nil {
		fmt.Printf("Could not parse the token: %v", err)
		return nil, err
	}
	err = token.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		fmt.Printf("Could not read the token: %v", err)
		return nil, err
	}
	return &AuthenticatedUser{
		username:        claims["username"].(string),
		uid:             claims["uid"].(int),
		roles:           claims["roles"].([]string),
		contractAddress: claims["contractAddress"].(string),
		accountAddress:  claims["accountAddress"].(string),
		clientID:        claims["clientId"].(int),
	}, nil
}

// JWTCheckerHooks Parse the user from the JWT
func JWTCheckerHooks() *twirp.ServerHooks {
	hooks := &twirp.ServerHooks{}
	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {
		jwt, ok := ctx.Value(jwtCtxKey).(string)
		if !ok {
			// jwt missing: either middleware wasn't run, or it wasnt in the req
			return nil, twirp.NewError(twirp.Unauthenticated, "jwt must be included")
		}
		svc, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)
		if !authorized(jwt, svc, method) {
			return ctx, twirp.NewError(twirp.PermissionDenied, "not authorized to use this method")
		}
		// parse the token and add the current user to the context
		user, err := getUserFromJWT(jwt)
		if err != nil {
			ctx = context.WithValue(ctx, UserKey, user)
		}
		return ctx, nil
	}
	return hooks
}
