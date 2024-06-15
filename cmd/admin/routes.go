package admin

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/ukane-philemon/labtracka-api/internal/response"
)

func (s *Server) registerRoutes() http.Handler {
	mux := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH"},
		Debug:          true,
	})
	mux.Use(cors.Handler)
	mux.Use(httprate.LimitByRealIP(20, time.Minute))

	mux.Get("/", s.root)

	/*** Authentication ***/
	mux.Post("/otp", s.handleSendOTP)
	mux.Post("/otp-validation", s.handleOTPValidation)
	mux.Post("/login", s.handleLogin)
	mux.Post("/reset-password", s.handleResetPassword)
	// mux.Post("/validate-public-key", s.handlePublicKeyValidation)

	/**** Account ****/
	// mux.Post("/create-account", nil)

	mux.Group(func(withAuth chi.Router) {
		// Authentication is required for these routes.
		withAuth.Use(s.accessTokenValidator)

		/**** Authentication ****/
		withAuth.Get("/refresh-auth-token", s.handleRefreshAuthToken)
		// withAuth.Post("/change-password", nil)

		/**** Dashboard & Profile ****/
		// withAuth.Get("/dashboard", nil)
		// withAuth.Get("/profile", nil)
		// withAuth.Patch("/profile", nil)
		// withAuth.Post("/profile-image", nil) // TODO:
		// withAuth.Get("/profile-image", nil)  // TODO:

		/**** ADMIN ****/
		// withAuth.Get("/admins", nil)
		// withAuth.Post("/admin", nil)
		// withAuth.Patch("/admin", nil)
		// withAuth.Delete("/admin", nil)

		/**** Payments ****/
		// withAuth.Get("/payments", nil)
		// withAuth.Post("/request-payment", nil)
		// withAuth.Post("/payment-approval", nil)

		/**** Labs ****/
		// withAuth.Get("/labs", nil)
		// withAuth.Post("/lab", nil)
		// withAuth.Patch("/lab", nil)

		/**** Tests ****/
		// withAuth.Get("/tests", nil)
		// withAuth.Post("/test", nil)
		// withAuth.Patch("/test", nil)

		/**** Orders ****/
		// withAuth.Get("/orders", nil)

		/**** Results ****/
		// withAuth.Get("/results", nil)
		// withAuth.Post("/upload-result", nil)

		/**** Notifications ****/
		withAuth.Get("/notifications", s.handleGetNotifications)
		withAuth.Post("/mark-notifications-as-read", s.handleMarkNotificationAsRead)

		/**** Miscellaneous ****/
		// withAuth.Get("/faqs", nil)
		// withAuth.Post("/faqs", nil)
	})

	return mux
}

func (s *Server) root(res http.ResponseWriter, _ *http.Request) {
	response.JSON(res, http.StatusOK, "Welcome to the Labtracka Admin API, visit https://labtracka.com to get started!")
}
