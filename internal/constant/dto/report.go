package dto

import (
	"github.com/go-playground/validator/v10"
)

type DailyReportReq struct {
	Date string `form:"date" validate:"required,datetime=2006-01-02"`
}

type DailyReportRes struct {
	TotalPlayers int64             `json:"total_players"`
	NewPlayers   int64             `json:"new_players"`
	BucksEarned  int64             `json:"bucks_earned"`
	BucksSpent   int64             `json:"bucks_spent"`
	NetBucksFlow int64             `json:"net_bucks_flow"`
	Revenue      RevenueStream     `json:"revenue_streams"`
	Store        StoreTransaction  `json:"store_transactions"`
	Airtime      AirtimeConversion `json:"airtime_conversions"`
}

type RevenueStream struct {
	AdRevenueViews      int     `json:"ad_revenue_views"`
	AdRevenueBucks      int     `json:"ad_revenue_bucks"`
	AdRevenueAvgPerView float64 `json:"ad_revenue_avg_per_view"`
}

type StoreTransaction struct {
	Purchases         int     `json:"purchases"`
	TotalSpentBucks   int     `json:"total_spent_bucks"`
	AvgPerTransaction float64 `json:"avg_per_transaction"`
}

type AirtimeConversion struct {
	Conversions      int     `json:"conversions"`
	TotalValueBucks  int     `json:"total_value_bucks"`
	AvgPerConversion float64 `json:"avg_per_conversion"`
}

func ValidateDailyReportReq(req DailyReportReq) error {
	validate := validator.New()
	return validate.Struct(req)
}
