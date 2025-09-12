# 🎰 TucanBIT World-Class Cashback & Level System

## 🌟 Overview

This is a comprehensive, world-class cashback and level system designed for TucanBIT online casino. It provides a sophisticated reward system that keeps players engaged while maintaining fair and transparent rewards based on actual Gross Gaming Revenue (GGR) contribution.

## 🏆 Key Features

### ✨ Multi-Tier Cashback System
- **5 Tiers**: Bronze, Silver, Gold, Platinum, Diamond
- **Progressive Rewards**: Higher tiers = better cashback rates
- **Automatic Progression**: Users level up based on GGR contribution
- **Special Benefits**: Each tier offers unique perks

### 💰 Real-Time Processing
- **Kafka Integration**: Real-time bet event processing
- **Instant Calculation**: GGR and cashback calculated immediately
- **Automatic Updates**: User levels updated in real-time
- **Background Jobs**: Expired cashback processing

### 🎯 Advanced Features
- **House Edge Configuration**: Game-specific house edges
- **Promotion System**: Special boosts and bonuses
- **Expiring Cashback**: 30-day expiry to encourage claims
- **Daily/Weekly/Monthly Limits**: Responsible gaming limits
- **Comprehensive Admin Tools**: Full management capabilities

## 📊 Cashback Tiers

| Tier | Level | Min GGR | Cashback Rate | Daily Limit | Special Benefits |
|------|-------|---------|---------------|-------------|------------------|
| Bronze | 1 | $0 | 0.5% | $50 | Basic support |
| Silver | 2 | $1,000 | 1.0% | $100 | Priority support |
| Gold | 3 | $5,000 | 1.5% | $250 | Exclusive games |
| Platinum | 4 | $15,000 | 2.0% | $500 | Personal manager |
| Diamond | 5 | $50,000 | 2.5% | $1,000 | VIP events |

## 🎮 Game House Edges

| Game | House Edge | Min Bet |
|------|------------|---------|
| Plinko | 2.00% | $0.10 |
| Crash | 1.00% | $0.10 |
| Dice | 1.00% | $0.10 |
| Blackjack | 0.48% | $1.00 |
| Roulette | 2.70% | $1.00 |
| Slots | 3.00% | $0.10 |

## 🏗️ Architecture

### Database Schema
```
user_levels          - User level and progress tracking
cashback_tiers       - Configurable tier definitions
cashback_earnings    - Individual earning records
cashback_claims      - Claim history and status
game_house_edges     - Game-specific house edges
cashback_promotions  - Special promotions
```

### Service Layer
```
CashbackService      - Core business logic
CashbackStorage      - Data access layer
CashbackHandler      - HTTP API handlers
CashbackKafkaConsumer - Real-time event processing
```

## 🚀 API Endpoints

### User Endpoints
- `GET /user/cashback` - Get user cashback summary
- `POST /user/cashback/claim` - Claim available cashback
- `GET /user/cashback/earnings` - Get earnings history
- `GET /user/cashback/claims` - Get claims history

### Public Endpoints
- `GET /cashback/tiers` - Get all available tiers

### Admin Endpoints
- `GET /admin/cashback/stats` - Get comprehensive statistics
- `POST /admin/cashback/tiers` - Create new tier
- `PUT /admin/cashback/tiers/:id` - Update tier
- `POST /admin/cashback/promotions` - Create promotion

## 💡 How It Works

### 1. User Registration
When a user registers, a `user_levels` record is created with Bronze tier.

### 2. Bet Processing
1. User places bet → Kafka event published
2. Cashback consumer processes event
3. GGR calculated: `bet_amount × house_edge`
4. Cashback earned: `GGR × cashback_rate`
5. User level updated automatically

### 3. Level Progression
Users automatically progress to higher tiers based on total GGR:
- Bronze → Silver: $1,000 GGR
- Silver → Gold: $5,000 GGR
- Gold → Platinum: $15,000 GGR
- Platinum → Diamond: $50,000 GGR

### 4. Cashback Claiming
Users can claim available cashback:
- Respects daily/weekly/monthly limits
- Processes multiple earnings
- Credits user's wallet
- Tracks claim history

## 🔧 Implementation Details

### Real-Time Processing
```go
// Kafka event processing
func (c *CashbackKafkaConsumer) handleBetEvent(ctx context.Context, message []byte) error {
    var betEvent BetEvent
    json.Unmarshal(message, &betEvent)
    
    // Calculate GGR and cashback
    expectedGGR := betEvent.Amount.Mul(betEvent.HouseEdge)
    earnedCashback := expectedGGR.Mul(cashbackRate.Div(decimal.NewFromInt(100)))
    
    // Create earning record
    return c.cashbackService.ProcessBetCashback(ctx, bet)
}
```

