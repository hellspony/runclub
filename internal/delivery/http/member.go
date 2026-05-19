package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// MemberHandler handles member endpoints.
type MemberHandler struct {
	memberUC usecase.MemberUseCase
}

// NewMemberHandler creates a new MemberHandler.
func NewMemberHandler(memberUC usecase.MemberUseCase) *MemberHandler {
	return &MemberHandler{memberUC: memberUC}
}

type createMemberRequest struct {
	FIO              string  `json:"fio"`
	TelegramUsername string  `json:"telegram_username"`
	TelegramID       int64   `json:"telegram_id"`
	BirthDate        *string `json:"birth_date"`
	Role             string  `json:"role"`
}

type updateMemberRequest struct {
	FIO              *string `json:"fio"`
	TelegramUsername *string `json:"telegram_username"`
	BirthDate        *string `json:"birth_date"`
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

type memberResponse struct {
	ID               int64     `json:"id"`
	FIO              string    `json:"fio"`
	TelegramUsername string    `json:"telegram_username"`
	TelegramID       int64     `json:"telegram_id"`
	BirthDate        *string   `json:"birth_date,omitempty"`
	Role             string    `json:"role,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func memberToResponse(m *entity.Member) memberResponse {
	resp := memberResponse{
		ID:               m.ID,
		FIO:              m.FIO,
		TelegramUsername: m.TelegramUsername,
		TelegramID:       m.TelegramID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.BirthDate != nil {
		s := m.BirthDate.Format(time.RFC3339)
		resp.BirthDate = &s
	}
	return resp
}

// List returns all members in a club.
func (h *MemberHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	members, err := h.memberUC.ListMembers(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list members"})
	}

	resp := make([]memberResponse, len(members))
	for i := range members {
		resp[i] = memberToResponse(&members[i])
		// Attach role from club_members join.
		cm, cmErr := h.memberUC.GetClubMember(c.Request().Context(), clubID, members[i].ID)
		if cmErr == nil && cm != nil {
			resp[i].Role = string(cm.Role)
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new member and adds them to a club.
func (h *MemberHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createMemberRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.FIO == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "fio is required"})
	}

	role := entity.MemberRole(req.Role)
	if role == "" {
		role = entity.RoleMember
	}

	member := &entity.Member{
		FIO:              req.FIO,
		TelegramUsername: req.TelegramUsername,
		TelegramID:       req.TelegramID,
	}

	memberID, err := h.memberUC.CreateMember(c.Request().Context(), member)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create member"})
	}
	member.ID = memberID

	// Update birth date if provided.
	if req.BirthDate != nil {
		var bd time.Time
		bd, err = time.Parse(time.RFC3339, *req.BirthDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid birth_date format, use RFC3339"})
		}
		err = h.memberUC.UpdateProfile(c.Request().Context(), member.ID, member.FIO, member.TelegramUsername, &bd)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update member profile"})
		}
	}

	// Refresh member after update.
	member, err = h.memberUC.GetByID(c.Request().Context(), member.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to get member"})
	}

	// Add member to the club with the specified role.
	err = h.memberUC.AddToClub(c.Request().Context(), clubID, member.ID, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to add member to club"})
	}

	resp := memberToResponse(member)
	resp.Role = string(role)
	return c.JSON(http.StatusCreated, resp)
}

// Get returns a single member by ID.
func (h *MemberHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidMemberID})
	}

	member, err := h.memberUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "member not found"})
	}

	return c.JSON(http.StatusOK, memberToResponse(member))
}

// Update updates an existing member's profile.
func (h *MemberHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidMemberID})
	}

	var req updateMemberRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	member, err := h.memberUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "member not found"})
	}

	fio := member.FIO
	if req.FIO != nil {
		fio = *req.FIO
	}

	telegramUsername := member.TelegramUsername
	if req.TelegramUsername != nil {
		telegramUsername = *req.TelegramUsername
	}

	var birthDate *time.Time
	if req.BirthDate != nil {
		var bd time.Time
		bd, err = time.Parse(time.RFC3339, *req.BirthDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid birth_date format, use RFC3339"})
		}
		birthDate = &bd
	} else {
		birthDate = member.BirthDate
	}

	err = h.memberUC.UpdateProfile(c.Request().Context(), id, fio, telegramUsername, birthDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update member"})
	}

	member, err = h.memberUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to get member"})
	}

	return c.JSON(http.StatusOK, memberToResponse(member))
}

// Delete removes a member by ID.
func (h *MemberHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidMemberID})
	}

	err = h.memberUC.DeleteMember(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete member"})
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateRole updates a member's role within a club.
func (h *MemberHandler) UpdateRole(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	memberID, err := strconv.ParseInt(c.Param("memberId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidMemberID})
	}

	var req updateRoleRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	role := entity.MemberRole(req.Role)
	if role != entity.RoleMember && role != entity.RoleTrainer && role != entity.RoleAdmin {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{jsonKeyError: "invalid role, must be one of: member, trainer, admin"},
		)
	}

	err = h.memberUC.UpdateRole(c.Request().Context(), clubID, memberID, role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update member role"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "role updated"})
}
