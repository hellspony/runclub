package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"runclub/internal/usecase"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authUC usecase.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type meResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login authenticates a user and returns a JWT token.
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "username and password are required"})
	}

	token, err := h.authUC.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{jsonKeyError: "invalid credentials"})
	}

	return c.JSON(http.StatusOK, loginResponse{Token: token})
}

// Me returns the current authenticated user info from the token.
func (h *AuthHandler) Me(c echo.Context) error {
	username, ok := c.Get(contextKeyUser).(string)
	if !ok || username == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{jsonKeyError: "unauthorized"})
	}

	userID := getUserIDFromContext(c)
	role := getRoleFromContext(c)

	return c.JSON(http.StatusOK, meResponse{
		ID:       userID,
		Username: username,
		Role:     string(role),
	})
}
