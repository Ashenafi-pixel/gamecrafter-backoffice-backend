package falcon_liquidity

import (
	"net/http"

	"github.com/gin-gonic/gin"
	falconHandler "github.com/tucanbit/internal/handler/falcon_liquidity"
	"github.com/tucanbit/internal/storage/falcon_liquidity"
	"go.uber.org/zap"
)

// GetFalconLiquidityRoutes returns the routes for Falcon Liquidity API
func GetFalconLiquidityRoutes(falconStorage falcon_liquidity.FalconMessageStorage, logger *zap.Logger) []Route {
	falconHandler := falconHandler.NewFalconLiquidityHandler(falconStorage, logger)

	return []Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/falcon-liquidity/data",
			Handler: falconHandler.GetAllFalconLiquidityData,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/falcon-liquidity/transaction/:transaction_id",
			Handler: falconHandler.GetFalconLiquidityByTransactionID,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/falcon-liquidity/user/:user_id",
			Handler: falconHandler.GetFalconLiquidityByUserID,
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/falcon-liquidity/summary",
			Handler: falconHandler.GetFalconLiquiditySummary,
		},
	}
}

// Route represents a single route configuration
type Route struct {
	Method  string
	Path    string
	Handler func(c *gin.Context)
}
