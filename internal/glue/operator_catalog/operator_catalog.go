package operator_catalog

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/storage"
	gameStorage "github.com/tucanbit/internal/storage/game"
	"go.uber.org/zap"
)

// Init registers operator-facing catalog routes for operators.
// These endpoints are authenticated via HMAC headers (X-Operator-Id, X-Operator-Signature)
// and are not part of the backoffice Authz module.
func Init(
	grp *gin.RouterGroup,
	log *zap.Logger,
	operatorStorage storage.Operator,
	gameStore gameStorage.GameStorage,
) {
	log.Info("Initializing operator catalog routes")

	catalogGroup := grp.Group("/api/operator")

	// GET /api/operator/games/catalog
	catalogGroup.GET(
		"/games/catalog",
		middleware.OperatorSignatureMiddleware(func(ctx context.Context, operatorID int32) (string, error) {
			return operatorStorage.GetActiveSigningKeyByOperatorID(ctx, operatorID)
		}),
		middleware.OperatorRateLimitMiddleware(0, 0),
		func(c *gin.Context) {
			handleGetCatalog(c, log, operatorStorage, gameStore)
		},
	)

	// DEV/QA ONLY: unsafe variant without HMAC, just X-Operator-Id header.
	// This is to make it easy to test that "operator only sees assigned games"
	// without having to generate signatures. Do NOT expose this in production.
	catalogGroup.GET(
		"/games/catalog/unsafe",
		middleware.OperatorRateLimitMiddleware(0, 0),
		func(c *gin.Context) {
			operatorIDStr := c.GetHeader(middleware.HeaderOperatorID)
			if operatorIDStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Missing X-Operator-Id header",
				})
				return
			}
			id, err := strconv.ParseInt(strings.TrimSpace(operatorIDStr), 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid X-Operator-Id",
				})
				return
			}
			c.Set("operator_id", int32(id))
			handleGetCatalog(c, log, operatorStorage, gameStore)
		},
	)
}

func handleGetCatalog(
	c *gin.Context,
	log *zap.Logger,
	operatorStorage storage.Operator,
	gameStore gameStorage.GameStorage,
) {
	ctx := c.Request.Context()

	operatorIDVal, ok := c.Get("operator_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "operator context missing",
		})
		return
	}

	operatorID, ok := operatorIDVal.(int32)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid operator id type",
		})
		return
	}

	gameIDs, err := operatorStorage.GetOperatorGameIDs(ctx, operatorID)
	if err != nil {
		log.Error("failed to get operator game ids", zap.Int32("operator_id", operatorID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to get operator catalog",
		})
		return
	}

	if len(gameIDs) == 0 {
		c.JSON(http.StatusOK, []dto.GameResponse{})
		return
	}

	result := make([]dto.GameResponse, 0, len(gameIDs))
	for _, gid := range gameIDs {
		gameUUID, err := uuid.Parse(gid)
		if err != nil {
			log.Warn("skipping invalid game id for operator", zap.Int32("operator_id", operatorID), zap.String("game_id", gid), zap.Error(err))
			continue
		}

		g, err := gameStore.GetGameByID(ctx, gameUUID)
		if err != nil {
			log.Warn("failed to load game for operator catalog", zap.Int32("operator_id", operatorID), zap.String("game_id", gid), zap.Error(err))
			continue
		}
		if g == nil {
			continue
		}

		// Only expose enabled & ACTIVE games
		if !g.Enabled || g.Status != "ACTIVE" {
			continue
		}

		result = append(result, dto.GameResponse{
			ID:                 g.ID,
			Name:               g.Name,
			Status:             g.Status,
			Timestamp:          g.Timestamp,
			Photo:              g.Photo,
			Price:              g.Price,
			Enabled:            g.Enabled,
			GameID:             g.GameID,
			InternalName:       g.InternalName,
			IntegrationPartner: g.IntegrationPartner,
			Provider:           g.Provider,
			CreatedAt:          g.CreatedAt,
			UpdatedAt:          g.UpdatedAt,
		})
	}

	if result == nil {
		result = []dto.GameResponse{}
	}

	c.JSON(http.StatusOK, result)
}

