package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

const (
	authHeaderKey    = "Authorization"
	bearerPrefix     = "Bearer "
	contextKeyUser   = "username"
	contextKeyRole   = "role"
	contextKeyUserID = "user_id"
)

// AuthMiddleware extracts the JWT token from the Authorization header,
// validates it via AuthUseCase, and stores user info in the echo context.
func AuthMiddleware(authUC usecase.AuthUseCase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(authHeaderKey)
			if header == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					jsonKeyError: "missing authorization header",
				})
			}

			if !strings.HasPrefix(header, bearerPrefix) {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					jsonKeyError: "invalid authorization format",
				})
			}

			token := strings.TrimPrefix(header, bearerPrefix)
			claims, err := authUC.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					jsonKeyError: "invalid or expired token",
				})
			}

			c.Set(contextKeyUser, claims.Username)
			c.Set(contextKeyRole, string(claims.Role))
			c.Set(contextKeyUserID, claims.UserID)
			return next(c)
		}
	}
}

// SuperAdminMiddleware blocks requests from non-superadmin users.
func SuperAdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(contextKeyRole).(string)
			if !ok || role != string(entity.AdminRoleSuperAdmin) {
				return c.JSON(http.StatusForbidden, echo.Map{
					jsonKeyError: "superadmin access required",
				})
			}
			return next(c)
		}
	}
}

// getRoleFromContext returns the role string from the echo context.
func getRoleFromContext(c echo.Context) entity.AdminRole {
	role, _ := c.Get(contextKeyRole).(string)
	return entity.AdminRole(role)
}

// getUserIDFromContext returns the user ID from the echo context.
func getUserIDFromContext(c echo.Context) int64 {
	id, _ := c.Get(contextKeyUserID).(int64)
	return id
}

// isSuperAdmin checks if the current user is a superadmin.
func isSuperAdmin(c echo.Context) bool {
	return getRoleFromContext(c) == entity.AdminRoleSuperAdmin
}
