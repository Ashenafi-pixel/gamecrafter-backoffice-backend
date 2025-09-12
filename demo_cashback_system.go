package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Simplified demonstration of the cashback system
func main() {
	fmt.Println("🎰 TucanBIT World-Class Cashback System Demo")
	fmt.Println("=============================================")

	// Demonstrate cashback calculation
	fmt.Println("\n💰 Cashback Calculation Example:")
	fmt.Println("User: Bronze tier (0.5% cashback)")
	fmt.Println("Bet: $100 on Plinko (2% house edge)")

	betAmount := decimal.NewFromFloat(100.0)
	houseEdge := decimal.NewFromFloat(0.02)     // 2%
	cashbackRate := decimal.NewFromFloat(0.005) // 0.5%

	expectedGGR := betAmount.Mul(houseEdge)
	earnedCashback := expectedGGR.Mul(cashbackRate)

	fmt.Printf("Expected GGR: $%s\n", expectedGGR.StringFixed(2))
	fmt.Printf("Cashback Earned: $%s\n", earnedCashback.StringFixed(2))

	// Demonstrate tier progression
	fmt.Println("\n🏆 Tier Progression Example:")
	tiers := []struct {
		name       string
		level      int
		minGGR     decimal.Decimal
		rate       decimal.Decimal
		dailyLimit decimal.Decimal
	}{
		{"Bronze", 1, decimal.Zero, decimal.NewFromFloat(0.5), decimal.NewFromFloat(50)},
		{"Silver", 2, decimal.NewFromFloat(1000), decimal.NewFromFloat(1.0), decimal.NewFromFloat(100)},
		{"Gold", 3, decimal.NewFromFloat(5000), decimal.NewFromFloat(1.5), decimal.NewFromFloat(250)},
		{"Platinum", 4, decimal.NewFromFloat(15000), decimal.NewFromFloat(2.0), decimal.NewFromFloat(500)},
		{"Diamond", 5, decimal.NewFromFloat(50000), decimal.NewFromFloat(2.5), decimal.NewFromFloat(1000)},
	}

	for _, tier := range tiers {
		fmt.Printf("  %s (Level %d): %.1f%% cashback, $%.0f daily limit\n",
			tier.name, tier.level, tier.rate, tier.dailyLimit)
	}

	// Demonstrate game house edges
	fmt.Println("\n🎮 Game House Edges:")
	games := []struct {
		name      string
		houseEdge decimal.Decimal
		minBet    decimal.Decimal
	}{
		{"Plinko", decimal.NewFromFloat(0.02), decimal.NewFromFloat(0.1)},
		{"Crash", decimal.NewFromFloat(0.01), decimal.NewFromFloat(0.1)},
		{"Dice", decimal.NewFromFloat(0.01), decimal.NewFromFloat(0.1)},
		{"Blackjack", decimal.NewFromFloat(0.0048), decimal.NewFromFloat(1.0)},
		{"Roulette", decimal.NewFromFloat(0.027), decimal.NewFromFloat(1.0)},
		{"Slots", decimal.NewFromFloat(0.03), decimal.NewFromFloat(0.1)},
	}

	for _, game := range games {
		fmt.Printf("  %s: %.2f%% house edge, $%.1f min bet\n",
			game.name, game.houseEdge.Mul(decimal.NewFromInt(100)), game.minBet)
	}

	// Demonstrate real-time processing flow
	fmt.Println("\n🚀 Real-time Processing Flow:")
	fmt.Println("1. User places bet → Kafka event published")
	fmt.Println("2. Cashback consumer processes event")
	fmt.Println("3. GGR calculated: bet_amount × house_edge")
	fmt.Println("4. Cashback earned: GGR × cashback_rate")
	fmt.Println("5. User level updated automatically")
	fmt.Println("6. Cashback available for claiming")

	// Demonstrate API endpoints
	fmt.Println("\n🔧 API Endpoints:")
	fmt.Println("GET  /user/cashback              - Get user cashback summary")
	fmt.Println("POST /user/cashback/claim        - Claim available cashback")
	fmt.Println("GET  /user/cashback/earnings     - Get earnings history")
	fmt.Println("GET  /user/cashback/claims       - Get claims history")
	fmt.Println("GET  /cashback/tiers             - Get all tiers")
	fmt.Println("GET  /admin/cashback/stats       - Get admin statistics")

	// Demonstrate benefits
	fmt.Println("\n🎯 Key Benefits:")
	fmt.Println("✅ Multi-tier cashback system (Bronze to Diamond)")
	fmt.Println("✅ Real-time GGR calculation and cashback earning")
	fmt.Println("✅ Automatic level progression based on GGR")
	fmt.Println("✅ Daily/Weekly/Monthly cashback limits")
	fmt.Println("✅ Special promotions and bonus multipliers")
	fmt.Println("✅ Expiring cashback earnings (30-day expiry)")
	fmt.Println("✅ Comprehensive admin statistics")
	fmt.Println("✅ Kafka integration for real-time processing")

	// Demonstrate database schema
	fmt.Println("\n📊 Database Schema:")
	fmt.Println("• user_levels - User level and progress tracking")
	fmt.Println("• cashback_tiers - Configurable tier definitions")
	fmt.Println("• cashback_earnings - Individual earning records")
	fmt.Println("• cashback_claims - Claim history and status")
	fmt.Println("• game_house_edges - Game-specific house edges")
	fmt.Println("• cashback_promotions - Special promotions")

	// Demonstrate level progression example
	fmt.Println("\n📈 Level Progression Example:")
	fmt.Println("Bronze → Silver: $1,000 GGR required")
	fmt.Println("Silver → Gold: $5,000 GGR required")
	fmt.Println("Gold → Platinum: $15,000 GGR required")
	fmt.Println("Platinum → Diamond: $50,000 GGR required")

	// Demonstrate special features
	fmt.Println("\n🎯 Special Features:")
	fmt.Println("✅ Automatic level progression")
	fmt.Println("✅ Expiring cashback (30-day expiry)")
	fmt.Println("✅ Daily/Weekly/Monthly limits")
	fmt.Println("✅ Promotion boosts")
	fmt.Println("✅ Comprehensive admin dashboard")
	fmt.Println("✅ Real-time statistics")

	fmt.Println("\n🎉 This cashback system provides:")
	fmt.Println("• World-class user experience")
	fmt.Println("• Fair and transparent rewards")
	fmt.Println("• Real-time processing")
	fmt.Println("• Comprehensive admin tools")
	fmt.Println("• Scalable architecture")
	fmt.Println("• High performance")

	fmt.Println("\n✨ Ready for production deployment!")
	fmt.Println("🚀 The system is designed to compete with the best online casinos in the world!")

	// Show current time for demo
	fmt.Printf("\n⏰ Demo completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
