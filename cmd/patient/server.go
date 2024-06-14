package patient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/tomasen/realip"
	"github.com/ukane-philemon/labtracka-api/internal/jwt"
	"github.com/ukane-philemon/labtracka-api/internal/otp"
	"github.com/ukane-philemon/labtracka-api/internal/paystack"
	"github.com/ukane-philemon/labtracka-api/internal/response"
	"github.com/ukane-philemon/labtracka-api/internal/smtp"
)

const (
	defaultIdleTimeout    = time.Minute
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultShutdownPeriod = 30 * time.Second

	// apiVersion1 is the current and latest api version.
	apiVersion1 = 1
)

type Server struct {
	cfg *Config

	ctx     context.Context
	db      Database
	adminDB AdminDatabase
	baseURL string
	mailer  *smtp.Mailer
	logger  *slog.Logger
	wg      sync.WaitGroup

	optManager *otp.Manager
	jwtManager *jwt.Manager

	paystack paystack.Client
}

func NewServer(patientDB Database, adminDB AdminDatabase, cfg *Config) (*Server, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:        cfg,
		db:         patientDB,
		adminDB:    adminDB,
		optManager: otp.NewManager(cfg.DevMode),
		logger:     slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})),
	}

	s.jwtManager, err = jwt.NewJWTManager(false)
	if err != nil {
		return nil, err
	}

	if cfg.hasValidSMTPConfig() {
		s.mailer, err = smtp.NewMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
		if err != nil {
			return nil, err
		}
	} else {
		s.logger.Info("Proceeding with patient server without email sending enabled...")
	}

	return s, nil
}

// Run starts the server and is blocking. It'll only return when the server is
// stopped.
func (s *Server) Run() {
	var cancelMainCtx context.CancelFunc
	s.ctx, cancelMainCtx = context.WithCancel(context.Background())
	defer cancelMainCtx()

	handler := s.registerRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.cfg.ServerHost, s.cfg.ServerPort),
		Handler:      handler,
		ErrorLog:     slog.NewLogLogger(s.logger.Handler(), slog.LevelWarn),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	// Listen for shutdown signals.
	shutdownErrorChan := make(chan error)
	s.doInBackground("detect shutdown", func() error {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(s.ctx, defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
		cancelMainCtx()
		s.db.Shutdown()
		return nil
	})

	s.baseURL = srv.Addr
	s.logger.Info("Starting user API server", slog.Group("server", "addr", s.baseURL))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		trace := string(debug.Stack())
		s.logger.Error("Server exited with error:", err, "Trace", trace)
		return
	}

	err = <-shutdownErrorChan
	if err != nil {
		trace := string(debug.Stack())
		s.logger.Error("Server shutdown with error:", err, "Trace", trace)
		return
	}

	s.logger.Info("Stopped server", slog.Group("server", "addr", srv.Addr))

	s.wg.Wait()
}

func (s *Server) reqAuthID(req *http.Request) string {
	emailCtxValue := req.Context().Value(patientEmailCtx)
	if emailCtxValue != nil {
		return emailCtxValue.(string)
	}
	return ""
}

func (s *Server) sendSuccessResponse(res http.ResponseWriter, req *http.Request, message string) {
	err := response.JSON(res, http.StatusOK, map[string]string{"message": message})
	if err != nil {
		s.reportServerError(req, err)
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) sendSuccessResponseWithData(res http.ResponseWriter, req *http.Request, data any) {
	err := response.JSON(res, http.StatusOK, map[string]any{
		"data": data,
	})
	if err != nil {
		s.reportServerError(req, err)
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func clientIP(req *http.Request) string {
	return realip.FromRequest(req)
}
