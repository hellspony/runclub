package http

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// ClubHandler handles club endpoints.
type ClubHandler struct {
	clubUC usecase.ClubUseCase
	authUC usecase.AuthUseCase
}

// NewClubHandler creates a new ClubHandler.
func NewClubHandler(clubUC usecase.ClubUseCase, authUC usecase.AuthUseCase) *ClubHandler {
	return &ClubHandler{clubUC: clubUC, authUC: authUC}
}

type createClubRequest struct {
	Name              string `json:"name"`
	TelegramChatID    int64  `json:"telegram_chat_id"`
	WelcomeEnabled    bool   `json:"welcome_enabled"`
	BirthdayEnabled   bool   `json:"birthday_enabled"`
	RaceNotifyEnabled bool   `json:"race_notify_enabled"`
}

type updateClubRequest struct {
	Name              *string `json:"name"`
	TelegramChatID    *int64  `json:"telegram_chat_id"`
	WelcomeEnabled    *bool   `json:"welcome_enabled"`
	BirthdayEnabled   *bool   `json:"birthday_enabled"`
	RaceNotifyEnabled *bool   `json:"race_notify_enabled"`
}

type clubResponse struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	TelegramChatID    int64  `json:"telegram_chat_id"`
	WelcomeEnabled    bool   `json:"welcome_enabled"`
	BirthdayEnabled   bool   `json:"birthday_enabled"`
	RaceNotifyEnabled bool   `json:"race_notify_enabled"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func clubToResponse(c *entity.Club) clubResponse {
	return clubResponse{
		ID:                c.ID,
		Name:              c.Name,
		TelegramChatID:    c.TelegramChatID,
		WelcomeEnabled:    c.WelcomeEnabled,
		BirthdayEnabled:   c.BirthdayEnabled,
		RaceNotifyEnabled: c.RaceNotifyEnabled,
		CreatedAt:         c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// List returns clubs. Superadmin sees all, admin sees only their assigned clubs.
func (h *ClubHandler) List(c echo.Context) error {
	if isSuperAdmin(c) {
		clubs, err := h.clubUC.List(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list clubs"})
		}
		resp := make([]clubResponse, len(clubs))
		for i := range clubs {
			resp[i] = clubToResponse(&clubs[i])
		}
		return c.JSON(http.StatusOK, resp)
	}

	// Admin user: only clubs they're assigned to
	userID := getUserIDFromContext(c)
	clubIDs, err := h.authUC.GetAdminUserClubs(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list clubs"})
	}

	resp := make([]clubResponse, 0, len(clubIDs))
	for _, cid := range clubIDs {
		var club *entity.Club
		club, err = h.clubUC.GetByID(c.Request().Context(), cid)
		if err != nil {
			continue
		}
		resp = append(resp, clubToResponse(club))
	}
	return c.JSON(http.StatusOK, resp)
}

// Create creates a new club. Only superadmin can create clubs.
func (h *ClubHandler) Create(c echo.Context) error {
	var req createClubRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errNameRequired})
	}

	club := &entity.Club{
		Name:              req.Name,
		TelegramChatID:    req.TelegramChatID,
		WelcomeEnabled:    req.WelcomeEnabled,
		BirthdayEnabled:   req.BirthdayEnabled,
		RaceNotifyEnabled: req.RaceNotifyEnabled,
	}

	id, err := h.clubUC.Create(c.Request().Context(), club)
	if err != nil {
		if isUniqueViolation(err) {
			return c.JSON(http.StatusConflict, echo.Map{jsonKeyError: "club with this telegram_chat_id already exists"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create club"})
	}

	club.ID = id

	// Auto-assign the creating superadmin to the new club
	_ = h.authUC.AssignClub(c.Request().Context(), getUserIDFromContext(c), id)

	return c.JSON(http.StatusCreated, clubToResponse(club))
}

// Get returns a single club by ID.
func (h *ClubHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	err = h.checkClubAccess(c, id)
	if err != nil {
		return err
	}

	club, err := h.clubUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "club not found"})
	}

	return c.JSON(http.StatusOK, clubToResponse(club))
}

// Update updates an existing club.
func (h *ClubHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	err = h.checkClubAccess(c, id)
	if err != nil {
		return err
	}

	club, err := h.clubUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "club not found"})
	}

	var req updateClubRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Name != nil {
		club.Name = *req.Name
	}
	if req.TelegramChatID != nil {
		club.TelegramChatID = *req.TelegramChatID
	}
	if req.WelcomeEnabled != nil {
		club.WelcomeEnabled = *req.WelcomeEnabled
	}
	if req.BirthdayEnabled != nil {
		club.BirthdayEnabled = *req.BirthdayEnabled
	}
	if req.RaceNotifyEnabled != nil {
		club.RaceNotifyEnabled = *req.RaceNotifyEnabled
	}

	err = h.clubUC.Update(c.Request().Context(), club)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update club"})
	}

	return c.JSON(http.StatusOK, clubToResponse(club))
}

// Delete removes a club by ID.
func (h *ClubHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	err = h.checkClubAccess(c, id)
	if err != nil {
		return err
	}

	err = h.clubUC.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete club"})
	}

	return c.NoContent(http.StatusNoContent)
}

// checkClubAccess returns nil if the user can access the club, otherwise an error response.
func (h *ClubHandler) checkClubAccess(c echo.Context, clubID int64) error {
	if isSuperAdmin(c) {
		return nil
	}
	userID := getUserIDFromContext(c)
	clubIDs, err := h.authUC.GetAdminUserClubs(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to check access"})
	}
	if slices.Contains(clubIDs, clubID) {
		return nil
	}
	return c.JSON(http.StatusForbidden, echo.Map{jsonKeyError: "access denied to this club"})
}
