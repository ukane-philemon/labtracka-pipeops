package patient

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
		AllowedMethods: []string{"POST", "GET", "DELETE"},
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

	/**** Account ****/
	mux.Post("/create-account", s.handleCreateAccount)

	mux.Group(func(withAuth chi.Router) {
		// Authentication is required for these routes.
		withAuth.Use(s.accessTokenValidator)

		/**** Authentication ****/
		withAuth.Get("/refresh-auth-token", s.handleRefreshAuthToken)
		withAuth.Post("/change-password", s.handleChangePassword)

		/**** Profile ****/
		withAuth.Get("/profile", s.handleGetProfile)
		withAuth.Post("/profile-image", s.handleProfileImageUpload)
		withAuth.Post("/sub-account", s.handleAddSubAccount)
		withAuth.Get("/sub-accounts", s.handleGetSubAccounts)
		withAuth.Delete("/sub-account", s.handleRemoveSubAccount)
		withAuth.Post("/add-address", s.handleAddAddress)
		withAuth.Get("/cards", s.handleGetCards)

		/**** Orders ****/
		withAuth.Post("/order", s.handleCreateOrder)
		withAuth.Get("/orders", s.handleGetOrders)

		/**** Notifications ****/
		withAuth.Get("/notifications", s.handleGetNotifications)
		withAuth.Patch("/mark-notifications-as-read", s.handleMarkNotificationAsRead)

		/**** Results ****/
		withAuth.Get("/results", s.handleGetResults)

		/**** Labs ****/
		withAuth.Get("/labs", s.handleGetLabs)
		withAuth.Get("/labs/{labID}/tests", s.handleGetLabTests)

		/**** Miscellaneous ****/
		withAuth.Get("/faqs", s.handleGetFaqs)
	})

	// Listen for Paystack webhook events.
	mux.Post("/paystack-webhook-events", s.handlePaystackWebhookEvent)

	return mux
}

func (s *Server) root(res http.ResponseWriter, _ *http.Request) {
	response.JSON(res, http.StatusOK, "Welcome to the Labtracka Patient API, visit https://labtracka.com to get started!")
}
