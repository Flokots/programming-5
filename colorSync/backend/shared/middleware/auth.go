package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/Flokots/programming-5/colorSync/shared/auth"
)

// Context keys for storing claims in request context
type contextKey string

const (
    // UserClaimsKey is used to store user JWT claims in request context
    UserClaimsKey contextKey = "user_claims"
    
    // ServiceClaimsKey is used to store service JWT claims in request context
    ServiceClaimsKey contextKey = "service_claims"
)

// RequireAuth middleware validates JWT token from Authorization header
// It adds user claims to request context if the token is valid
// Usage: http.HandleFunc("/protected", middleware.RequireAuth(handler))
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get Authorization header
        // Expected format: "Authorization: Bearer <token>"
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, `{"error": "Missing authorization token"}`, http.StatusUnauthorized)
            return
        }

        // Extract token from "Bearer <token>" format
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            // No "Bearer " prefix found
            http.Error(w, `{"error": "Invalid authorization format. Use: Bearer <token>"}`, http.StatusUnauthorized)
            return
        }

        // Verify JWT token using shared auth package
        claims, err := auth.VerifyUserToken(tokenString)
        if err != nil {
            // Token is invalid or expired
            http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
            return
        }

        // Token is valid! Add claims to request context
        // Next handlers can retrieve user info from context
        ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
        
        // Call next handler with updated context
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// RequireServiceAuth middleware validates service-to-service JWT token
// Used for Zero Trust architecture between microservices
// Usage: http.HandleFunc("/internal", middleware.RequireServiceAuth(handler))
func RequireServiceAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get X-Service-Token header
        // This is a custom header for service-to-service communication
        tokenString := r.Header.Get("X-Service-Token")
        if tokenString == "" {
            http.Error(w, `{"error": "Missing service authentication token"}`, http.StatusUnauthorized)
            return
        }

        // Verify service JWT token using shared auth package
        claims, err := auth.VerifyServiceToken(tokenString)
        if err != nil {
            // Service token is invalid or expired
            http.Error(w, `{"error": "Invalid service token"}`, http.StatusUnauthorized)
            return
        }

        // Token is valid! Add service claims to request context
        ctx := context.WithValue(r.Context(), ServiceClaimsKey, claims)
        
        // Call next handler with updated context
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// GetUserClaims extracts user JWT claims from request context
// Returns nil if no claims found (user not authenticated)
// Usage in handler:
//   claims := middleware.GetUserClaims(r)
//   if claims != nil {
//       userID := claims.UserID
//   }
func GetUserClaims(r *http.Request) *auth.UserClaims {
    // Try to get claims from context
    if claims, ok := r.Context().Value(UserClaimsKey).(*auth.UserClaims); ok {
        return claims
    }
    return nil
}

// GetServiceClaims extracts service JWT claims from request context
// Returns nil if no claims found (service not authenticated)
// Usage in handler:
//   claims := middleware.GetServiceClaims(r)
//   if claims != nil {
//       serviceName := claims.ServiceName
//   }
func GetServiceClaims(r *http.Request) *auth.ServiceClaims {
    // Try to get claims from context
    if claims, ok := r.Context().Value(ServiceClaimsKey).(*auth.ServiceClaims); ok {
        return claims
    }
    return nil
}