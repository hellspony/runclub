package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// JointRunHandler handles joint run endpoints.
type JointRunHandler struct {
	jointRunUC usecase.JointRunUseCase
}

// NewJointRunHandler creates a new JointRunHandler.
func NewJointRunHandler(jointRunUC usecase.JointRunUseCase) *JointRunHandler {
	return &JointRunHandler{jointRunUC: jointRunUC}
}

type createJointRunRequest struct {
	LocationID int64     `json:"location_id"`
	CreatorID  int64     `json:"creator_id"`
	Date       time.Time `json:"date"`
}

type jointRunResponse struct {
	ID         int64     `json:"id"`
	ClubID     int64     `json:"club_id"`
	LocationID int64     `json:"location_id"`
	CreatorID  int64     `json:"creator_id"`
	Date       time.Time `json:"date"`
	MessageID  int64     `json:"message_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func jointRunToResponse(j *entity.JointRun) jointRunResponse {
	return jointRunResponse{
		ID:         j.ID,
		ClubID:     j.ClubID,
		LocationID: j.LocationID,
		CreatorID:  j.CreatorID,
		Date:       j.Date,
		MessageID:  j.MessageID,
		CreatedAt:  j.CreatedAt,
		UpdatedAt:  j.UpdatedAt,
	}
}

// List returns all joint runs for a club.
func (h *JointRunHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	runs, err := h.jointRunUC.ListJointRuns(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list joint runs"})
	}

	resp := make([]jointRunResponse, len(runs))
	for i := range runs {
		resp[i] = jointRunToResponse(&runs[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new joint run.
func (h *JointRunHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createJointRunRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	run, err := h.jointRunUC.CreateJointRun(
		c.Request().Context(),
		clubID,
		req.LocationID,
		req.CreatorID,
		req.Date,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create joint run"})
	}

	return c.JSON(http.StatusCreated, jointRunToResponse(run))
}

// Get returns a single joint run by ID.
func (h *JointRunHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid joint run id"})
	}

	run, err := h.jointRunUC.GetJointRun(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "joint run not found"})
	}

	return c.JSON(http.StatusOK, jointRunToResponse(run))
}

// Delete removes a joint run by ID.
func (h *JointRunHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid joint run id"})
	}

	err = h.jointRunUC.DeleteJointRun(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete joint run"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListParticipants returns all participants for a joint run as members.
func (h *JointRunHandler) ListParticipants(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid joint run id"})
	}

	members, err := h.jointRunUC.ListParticipants(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list participants"})
	}

	resp := make([]memberResponse, len(members))
	for i := range members {
		resp[i] = memberToResponse(&members[i])
	}

	return c.JSON(http.StatusOK, resp)
}
