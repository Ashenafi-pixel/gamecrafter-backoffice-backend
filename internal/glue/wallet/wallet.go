package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/handler/wallet"
	"go.uber.org/zap"
)

func Init(group *gin.RouterGroup, log *zap.Logger) {
	walletHandler := wallet.NewWalletHandler(log)

	walletRoutes := []routing.Route{
		{
			Method:  "GET",
			Path:    "/api/admin/wallets/hot-wallet-data",
			Handler: walletHandler.GetHotWalletData,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "GET",
			Path:    "/api/admin/wallets/cold-storage-data",
			Handler: walletHandler.GetColdWalletData,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
		{
			Method:  "POST",
			Path:    "/api/admin/wallets/move-funds-to-hot",
			Handler: walletHandler.MoveFundsToHot,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
			},
		},
	}

	for _, route := range walletRoutes {
		group.Handle(route.Method, route.Path, append(route.Middleware, route.Handler)...)
	}
}
