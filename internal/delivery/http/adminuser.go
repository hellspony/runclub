package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// AdminUserHandler handles admin user management endpoints.
type AdminUserHandler struct {
	authUC usecase.AuthUseCase
	clubUC usecase.ClubUseCase
}

// NewAdminUserHandler creates a new AdminUserHandler.
func NewAdminUserHandler(authUC usecase.AuthUseCase, clubUC usecase.ClubUseCase) *AdminUserHandler {
	return &AdminUserHandler{authUC: authUC, clubUC: clubUC}
}

type createAdminUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type adminUserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type adminUserClubResponse struct {
	ClubID   int64  `json:"club_id"`
	ClubName string `json:"club_name"`
}

func adminUserToResponse(u *entity.AdminUser) adminUserResponse {
	return adminUserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// List returns all admin users.
func (h *AdminUserHandler) List(c echo.Context) error {
	users, err := h.authUC.ListAdminUsers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list admin users"})
	}

	resp := make([]adminUserResponse, len(users))
	for i := range users {
		resp[i] = adminUserToResponse(&users[i])
	}
	return c.JSON(http.StatusOK, resp)
}

// Create creates a new admin user.
func (h *AdminUserHandler) Create(c echo.Context) error {
	var req createAdminUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "username and password are required"})
	}

	role := req.Role
	if role == "" {
		role = string(entity.AdminRoleAdmin)
	}
	if role != string(entity.AdminRoleSuperAdmin) && role != string(entity.AdminRoleAdmin) {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "role must be 'superadmin' or 'admin'"})
	}

	id, err := h.authUC.CreateAdminUser(c.Request().Context(), req.Username, req.Password, role)
	if err != nil {
		if isUniqueViolation(err) {
			return c.JSON(http.StatusConflict, echo.Map{jsonKeyError: "username already exists"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create admin user"})
	}

	return c.JSON(http.StatusCreated, adminUserResponse{ID: id, Username: req.Username, Role: role})
}

// Delete removes an admin user.
func (h *AdminUserHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidUserID})
	}

	// Prevent deleting yourself
	if id == getUserIDFromContext(c) {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "cannot delete yourself"})
	}

	err = h.authUC.DeleteAdminUser(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete admin user"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListClubs returns clubs assigned to an admin user.
func (h *AdminUserHandler) ListClubs(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidUserID})
	}

	clubIDs, err := h.authUC.GetAdminUserClubs(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list clubs"})
	}

	resp := make([]adminUserClubResponse, 0, len(clubIDs))
	for _, cid := range clubIDs {
		var club *entity.Club
		club, err = h.clubUC.GetByID(c.Request().Context(), cid)
		if err != nil {
			continue
		}
		resp = append(resp, adminUserClubResponse{ClubID: club.ID, ClubName: club.Name})
	}
	return c.JSON(http.StatusOK, resp)
}

// AssignClub assigns a club to an admin user.
func (h *AdminUserHandler) AssignClub(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidUserID})
	}
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	err = h.authUC.AssignClub(c.Request().Context(), userID, clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to assign club"})
	}
	return c.NoContent(http.StatusNoContent)
}

// UnassignClub removes a club from an admin user.
func (h *AdminUserHandler) UnassignClub(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidUserID})
	}
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	err = h.authUC.UnassignClub(c.Request().Context(), userID, clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to unassign club"})
	}
	return c.NoContent(http.StatusNoContent)
}
