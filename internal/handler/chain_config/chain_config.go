package chain_config

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChainConfigHandler struct {
	log *zap.Logger
}

func NewChainConfigHandler(log *zap.Logger) *ChainConfigHandler {
	return &ChainConfigHandler{
		log: log,
	}
}

// ChainConfig represents a chain configuration
type ChainConfig struct {
	ID               string   `json:"id"`
	ChainID          string   `json:"chain_id"`
	Name             string   `json:"name"`
	Networks         []string `json:"networks"`
	CryptoCurrencies []string `json:"crypto_currencies"`
	Processor        string   `json:"processor"`
	IsTestnet        bool     `json:"is_testnet"`
	Status           string   `json:"status"`
	CreatedAt        string   `json:"created_at"`
}

// GetChainConfigs returns all chain configurations
func (h *ChainConfigHandler) GetChainConfigs(c *gin.Context) {
	// Mock data for now - replace with actual database queries
	chainConfigs := []ChainConfig{
		{
			ID:               "1",
			ChainID:          "1",
			Name:             "Ethereum",
			Networks:         []string{"mainnet", "goerli", "sepolia"},
			CryptoCurrencies: []string{"ETH", "USDC", "USDT"},
			Processor:        "internal",
			IsTestnet:        false,
			Status:           "active",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "2",
			ChainID:          "56",
			Name:             "Binance Smart Chain",
			Networks:         []string{"mainnet", "testnet"},
			CryptoCurrencies: []string{"BNB", "USDC", "USDT"},
			Processor:        "internal",
			IsTestnet:        false,
			Status:           "active",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "3",
			ChainID:          "137",
			Name:             "Polygon",
			Networks:         []string{"mainnet", "mumbai"},
			CryptoCurrencies: []string{"MATIC", "USDC", "USDT"},
			Processor:        "internal",
			IsTestnet:        false,
			Status:           "active",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"chain_configs": chainConfigs,
			"total":         len(chainConfigs),
		},
		"message": "Chain configs fetched successfully",
	})
}
