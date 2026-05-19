package http

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"runclub/internal/config"
	"runclub/internal/usecase"
)

// Server wires together the echo engine, handlers, and routes.
type Server struct {
	echo       *echo.Echo
	clubH      *ClubHandler
	memberH    *MemberHandler
	locationH  *LocationHandler
	raceH      *RaceHandler
	templateH  *TemplateHandler
	trainingH  *TrainingHandler
	jointRunH  *JointRunHandler
	authH      *AuthHandler
	adminUserH *AdminUserHandler
	authUC     usecase.AuthUseCase
	staticDir  string
}

// NewServer creates a new Server with the given configuration and use cases.
func NewServer(
	cfg config.Config,
	clubUC usecase.ClubUseCase,
	memberUC usecase.MemberUseCase,
	locationUC usecase.LocationUseCase,
	raceUC usecase.RaceUseCase,
	templateUC usecase.TemplateUseCase,
	trainingUC usecase.TrainingUseCase,
	jointRunUC usecase.JointRunUseCase,
	authUC usecase.AuthUseCase,
) *Server {
	e := echo.New()
	e.HideBanner = true

	s := &Server{
		echo:       e,
		clubH:      NewClubHandler(clubUC, authUC),
		memberH:    NewMemberHandler(memberUC),
		locationH:  NewLocationHandler(locationUC),
		raceH:      NewRaceHandler(raceUC),
		templateH:  NewTemplateHandler(templateUC),
		trainingH:  NewTrainingHandler(trainingUC),
		jointRunH:  NewJointRunHandler(jointRunUC),
		authH:      NewAuthHandler(authUC),
		adminUserH: NewAdminUserHandler(authUC, clubUC),
		authUC:     authUC,
		staticDir:  cfg.StaticDir,
	}

	s.registerMiddleware()
	s.registerRoutes()

	return s
}

// registerMiddleware adds global middleware to the echo instance.
func (s *Server) registerMiddleware() {
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			return nil
		},
	}))
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))
}

// registerRoutes sets up all API and static routes.
func (s *Server) registerRoutes() {
	api := s.echo.Group("/api/v1")

	// Public auth routes.
	api.POST("/auth/login", s.authH.Login)

	// Protected auth routes.
	authProtected := api.Group("", AuthMiddleware(s.authUC))
	authProtected.GET("/auth/me", s.authH.Me)

	// All resource routes are protected.
	protected := api.Group("", AuthMiddleware(s.authUC))

	// Clubs — list/get available to all authenticated users (filtered by role in handler),
	// create is superadmin only.
	protected.GET("/clubs", s.clubH.List)
	protected.GET("/clubs/:id", s.clubH.Get)
	protected.PUT("/clubs/:id", s.clubH.Update)
	protected.DELETE("/clubs/:id", s.clubH.Delete)

	// Club creation — superadmin only.
	superAdminOnly := protected.Group("", SuperAdminMiddleware())
	superAdminOnly.POST("/clubs", s.clubH.Create)

	// Admin user management — superadmin only.
	superAdminOnly.GET("/admin-users", s.adminUserH.List)
	superAdminOnly.POST("/admin-users", s.adminUserH.Create)
	superAdminOnly.DELETE("/admin-users/:id", s.adminUserH.Delete)
	superAdminOnly.GET("/admin-users/:id/clubs", s.adminUserH.ListClubs)
	superAdminOnly.POST("/admin-users/:id/clubs/:clubId", s.adminUserH.AssignClub)
	superAdminOnly.DELETE("/admin-users/:id/clubs/:clubId", s.adminUserH.UnassignClub)

	// Club members.
	protected.GET("/clubs/:clubId/members", s.memberH.List)
	protected.POST("/clubs/:clubId/members", s.memberH.Create)
	protected.PUT("/clubs/:clubId/members/:memberId/role", s.memberH.UpdateRole)

	// Members.
	protected.GET("/members/:id", s.memberH.Get)
	protected.PUT("/members/:id", s.memberH.Update)
	protected.DELETE("/members/:id", s.memberH.Delete)

	// Club locations.
	protected.GET("/clubs/:clubId/locations", s.locationH.List)
	protected.POST("/clubs/:clubId/locations", s.locationH.Create)

	// Locations.
	protected.GET("/locations/:id", s.locationH.Get)
	protected.PUT("/locations/:id", s.locationH.Update)
	protected.DELETE("/locations/:id", s.locationH.Delete)

	// Club races.
	protected.GET("/clubs/:clubId/races", s.raceH.List)
	protected.POST("/clubs/:clubId/races", s.raceH.Create)

	// Races.
	protected.GET("/races/:id", s.raceH.Get)
	protected.PUT("/races/:id", s.raceH.Update)
	protected.DELETE("/races/:id", s.raceH.Delete)
	protected.GET("/races/:id/registrations", s.raceH.ListRegistrations)
	protected.POST("/races/:id/register", s.raceH.RegisterMember)
	protected.DELETE("/races/:id/register", s.raceH.UnregisterMember)

	// Club templates.
	protected.GET("/clubs/:clubId/templates", s.templateH.List)
	protected.POST("/clubs/:clubId/templates", s.templateH.Create)

	// Templates.
	protected.PUT("/templates/:id", s.templateH.Update)
	protected.DELETE("/templates/:id", s.templateH.Delete)

	// Club trainings.
	protected.GET("/clubs/:clubId/trainings", s.trainingH.List)
	protected.POST("/clubs/:clubId/trainings", s.trainingH.Create)

	// Trainings.
	protected.GET("/trainings/:id", s.trainingH.Get)
	protected.PUT("/trainings/:id", s.trainingH.Update)
	protected.DELETE("/trainings/:id", s.trainingH.Delete)
	protected.GET("/trainings/:id/participants", s.trainingH.ListParticipants)

	// Club joint runs.
	protected.GET("/clubs/:clubId/joint-runs", s.jointRunH.List)
	protected.POST("/clubs/:clubId/joint-runs", s.jointRunH.Create)

	// Joint runs.
	protected.GET("/joint-runs/:id", s.jointRunH.Get)
	protected.DELETE("/joint-runs/:id", s.jointRunH.Delete)
	protected.GET("/joint-runs/:id/participants", s.jointRunH.ListParticipants)

	// Serve static files for non-API routes (SPA fallback).
	if s.staticDir != "" {
		s.echo.Static("/", s.staticDir)
	}
}

// Start starts the HTTP server on the given address.
func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
