package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// RaceHandler handles race endpoints.
type RaceHandler struct {
	raceUC usecase.RaceUseCase
}

// NewRaceHandler creates a new RaceHandler.
func NewRaceHandler(raceUC usecase.RaceUseCase) *RaceHandler {
	return &RaceHandler{raceUC: raceUC}
}

type createRaceRequest struct {
	Date      time.Time `json:"date"`
	Type      string    `json:"type"`
	Place     string    `json:"place"`
	Distances string    `json:"distances"`
	Name      string    `json:"name"`
}

type updateRaceRequest struct {
	Date      *time.Time `json:"date"`
	Type      *string    `json:"type"`
	Place     *string    `json:"place"`
	Distances *string    `json:"distances"`
	Name      *string    `json:"name"`
}

type raceResponse struct {
	ID        int64     `json:"id"`
	ClubID    int64     `json:"club_id"`
	Date      time.Time `json:"date"`
	Type      string    `json:"type"`
	Place     string    `json:"place"`
	Distances string    `json:"distances"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type raceRegistrationResponse struct {
	ID        int64     `json:"id"`
	RaceID    int64     `json:"race_id"`
	MemberID  int64     `json:"member_id"`
	Distance  string    `json:"distance"`
	CreatedAt time.Time `json:"created_at"`
}

type registerMemberRequest struct {
	MemberID int64  `json:"member_id"`
	Distance string `json:"distance"`
}

type unregisterMemberRequest struct {
	MemberID int64 `json:"member_id"`
}

func raceToResponse(r *entity.Race) raceResponse {
	return raceResponse{
		ID:        r.ID,
		ClubID:    r.ClubID,
		Date:      r.Date,
		Type:      r.Type,
		Place:     r.Place,
		Distances: r.Distances,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func raceRegistrationToResponse(r *entity.RaceRegistration) raceRegistrationResponse {
	return raceRegistrationResponse{
		ID:        r.ID,
		RaceID:    r.RaceID,
		MemberID:  r.MemberID,
		Distance:  r.Distance,
		CreatedAt: r.CreatedAt,
	}
}

// List returns all races for a club.
func (h *RaceHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	races, err := h.raceUC.ListByClub(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list races"})
	}

	resp := make([]raceResponse, len(races))
	for i := range races {
		resp[i] = raceToResponse(&races[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new race.
func (h *RaceHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createRaceRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errNameRequired})
	}

	race := &entity.Race{
		ClubID:    clubID,
		Date:      req.Date,
		Type:      req.Type,
		Place:     req.Place,
		Distances: req.Distances,
		Name:      req.Name,
	}

	id, err := h.raceUC.Create(c.Request().Context(), race)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create race"})
	}

	race.ID = id
	return c.JSON(http.StatusCreated, raceToResponse(race))
}

// Get returns a single race by ID.
func (h *RaceHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	race, err := h.raceUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "race not found"})
	}

	return c.JSON(http.StatusOK, raceToResponse(race))
}

// Update updates an existing race.
func (h *RaceHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	race, err := h.raceUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "race not found"})
	}

	var req updateRaceRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Date != nil {
		race.Date = *req.Date
	}
	if req.Type != nil {
		race.Type = *req.Type
	}
	if req.Place != nil {
		race.Place = *req.Place
	}
	if req.Distances != nil {
		race.Distances = *req.Distances
	}
	if req.Name != nil {
		race.Name = *req.Name
	}

	err = h.raceUC.Update(c.Request().Context(), race)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update race"})
	}

	return c.JSON(http.StatusOK, raceToResponse(race))
}

// Delete removes a race by ID.
func (h *RaceHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	err = h.raceUC.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete race"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListRegistrations returns all registrations for a race.
func (h *RaceHandler) ListRegistrations(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	regs, err := h.raceUC.ListRegistrations(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list registrations"})
	}

	resp := make([]raceRegistrationResponse, len(regs))
	for i := range regs {
		resp[i] = raceRegistrationToResponse(&regs[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// RegisterMember registers a member for a race.
func (h *RaceHandler) RegisterMember(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	var req registerMemberRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.MemberID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "member_id is required"})
	}

	if req.Distance == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "distance is required"})
	}

	err = h.raceUC.RegisterMember(c.Request().Context(), id, req.MemberID, req.Distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to register member"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "member registered"})
}

// UnregisterMember removes a member registration from a race.
func (h *RaceHandler) UnregisterMember(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidRaceID})
	}

	var req unregisterMemberRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.MemberID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "member_id is required"})
	}

	err = h.raceUC.UnregisterMember(c.Request().Context(), id, req.MemberID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to unregister member"})
	}

	return c.NoContent(http.StatusNoContent)
}
