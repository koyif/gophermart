package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/gophermart/internal/handler/middleware"
	"github.com/koyif/gophermart/internal/handler/order"
	"github.com/koyif/gophermart/internal/handler/user"
	"github.com/koyif/gophermart/internal/postgres"
	"github.com/koyif/gophermart/internal/service"
)

func (app App) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.WithGzip)
	r.Use(middleware.WithAuth(app.Config))

	p := postgres.New(app.DB)
	userService := service.NewUserService(p, app.Config)
	userHandler := userhandler.New(userService)

	orderService := service.NewOrderService(p)
	orderHandler := orderhandler.New(orderService)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)

		r.Post("/orders", orderHandler.CreateOrder)
		r.Get("/orders", orderHandler.ListOrders)
		//r.Get("balance", balanceHandler.GetBalance)
		//r.Post("balance/withdraw", balanceHandler.Withdraw)
		//r.Get("withdrawals", balanceHandler.ListWithdrawals)
	})

	return r
}