### Cashback Calculation
```go
// Example calculation
betAmount := decimal.NewFromFloat(100.0)      // $100 bet
houseEdge := decimal.NewFromFloat(0.02)       // 2% house edge
cashbackRate := decimal.NewFromFloat(0.5)      // 0.5% cashback

expectedGGR := betAmount.Mul(houseEdge)       // $2.00
earnedCashback := expectedGGR.Mul(cashbackRate.Div(decimal.NewFromInt(100))) // $0.01
```

### Level Progression
```sql
-- Automatic level progression trigger
CREATE OR REPLACE FUNCTION check_and_update_user_level(p_user_id UUID)
RETURNS VOID AS $$
BEGIN
    -- Find highest tier user qualifies for
    SELECT ct.* INTO new_tier
    FROM cashback_tiers ct
    WHERE ct.min_ggr_required <= user_stats.total_ggr
    ORDER BY ct.tier_level DESC LIMIT 1;
    
    -- Update user level if higher tier found
    UPDATE user_levels SET current_level = new_tier.tier_level
    WHERE user_id = p_user_id;
END;
$$ LANGUAGE plpgsql;
```

## 📈 Benefits

### For Players
- **Fair Rewards**: Based on actual GGR contribution
- **Transparent**: Clear tier progression and limits
- **Engaging**: Gamification through level progression
- **Valuable**: Real cashback that can be claimed

### For Business
- **Player Retention**: Rewards keep players engaged
- **Increased Revenue**: Higher engagement = more bets
- **Data Insights**: Comprehensive analytics
- **Competitive Advantage**: World-class reward system

### For Operations
- **Automated**: Minimal manual intervention required
- **Scalable**: Handles high-volume real-time processing
- **Configurable**: Easy to adjust tiers and rates
- **Compliant**: Built with responsible gaming in mind

## 🛠️ Setup Instructions

### 1. Database Migration
```bash
# Run the migration
docker exec -it tucanbit-db psql -U tucanbit -d tucanbit -f /tmp/cashback_migration.sql
```

### 2. Service Integration
```go
// Initialize cashback service
cashbackStorage := cashback.NewCashbackStorage(db, logger)
cashbackService := cashback.NewCashbackService(cashbackStorage, logger)

// Initialize Kafka consumer
kafkaConsumer := kafka.NewKafkaController(config)
cashbackKafkaConsumer := cashback.NewCashbackKafkaConsumer(cashbackService, kafkaConsumer, logger)

// Start processing
cashbackKafkaConsumer.StartConsumer(ctx)
go cashbackKafkaConsumer.ProcessExpiredCashbackJob(ctx)
```

### 3. Route Registration
```go
// Register cashback routes
cashbackHandler := cashback.NewCashbackHandler(cashbackService, logger)
cashbackGlue := cashback.NewCashbackGlue(cashbackHandler, logger)
cashbackGlue.Init(router.Group("/api"))
```

## 🧪 Testing

### Test the System
```bash
# Run the test file
go run test_cashback_system.go
```

### Example API Calls
```bash
# Get user cashback summary
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/user/cashback

# Claim cashback
curl -X POST -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"amount": "10.00"}' \
     http://localhost:8080/user/cashback/claim

# Get admin statistics
curl -H "Authorization: Bearer <admin_token>" \
     http://localhost:8080/admin/cashback/stats
```

## 📊 Monitoring

### Key Metrics
- Total users with cashback
- Total cashback earned vs claimed
- Tier distribution
- Daily/weekly/monthly claim volumes
- Average cashback rates

### Admin Dashboard
- Real-time statistics
- User level distribution
- Cashback claim trends
- Promotion effectiveness
- System health metrics

## 🔒 Security & Compliance

### Data Protection
- All financial calculations use decimal precision
- Audit trails for all transactions
- Secure API endpoints with authentication
- Rate limiting on claim endpoints

### Responsible Gaming
- Daily/weekly/monthly limits
- Expiring cashback (30-day expiry)
- Transparent tier progression
- Admin oversight capabilities

## 🚀 Future Enhancements

### Planned Features
- **Referral Cashback**: Bonus cashback for referrals
- **Seasonal Promotions**: Special events and boosts
- **Mobile App Integration**: Push notifications for claims
- **Advanced Analytics**: Machine learning insights
- **Multi-Currency Support**: Support for different currencies

### Scalability
- **Microservices**: Split into separate services
- **Caching**: Redis caching for frequent queries
- **Load Balancing**: Handle high-volume traffic
- **Database Sharding**: Scale database operations

## 📞 Support

For technical support or questions about the cashback system:
- **Documentation**: This README file
- **Code Examples**: See `test_cashback_system.go`
- **API Documentation**: Swagger UI at `/swagger/`
- **Database Schema**: See migration files

---

## 🎉 Conclusion

This cashback system provides TucanBIT with a world-class reward system that:
- ✅ Keeps players engaged and loyal
- ✅ Provides fair and transparent rewards
- ✅ Scales to handle high-volume traffic
- ✅ Offers comprehensive admin tools
- ✅ Maintains responsible gaming practices

The system is production-ready and designed to compete with the best online casinos in the world! 🚀