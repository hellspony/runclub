package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"runclub/internal/domain/entity"
	"runclub/internal/usecase"
)

// TemplateHandler handles template endpoints.
type TemplateHandler struct {
	templateUC usecase.TemplateUseCase
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(templateUC usecase.TemplateUseCase) *TemplateHandler {
	return &TemplateHandler{templateUC: templateUC}
}

type createTemplateRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type updateTemplateRequest struct {
	Type    *string `json:"type"`
	Name    *string `json:"name"`
	Content *string `json:"content"`
}

type templateResponse struct {
	ID      int64  `json:"id"`
	ClubID  int64  `json:"club_id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

func templateToResponse(t *entity.Template) templateResponse {
	return templateResponse{
		ID:      t.ID,
		ClubID:  t.ClubID,
		Type:    string(t.Type),
		Name:    t.Name,
		Content: t.Content,
	}
}

// List returns all templates for a club.
func (h *TemplateHandler) List(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	templates, err := h.templateUC.ListByClub(c.Request().Context(), clubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to list templates"})
	}

	resp := make([]templateResponse, len(templates))
	for i := range templates {
		resp[i] = templateToResponse(&templates[i])
	}

	return c.JSON(http.StatusOK, resp)
}

// Create creates a new template.
func (h *TemplateHandler) Create(c echo.Context) error {
	clubID, err := strconv.ParseInt(c.Param("clubId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidClubID})
	}

	var req createTemplateRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Type == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "type is required"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errNameRequired})
	}

	tmpl := &entity.Template{
		ClubID:  clubID,
		Type:    entity.TemplateType(req.Type),
		Name:    req.Name,
		Content: req.Content,
	}

	id, err := h.templateUC.Create(c.Request().Context(), tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to create template"})
	}

	tmpl.ID = id
	return c.JSON(http.StatusCreated, templateToResponse(tmpl))
}

// Update updates an existing template.
func (h *TemplateHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid template id"})
	}

	tmpl, err := h.templateUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{jsonKeyError: "template not found"})
	}

	var req updateTemplateRequest
	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: errInvalidReqBody})
	}

	if req.Type != nil {
		tmpl.Type = entity.TemplateType(*req.Type)
	}
	if req.Name != nil {
		tmpl.Name = *req.Name
	}
	if req.Content != nil {
		tmpl.Content = *req.Content
	}

	err = h.templateUC.Update(c.Request().Context(), tmpl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to update template"})
	}

	return c.JSON(http.StatusOK, templateToResponse(tmpl))
}

// Delete removes a template by ID.
func (h *TemplateHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{jsonKeyError: "invalid template id"})
	}

	err = h.templateUC.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{jsonKeyError: "failed to delete template"})
	}

	return c.NoContent(http.StatusNoContent)
}
