package brand_catalog

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

// Init registers operator-facing catalog routes for brands (operators).
// These endpoints are authenticated via HMAC headers (X-Brand-Id, X-Brand-Signature)
// and are not part of the backoffice Authz module.
func Init(
	grp *gin.RouterGroup,
	log *zap.Logger,
	brandStorage storage.Brand,
	gameStore gameStorage.GameStorage,
) {
	log.Info("Initializing brand catalog routes")

	catalogGroup := grp.Group("/api/brand")

	// GET /api/brand/games/catalog
	catalogGroup.GET(
		"/games/catalog",
		// First validate HMAC signature and set brand_id in context
		middleware.BrandSignatureMiddleware(func(ctx context.Context, brandID int32) (string, error) {
			return brandStorage.GetActiveSigningKeyByBrandID(ctx, brandID)
		}),
		// Then apply per-brand rate limiting
		middleware.BrandRateLimitMiddleware(0, 0),
		func(c *gin.Context) {
			handleGetCatalog(c, log, brandStorage, gameStore)
		},
	)

	// DEV/QA ONLY: unsafe variant without HMAC, just X-Brand-Id header.
	// This is to make it easy to test that "operator only sees assigned games"
	// without having to generate signatures. Do NOT expose this in production.
	catalogGroup.GET(
		"/games/catalog/unsafe",
		middleware.BrandRateLimitMiddleware(0, 0),
		func(c *gin.Context) {
			brandIDStr := c.GetHeader(middleware.HeaderBrandID)
			if brandIDStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Missing X-Brand-Id header",
				})
				return
			}
			// Reuse core handler by setting brand_id in context
			// Brand IDs in your DB are int32 sequence ids.
			id, err := strconv.ParseInt(strings.TrimSpace(brandIDStr), 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "Invalid X-Brand-Id",
				})
				return
			}
			c.Set("brand_id", int32(id))
			handleGetCatalog(c, log, brandStorage, gameStore)
		},
	)
}

func handleGetCatalog(
	c *gin.Context,
	log *zap.Logger,
	brandStorage storage.Brand,
	gameStore gameStorage.GameStorage,
) {
	ctx := c.Request.Context()

	brandIDVal, ok := c.Get("brand_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "brand context missing",
		})
		return
	}

	brandID, ok := brandIDVal.(int32)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid brand id type",
		})
		return
	}

	gameIDs, err := brandStorage.GetBrandGameIDs(ctx, brandID)
	if err != nil {
		log.Error("failed to get brand game ids", zap.Int32("brand_id", brandID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to get brand catalog",
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
			log.Warn("skipping invalid game id for brand", zap.Int32("brand_id", brandID), zap.String("game_id", gid), zap.Error(err))
			continue
		}

		g, err := gameStore.GetGameByID(ctx, gameUUID)
		if err != nil {
			log.Warn("failed to load game for brand catalog", zap.Int32("brand_id", brandID), zap.String("game_id", gid), zap.Error(err))
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

