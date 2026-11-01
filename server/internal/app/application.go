package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/db"
	v1handler "github.com/tlong15306/ttcsn-gym-management/server/internal/handler/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/job"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	v1routes "github.com/tlong15306/ttcsn-gym-management/server/internal/routes/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

// Application is the composition root: it wires repository, service, handler
// and routes without letting transport code construct its own dependencies.
type Application struct {
	config   config.Config
	pool     *pgxpool.Pool
	server   *http.Server
	calendar *repository.PostgresCalendarRepository
	booking  *repository.PostgresBookingRepository
}

func NewApplication(cfg config.Config) (*Application, error) {
	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	userRepository := repository.NewIdentityRepository(pool)
	tokens := auth.NewTokenService(cfg.JWTSecret, cfg.JWTTTL)
	identityService := v1service.NewService(userRepository, auth.NewPasswordService(cfg.BcryptCost), tokens)
	identityHandler := v1handler.NewIdentityHandler(identityService, int64(cfg.JWTTTL.Seconds()))
	calendarRepository := repository.NewCalendarRepository(pool)
	bookingRepository := repository.NewBookingRepository(pool)
	reviewHandler := v1handler.NewReviewHandler(v1service.NewReviewService(repository.NewReviewRepository(pool)))
	reportHandler := v1handler.NewReportHandler(v1service.NewReportService(repository.NewReportRepository(pool)))
	managementHandler := v1handler.NewManagementHandler(v1service.NewMemberService(repository.NewMemberRepository(pool), auth.NewPasswordService(cfg.BcryptCost)), v1service.NewFacilityService(repository.NewFacilityRepository(pool)), v1service.NewTrainerService(repository.NewTrainerRepository(pool), cfg.UploadDir))
	router, err := v1routes.New(v1routes.Dependencies{
		DiscoveryHandler:  v1handler.NewDiscoveryHandler(v1service.NewDiscoveryService(repository.NewDiscoveryRepository(pool))),
		AttendanceHandler: v1handler.NewAttendanceHandler(v1service.NewAttendanceService(repository.NewAttendanceRepository(pool))), BookingHandler: v1handler.NewBookingHandler(v1service.NewBookingService(bookingRepository)), ReviewHandler: reviewHandler, ReportHandler: reportHandler,
		RoomCoverHandler:  v1handler.NewRoomCoverHandler(v1service.NewRoomCoverService(repository.NewRoomCoverRepository(pool))),
		Readiness:         pool,
		IdentityHandler:   identityHandler,
		ManagementHandler: managementHandler,
		CalendarHandler:   v1handler.NewCalendarHandler(v1service.NewCalendarService(calendarRepository)),
		Authenticator:     middleware.NewAuthenticator(tokens, userRepository),
		CORSOrigins:       cfg.CORSOrigins,
	})
	if err != nil {
		pool.Close()
		return nil, err
	}

	return &Application{
		config:   cfg,
		pool:     pool,
		calendar: calendarRepository,
		booking:  bookingRepository,
		server:   &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 5 * time.Second},
	}, nil
}

func (a *Application) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer a.pool.Close()
	jobCtx, cancelJob := context.WithCancel(ctx)
	jobDone := make(chan struct{})
	go func() {
		defer close(jobDone)
		bookingDone := make(chan struct{})
		go func() { defer close(bookingDone); job.RunBooking(jobCtx, a.booking) }()
		job.RunCalendar(jobCtx, a.calendar)
		<-bookingDone
	}()
	defer func() { cancelJob(); <-jobDone }()

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting API", "environment", a.config.AppEnv, "address", a.config.HTTPAddr)
		serverErrors <- a.server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("start HTTP server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		err := <-serverErrors
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server stopped: %w", err)
		}
		return nil
	}
}
