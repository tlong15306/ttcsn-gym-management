package v1routes

import (
	"github.com/gin-gonic/gin"
	v1handler "github.com/tlong15306/ttcsn-gym-management/server/internal/handler/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
)

type Dependencies struct {
	DiscoveryHandler  *v1handler.DiscoveryHandler
	AttendanceHandler *v1handler.AttendanceHandler
	BookingHandler    *v1handler.BookingHandler
	ReviewHandler     *v1handler.ReviewHandler
	ReportHandler     *v1handler.ReportHandler
	RoomCoverHandler  *v1handler.RoomCoverHandler
	Readiness         v1handler.ReadinessChecker
	IdentityHandler   *v1handler.IdentityHandler
	ManagementHandler *v1handler.ManagementHandler
	CalendarHandler   *v1handler.CalendarHandler
	Authenticator     *middleware.Authenticator
	CORSOrigins       []string
}

func New(deps Dependencies) (*gin.Engine, error) {
	r := gin.New()

	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.AccessLog(), middleware.CORS(deps.CORSOrigins))

	healthHandler := v1handler.NewHealthHandler(deps.Readiness)
	r.GET("/health/live", healthHandler.Live)
	r.GET("/health/ready", healthHandler.Ready)
	if h := deps.DiscoveryHandler; h != nil {
		public := r.Group("/api/v1/public")
		public.GET("/rooms", h.Rooms)
		public.GET("/rooms/:id", h.Room)
		public.GET("/rooms/:id/cover-image", h.Cover)
		public.GET("/rooms/:id/sessions", h.Sessions)
		public.GET("/trainers", h.Trainers)
	}

	if deps.IdentityHandler != nil && deps.Authenticator != nil {
		api := r.Group("/api/v1")
		api.POST("/auth/login", deps.IdentityHandler.Login)

		authenticated := api.Group("")
		authenticated.Use(deps.Authenticator.Require())
		authenticated.POST("/auth/logout", deps.IdentityHandler.Logout)
		authenticated.GET("/me", deps.IdentityHandler.Me)
		authenticated.PATCH("/me", deps.IdentityHandler.UpdateMe)
		if h := deps.RoomCoverHandler; h != nil {
			authenticated.GET("/rooms/:id/cover-image", h.Read)
			authenticated.POST("/rooms/:id/cover-image", h.Upload)
			authenticated.DELETE("/rooms/:id/cover-image", h.Delete)
		}
		if h := deps.AttendanceHandler; h != nil {
			authenticated.GET("/bookings/:id/attendance", h.Get)
			authenticated.POST("/bookings/:id/attendance", h.Record)
		}
		if h := deps.BookingHandler; h != nil {
			authenticated.GET("/booking-options/sessions", h.Sessions)
			authenticated.GET("/booking-options/pts", h.PTs)
			authenticated.POST("/bookings", h.Create)
			authenticated.POST("/bookings/batch", h.CreateBatch)
			authenticated.POST("/members/:id/bookings", h.Create)
			authenticated.POST("/members/:id/bookings/batch", h.CreateBatch)
			authenticated.GET("/bookings", h.List)
			authenticated.GET("/bookings/:id", h.Get)
			authenticated.POST("/bookings/:id/pt-response", h.Act)
			authenticated.POST("/bookings/:id/cancel", h.Act)
		}
		if h := deps.ReviewHandler; h != nil {
			authenticated.POST("/bookings/:id/review", h.Create)
			authenticated.GET("/bookings/:id/review", h.Get)
			authenticated.GET("/pts/:id/reviews", h.ListPT)
		}
		if h := deps.ReportHandler; h != nil {
			authenticated.GET("/reports/operations", h.Operations)
			authenticated.GET("/reports/pts/:id", h.Trainer)
		}

		accounts := authenticated.Group("/users")
		accounts.Use(middleware.RequireAdmin())
		accounts.GET("", deps.IdentityHandler.ListUsers)
		accounts.POST("", deps.IdentityHandler.CreateUser)
		accounts.GET("/:id", deps.IdentityHandler.GetAccount)
		accounts.PATCH("/:id", deps.IdentityHandler.UpdateAccount)
		accounts.PATCH("/:id/status", deps.IdentityHandler.UpdateStatus)
		accounts.PATCH("/:id/role", deps.IdentityHandler.UpdateRole)
		accounts.POST("/:id/reset-password", deps.IdentityHandler.ResetPassword)
		if h := deps.CalendarHandler; h != nil {
			authenticated.GET("/rooms/:id/weekly-schedule", h.Weekly)
			authenticated.GET("/rooms/:id/calendar", h.Calendar)
			authenticated.GET("/rooms/:id/calendar/history", h.History)
			authenticated.POST("/rooms/:id/calendar/preview", h.Preview)
			authenticated.PUT("/rooms/:id/weekly-schedule", h.Publish)
			authenticated.POST("/rooms/:id/calendar/exceptions", h.Exception)
			authenticated.PATCH("/rooms/:id/calendar/exceptions/:exceptionId", h.Exception)
			authenticated.POST("/rooms/:id/calendar/exceptions/:exceptionId/withdraw", h.Withdraw)
			authenticated.POST("/rooms/:id/calendar/materialize", h.Materialize)
		}
		if h := deps.ManagementHandler; h != nil {
			authenticated.GET("/members", h.Members)
			authenticated.GET("/members/eligible-accounts", h.EligibleAccounts)
			authenticated.POST("/members", h.SaveMember)
			authenticated.GET("/members/:id", h.Member)
			authenticated.PATCH("/members/:id", h.SaveMember)
			authenticated.PATCH("/members/:id/membership", h.Membership)
			authenticated.GET("/rooms", h.Rooms)
			authenticated.POST("/rooms", h.SaveRoom)
			authenticated.GET("/rooms/:id", h.Room)
			authenticated.GET("/rooms/:id/impact", h.RoomImpact)
			authenticated.PATCH("/rooms/:id", h.SaveRoom)
			authenticated.GET("/time-slots", h.Slots)
			authenticated.GET("/time-slots/:id", h.Slot)
			authenticated.PATCH("/time-slots/:id", h.UpdateSlot)
			authenticated.POST("/time-slots", h.CreateSlot)
			authenticated.PATCH("/time-slots/:id/status", h.SlotStatus)
			authenticated.GET("/room-sessions", h.Sessions)
			authenticated.POST("/room-sessions", h.SaveSession)
			authenticated.GET("/room-sessions/:id", h.Session)
			authenticated.PATCH("/room-sessions/:id", h.SaveSession)
			authenticated.GET("/pts", h.Trainers)
			authenticated.GET("/pts/me", h.OwnTrainer)
			authenticated.PUT("/pts/me/profile", h.SaveTrainer)
			authenticated.POST("/pts/me/submit", h.SubmitTrainer)
			authenticated.POST("/pts/me/certificates", h.UploadCertificate)
			authenticated.GET("/pts/:id", h.Trainer)
			authenticated.POST("/pts/:id/decision", h.TrainerDecision)
			authenticated.GET("/pts/:id/certificate", h.DownloadCertificate)
			for _, prefix := range []string{"/pts/me", "/pts/:id"} {
				authenticated.GET(prefix+"/availability", h.Hours)
				authenticated.PUT(prefix+"/availability", h.SaveHours)
				authenticated.GET(prefix+"/unavailability", h.Absences)
				authenticated.POST(prefix+"/unavailability", h.SaveAbsence)
				authenticated.DELETE(prefix+"/unavailability/:absenceId", h.DeleteAbsence)
			}
		}
	}

	return r, nil
}
