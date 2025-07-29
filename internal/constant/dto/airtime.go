package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AirtimeLoginReq struct {
	Vaspid   int    `json:"vaspid"`
	Password string `json:"password"`
}

type AirtimeLoginResp struct {
	Success                bool   `json:"Success"`
	StatusCode             string `json:"StatusCode"`
	Message                string `json:"Message"`
	Provider               string `json:"Provider"`
	PisiAuthorizationToken string `json:"Pisi-authorization-token"`
	Pisisid                int    `json:"Pisisid"`
	Expiration             string `json:"Expiration"`
}

type AirtimeUtility struct {
	LocalID          uuid.UUID       `json:"local_id"`
	ID               int             `json:"id"`
	ProductName      string          `json:"productName"`
	BillerName       string          `json:"billerName"`
	Amount           string          `json:"amount"`
	IsAmountFixed    bool            `json:"isAmountFixed"`
	Status           string          `json:"status"`
	Price            decimal.Decimal `json:"price"`
	Timestamp        time.Time       `json:"timestamp"`
	TotalRedemptions decimal.Decimal `json:"total_redemptions"`
	TotalBuckets     decimal.Decimal `json:"total_buckets"`
}

type AirtimeUtilitiesResp struct {
	Data []AirtimeUtility `json:"data"`
}

type GetAirtimeUtilitiesData struct {
	TotalPages       int              `json:"total_pages"`
	AirtimeUtilities []AirtimeUtility `json:"airtime_utilities"`
}

type GetAirtimeUtilitiesResp struct {
	Message string                  `json:"message"`
	Data    GetAirtimeUtilitiesData `json:"data"`
}

type UpdateAirtimeStatusResp struct {
	Message string         `json:"message"`
	Data    AirtimeUtility `json:"data"`
}

type UpdateAirtimeStatusReq struct {
	LocalID uuid.UUID `json:"local_id"`
	Status  string    `json:"status"`
}

type UpdateAirtimeUtilityPriceReq struct {
	LocalID uuid.UUID       `json:"local_id"`
	Price   decimal.Decimal `json:"price"`
}

type UpdateAirtimeUtilityPriceRes struct {
	Message string         `json:"message"`
	Data    AirtimeUtility `json:"data"`
}

type ClaimPointsReq struct {
	AirtimeLocalID uuid.UUID `json:"airtime_local_id"`
	UserID         uuid.UUID `json:"user_id" swaggerignore:"true"`
}
type ClaimPointsData struct {
	TransactionID string               `json:"transaction_id"`
	Data          ClaimAirtimeRespData `json:"data"`
}
type ClaimPointsResp struct {
	Message string          `json:"message"`
	Data    ClaimPointsData `json:"data"`
}

type ClaimAirtimeReq struct {
	Msisdn               string `json:"msisdn"`
	CustomerId           string `json:"customerId"`
	UtilityPackageId     int    `json:"utilityPackageId"`
	TransactionReference string `json:"transactionReference"`
	Amount               int    `json:"amount"`
}

type ClaimAirtimeData struct {
	UtilityPackageId     int    `json:"utilityPackageId"`
	Amount               int    `json:"amount"`
	TransactionStatus    string `json:"transactionStatus"`
	TransactionReference string `json:"transactionReference"`
	BillerName           string `json:"billerName"`
	PackageName          string `json:"packageName"`
}
type ClaimAirtimeRespData struct {
	ResultCode int              `json:"result_code"`
	Data       ClaimAirtimeData `json:"data"`
}

type ClaimAirtimeResp struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ClaimAirtimeRespFromProvider struct {
	ResultCode    int    `json:"resultCode"`
	Description   string `json:"description"`
	TransactionID string `json:"transactionID"`
}

type AirtimeTransactions struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	TransactionID    string          `json:"transactionId"`
	Cashout          decimal.Decimal `json:"cashout"`
	BillerName       string          `json:"billerName"`
	UtilityPackageId int             `json:"utilityPackageId"`
	PackageName      string          `json:"packageName"`
	Amount           decimal.Decimal `json:"amount"`
	Status           string          `json:"status"`
	Timestamp        time.Time       `json:"timestamp"`
}

type GetAirtimeTransactionsRespData struct {
	TotalPages   int                   `json:"total_pages"`
	Transactions []AirtimeTransactions `json:"transactions"`
}
type GetAirtimeTransactionsResp struct {
	Message string                         `json:"message"`
	Data    GetAirtimeTransactionsRespData `json:"data"`
}

type UpdateAirtimeAmountReq struct {
	LocalID uuid.UUID       `json:"local_id"`
	Amount  decimal.Decimal `json:"amount"`
}

type AirtimeUtilitiesStats struct {
	TotalRedemptions       decimal.Decimal `json:"total_redemptions"`
	TotalBucksSpent        decimal.Decimal `json:"total_bucks_spent"`
	TotalActiveUtilities   int             `json:"total_active_utilities"`
	TotalInactiveUtilities int             `json:"total_inactive_utilities"`
	TotalUtilities         int             `json:"total_utilities"`
}
