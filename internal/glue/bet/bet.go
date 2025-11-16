package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tucanbit/internal/glue/routing"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/handler/middleware"
	"github.com/tucanbit/internal/module"
	"go.uber.org/zap"
)

func Init(
	group *gin.RouterGroup,
	log zap.Logger,
	bet handler.Bet,
	userModule module.User,
	authModule module.Authz,
	systemLogs module.SystemLogs,
) {

	balanceRoutes := []routing.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/game/round",
			Handler: bet.GetOpenRound,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/game/place-bet",
			Handler: bet.PlaceBet,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/game/cash-out",
			Handler: bet.CashOut,
			Middleware: []gin.HandlerFunc{
				middleware.Auth(),
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/game/history",
			Handler: bet.GetBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.IpFilter(userModule),
				middleware.Authz(authModule, "get game history", http.MethodGet),
				middleware.SystemLogs("Get Game History", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/game/cancel",
			Handler: bet.CancelBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
				middleware.SystemLogs("Cancel Bet", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/game/leaders",
			Handler: bet.GetLeaders,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/user/game/history",
			Handler: bet.GetMyBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.IpFilter(userModule),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/failed/rounds",
			Handler: bet.GetAllFailedRounds,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get failed rounds", http.MethodGet),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/failed/rounds",
			Handler: bet.ManualRefundFailedRound,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "manual refund failed rounds", http.MethodPost),
				middleware.SystemLogs("Manual Refund Failed Round", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/plinko/config",
			Handler: bet.GetPlinkoGameConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/plinko/drop",
			Handler: bet.PlacePlinkoBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/plinko/history",
			Handler: bet.GetUserPlinkoBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/plinko/stats",
			Handler: bet.GetPlinkoGameStats,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/leagues",
			Handler: bet.CreateLeague,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create league", http.MethodPost),
				middleware.SystemLogs("Create League", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/football/leagues",
			Handler: bet.GetLeagues,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get leagues", http.MethodGet),
				middleware.SystemLogs("Get Leagues", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/clubs",
			Handler: bet.CreateClub,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create clubs", http.MethodPost),
				middleware.SystemLogs("Create Club", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/football/clubs",
			Handler: bet.GetClubs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get clubs", http.MethodGet),
				middleware.SystemLogs("Get Clubs", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/football/configs",
			Handler: bet.UpdateFootballCardMultiplierValue,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update football match multiplier", http.MethodPut),
				middleware.SystemLogs("Update Football Match Multiplier", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/football/configs",
			Handler: bet.GetFootballCardMultiplier,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get football match multiplier", http.MethodGet),
				middleware.SystemLogs("Get Football Match Multiplier", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/rounds",
			Handler: bet.CreateFootballMatchRound,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create football match round", http.MethodPost),
				middleware.SystemLogs("Create Football Match Round", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/football/rounds",
			Handler: bet.GetFootballMatchRounds,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get football match round", http.MethodGet),
				middleware.SystemLogs("Get Football Match Round", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/matches",
			Handler: bet.CreateFootballMatch,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create football matches", http.MethodPost),
				middleware.SystemLogs("Create Football Match", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/rounds/mathces",
			Handler: bet.GetFootballRoundMatchs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get football match round matches", http.MethodPost),
				middleware.SystemLogs("Get Football Match Round Matches", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/football/rounds/current",
			Handler: bet.GetCurrentFootballRound,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
			},
		}, {
			Method:  http.MethodPatch,
			Path:    "/api/admin/football/rounds/matches",
			Handler: bet.CloseFootballMatch,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "close football matche", http.MethodPatch),
				middleware.SystemLogs("Close Football Match", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/football/rounds/prices",
			Handler: bet.UpdateFootballRoundPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "udpate football round price", http.MethodPost),
				middleware.SystemLogs("Update Football Round Price", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/football/rounds/prices",
			Handler: bet.GetFootballRoundPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/football/rounds/bets",
			Handler: bet.PleceBetOnFootballRound,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth()},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/football/rounds/bets",
			Handler: bet.GetUserFootballBets,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth()},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/streetkings/bets",
			Handler: bet.CreateStreetKingsGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth()},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/streetkings/bets/cashout",
			Handler: bet.CashOutStreetKingsBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth()},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/streetkings/bets",
			Handler: bet.GetStreetkingHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth()},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/cryptokings/configs",
			Handler: bet.SetCrytoKingsConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update cryptokings config", http.MethodPost),
				middleware.SystemLogs("Update CryptoKings Config", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/cryptokings",
			Handler: bet.PlaceCryptoKingsBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/cryptokings",
			Handler: bet.GetCryptoKingsBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/cryptokings/price",
			Handler: bet.GetCryptoKingsCurrentCryptoPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/quickhustles",
			Handler: bet.PlaceQuickHustleBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/quickhustles/selects",
			Handler: bet.UserSelectCard,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/quickhustles",
			Handler: bet.GetQuickHustleBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/rolldadices",
			Handler: bet.CreateRollDaDice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/rolldadices",
			Handler: bet.GetRollDaDiceHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/scratchcards/price",
			Handler: bet.GetScratchGamePrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/scratchcards",
			Handler: bet.PlaceScratchCardBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/scratchcards",
			Handler: bet.GetUserScratchCardBetHistories,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/spinningwheels/price",
			Handler: bet.GetSpinningWheelPrice,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/spinningwheels",
			Handler: bet.PlaceSpinningWheelBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/spinningwheels",
			Handler: bet.GetSpinningWheelUserBetHistory,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/games",
			Handler: bet.GetGames,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get games", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/games/summary",
			Handler: bet.GetGameSummary,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get game summary", http.MethodGet),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/transactions/summary",
			Handler: bet.GetTransactionSummary,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get transaction summary", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/games",
			Handler: bet.UpdateGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update games", http.MethodPut),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/games/disable",
			Handler: bet.DisableAllGames,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "disable games", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/games/available",
			Handler: bet.GetAvailableGames,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get available games", http.MethodGet),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/api/admin/games",
			Handler: bet.DeleteGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete games", http.MethodDelete),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/games",
			Handler: bet.AddGame,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "add games", http.MethodDelete),
			},
		},
		{
			Method:  http.MethodPut,
			Path:    "/api/admin/games/status",
			Handler: bet.UpdateGameStatus,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update game status", http.MethodPost),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/spinningwheels/mysteries",
			Handler: bet.CreateSpinningWheelMysteries,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create mysteries", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/spinningwheels/mysteries",
			Handler: bet.GetSpinningWheelMysteries,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get mysteries", http.MethodGet),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/spinningwheels/mysteries",
			Handler: bet.DeleteSpinningWheelMystery,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete mysteries", http.MethodDelete),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/spinningwheels/mysteries",
			Handler: bet.UpdateSpinningWheelMystery,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update mysteries", http.MethodPut),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/spinningwheels/config",
			Handler: bet.CreateSpinningWheelConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create spinning wheel config", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/spinningwheels/configs",
			Handler: bet.GetSpinningWheelConfigs,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/spinningwheels/config",
			Handler: bet.DeleteSpinningWheelConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "delete spinning wheel config", http.MethodDelete),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/spinningwheels/config",
			Handler: bet.UpdateSpinningWheelConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update spinning wheel config", http.MethodPut),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/bets/icons",
			Handler: bet.UpdateBetIcon,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update bet icon", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/scratchcard/configs",
			Handler: bet.GetScratchCardsConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get scratch card configs", http.MethodGet),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/scratchcard/configs",
			Handler: bet.UpdateScratchGameConfig,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update scratch card configs", http.MethodPut),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/levels",
			Handler: bet.CreateLevel,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create bet level", http.MethodPost),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/levels",
			Handler: bet.GetLevels,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "get bet levels", http.MethodGet),
			},
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/admin/levels/requirements",
			Handler: bet.CreateLevelRequirements,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "create level requirements", http.MethodPost),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/levels/requirements",
			Handler: bet.UpdateLevelRequirement,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "update level requirements", http.MethodPut),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/users/level",
			Handler: bet.GetUserLevel,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/test/fake/transaction",
			Handler: bet.AddFakeBalanceLog,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/admin/lootboxes",
			Handler: bet.CreateLootBox,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Create Loot Box", http.MethodPost),
				middleware.SystemLogs("Create Loot Box", &log, systemLogs),
			},
		}, {
			Method:  http.MethodPut,
			Path:    "/api/admin/lootboxes",
			Handler: bet.UpdateLootBox,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Update Loot Box", http.MethodPut),
				middleware.SystemLogs("Update Loot Box", &log, systemLogs),
			},
		}, {
			Method:  http.MethodDelete,
			Path:    "/api/admin/lootboxes/:id",
			Handler: bet.DeleteLootBox,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Delete Loot Box", http.MethodDelete),
				middleware.SystemLogs("Delete Loot Box", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/admin/lootboxes",
			Handler: bet.GetLootBox,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
				middleware.Authz(authModule, "Get Loot Box", http.MethodGet),
				middleware.SystemLogs("Get Loot Boxes", &log, systemLogs),
			},
		}, {
			Method:  http.MethodGet,
			Path:    "/api/bet/lootboxes",
			Handler: bet.PlaceLootBoxBet,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		}, {
			Method:  http.MethodPost,
			Path:    "/api/bet/lootboxes/select",
			Handler: bet.SelectLootBox,
			Middleware: []gin.HandlerFunc{
				middleware.RateLimiter(),
				middleware.Auth(),
			},
		},
	}
	routing.RegisterRoute(group, balanceRoutes, log)
}
