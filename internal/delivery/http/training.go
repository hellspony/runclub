package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// TrainingHandler handles training endpoints.
type TrainingHandler struct {
	trainingUC usecase.TrainingUseCase
}

// NewTrainingHandler creates a new TrainingHandler.
func NewTrainingHandler(trainingUC usecase.TrainingUseCase) *TrainingHandler {
	return &TrainingHandler{trainingUC: trainingUC}
}

type createTrainingRequest struct {
	LocationID int64     `json:"location_id"`
	Date       time.Time `json:"date"`
	Duration   int       `json:"duration"`
	TrainerIDs []int64   `json:"trainer_ids"`
}

type updateTrainingRequest struct {
	LocationID  *int64     `json:"location_id"`
	Date        *time.Time `json:"date"`
	Duration    *int       `json:"duration"`
	Status      *string    `json:"status"`
	PhotoFileID *string    `json:"photo_file_id"`
	MessageID   *int64     `json:"message_id"`
}

type trainingResponse struct {
	ID          int64     `json:"id"`
	ClubID      int64     `json:"club_id"`
	LocationID  int64     `json:"location_id"`
	Date        time.Time `json:"date"`
	Duration    int       `json:"duration"`
	Status      string    `json:"status"`
	PhotoFileID string    `json:"photo_file_id"`
	MessageID   int64     `json:"message_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func trainingToResponse(t *entity.Training) trainingResponse {
	return trainingResponse{
		ID:          t.ID,
		ClubID:      t.ClubID,
		LocationID:  t.LocationID,
		Date:        t.Date,
		Duration:    t.Duration,
		Status:      string(t.Status),
		PhotoFileID: t.PhotoFileID,
		MessageID:   t.MessageID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// List returns all trainings for a club.
func (h *TrainingHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	trainings, err := h.trainingUC.ListTrainings(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list trainings"})
	}

	resp := make([]trainingResponse, len(trainings))
	for i := range trainings {
		resp[i] = trainingToResponse(&trainings[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new training.
func (h *TrainingHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createTrainingRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	training, err := h.trainingUC.CreateTraining(
		c.Request().Context(),
		clubID,
		req.LocationID,
		req.Date,
		req.Duration,
		req.TrainerIDs,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create training"})
	}

	return c.JSON(http.StatusCreated, trainingToResponse(training))
}

// Get returns a single training by ID.
func (h *TrainingHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidTrainingID})
	}

	training, err := h.trainingUC.GetTraining(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "training not found"})
	}

	return c.JSON(http.StatusOK, trainingToResponse(training))
}

// Update updates an existing training.
func (h *TrainingHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidTrainingID})
	}

	training, err := h.trainingUC.GetTraining(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "training not found"})
	}

	var req updateTrainingRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.LocationID != nil {
		training.LocationID = *req.LocationID
	}
	if req.Date != nil {
		training.Date = *req.Date
	}
	if req.Duration != nil {
		training.Duration = *req.Duration
	}
	if req.Status != nil {
		training.Status = entity.TrainingStatus(*req.Status)
	}
	if req.PhotoFileID != nil {
		training.PhotoFileID = *req.PhotoFileID
	}
	if req.MessageID != nil {
		training.MessageID = *req.MessageID
	}

	err = h.trainingUC.UpdateTraining(c.Request().Context(), training)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update training"})
	}

	return c.JSON(http.StatusOK, trainingToResponse(training))
}

// Delete removes a training by ID.
func (h *TrainingHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidTrainingID})
	}

	err = h.trainingUC.DeleteTraining(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete training"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListParticipants returns all participants for a training as members.
func (h *TrainingHandler) ListParticipants(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidTrainingID})
	}

	members, err := h.trainingUC.ListParticipants(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list participants"})
	}

	resp := make([]memberResponse, len(members))
	for i := range members {
		resp[i] = memberToResponse(&members[i])
	}

	return c.JSON(http.StatusOK, resp)
}
