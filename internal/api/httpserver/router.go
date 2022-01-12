package httpserver

import (
	v1 "Smoozzy/TicketSystemController/internal/api/httpserver/v1"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewRouter(mux *chi.Mux,
	lg *zap.Logger,
	ticketController *v1.Ticket,
	cacheController *v1.CacheController) chi.Mux {
	mux.Use(middleware.Logger)
	mux.Route("/api/v1", func(router chi.Router) {
		ticketRouter(router, ticketController)
		cacheRouter(router, cacheController)
	})
	lg.Info("Router is started")
	return *mux
}

func ticketRouter(router chi.Router, ticketController *v1.Ticket) chi.Router {
	router.Post("/createticket", ticketController.CreateTicket)
	router.Post("/reopenticket", ticketController.ReopenTicket)
	router.Post("/changeticketstatus", ticketController.ChangeTicketStatus)
	router.Post("/checkticketstatus", ticketController.CheckTicketStatus)
	router.Post("/addnotetoticket", ticketController.AddNoteToTicket)
	router.Post("/closeticket", ticketController.CloseTicket)

	return router
}

func cacheRouter(router chi.Router, cacheController *v1.CacheController) chi.Router {
	router.Post("/cache/checkticketstatus", cacheController.CheckStatus)
	return router
}
