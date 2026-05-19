package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// LocationHandler handles location endpoints.
type LocationHandler struct {
	locationUC usecase.LocationUseCase
}

// NewLocationHandler creates a new LocationHandler.
func NewLocationHandler(locationUC usecase.LocationUseCase) *LocationHandler {
	return &LocationHandler{locationUC: locationUC}
}

type createLocationRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	MapURL      string `json:"map_url"`
	Description string `json:"description"`
}

type updateLocationRequest struct {
	Name        *string `json:"name"`
	Address     *string `json:"address"`
	MapURL      *string `json:"map_url"`
	Description *string `json:"description"`
}

type locationResponse struct {
	ID          int64     `json:"id"`
	ClubID      int64     `json:"club_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	MapURL      string    `json:"map_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func locationToResponse(l *entity.Location) locationResponse {
	return locationResponse{
		ID:          l.ID,
		ClubID:      l.ClubID,
		Name:        l.Name,
		Address:     l.Address,
		MapURL:      l.MapURL,
		Description: l.Description,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

// List returns all locations for a club.
func (h *LocationHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	locations, err := h.locationUC.ListByClub(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list locations"})
	}

	resp := make([]locationResponse, len(locations))
	for i := range locations {
		resp[i] = locationToResponse(&locations[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new location.
func (h *LocationHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createLocationRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errNameRequired})
	}

	location := &entity.Location{
		ClubID:      clubID,
		Name:        req.Name,
		Address:     req.Address,
		MapURL:      req.MapURL,
		Description: req.Description,
	}

	id, err := h.locationUC.Create(c.Request().Context(), location)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create location"})
	}

	location.ID = id
	return c.JSON(http.StatusCreated, locationToResponse(location))
}

// Get returns a single location by ID.
func (h *LocationHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid location id"})
	}

	location, err := h.locationUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "location not found"})
	}

	return c.JSON(http.StatusOK, locationToResponse(location))
}

// Update updates an existing location.
func (h *LocationHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid location id"})
	}

	location, err := h.locationUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "location not found"})
	}

	var req updateLocationRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Name != nil {
		location.Name = *req.Name
	}
	if req.Address != nil {
		location.Address = *req.Address
	}
	if req.MapURL != nil {
		location.MapURL = *req.MapURL
	}
	if req.Description != nil {
		location.Description = *req.Description
	}

	err = h.locationUC.Update(c.Request().Context(), location)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update location"})
	}

	return c.JSON(http.StatusOK, locationToResponse(location))
}

// Delete removes a location by ID.
func (h *LocationHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid location id"})
	}

	err = h.locationUC.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete location"})
	}

	return c.NoContent(http.StatusNoContent)
}
