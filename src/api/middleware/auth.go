package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jamesjscully/un-tie_code/src/api/models"
)

// SessionMiddleware creates middleware that checks if the user is authenticated
// and sets the user in the context if they are
func SessionMiddleware(authService models.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := c.Get("traceID")
		path := c.Request.URL.Path
		
		fmt.Printf("[%s] SessionMiddleware: Checking session for path %s\n", traceID, path)
		
		// Get session token from cookie
		sessionToken, err := c.Cookie("session_token")
		if err != nil {
			fmt.Printf("[%s] SessionMiddleware: No session token found for path %s: %v\n", traceID, path, err)
			// No session token, continue as unauthenticated
			c.Set("authenticated", false)
			c.Next()
			return
		}
		
		fmt.Printf("[%s] SessionMiddleware: Found session token for path %s\n", traceID, path)
		
		// Verify session
		user, err := authService.VerifySession(sessionToken)
		if err != nil {
			fmt.Printf("[%s] SessionMiddleware: Invalid session for path %s: %v\n", traceID, path, err)
			// Invalid session, clear cookie and continue as unauthenticated
			c.SetCookie("session_token", "", -1, "/", "", false, true)
			c.Set("authenticated", false)
			c.Next()
			return
		}
		
		// Valid session, set user in context
		fmt.Printf("[%s] SessionMiddleware: User authenticated for path %s: %s (%s)\n", 
			traceID, path, user.ID, user.Email)
		c.Set("user", user)
		c.Set("authenticated", true)
		
		c.Next()
	}
}

// RequireAuth creates middleware that requires authentication
// If the user is not authenticated, they will be redirected to the login page
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID, _ := c.Get("traceID")
		path := c.Request.URL.Path
		
		fmt.Printf("[%s] RequireAuth: Checking authentication for path %s\n", traceID, path)
		
		authenticated, exists := c.Get("authenticated")
		if !exists {
			fmt.Printf("[%s] RequireAuth: 'authenticated' not set in context for path %s\n", traceID, path)
			authenticated = false
		}
		
		fmt.Printf("[%s] RequireAuth: authenticated=%v for path %s\n", traceID, authenticated, path)
		
		if authenticated != true {
			fmt.Printf("[%s] RequireAuth: Authentication required for path %s, redirecting to login\n", 
				traceID, path)
			
			// Store the original URL for redirection after login
			returnTo := c.Request.URL.String()
			c.SetCookie("return_to", returnTo, 300, "/", "", false, true) // 5 minute expiry
			
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}
		
		fmt.Printf("[%s] RequireAuth: Authentication verified for path %s\n", traceID, path)
		c.Next()
	}
}

// GetCurrentUser is a helper function to get the current authenticated user from the context
// Returns nil if no user is authenticated
func GetCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	
	if u, ok := user.(*models.User); ok {
		return u
	}
	
	return nil
}
