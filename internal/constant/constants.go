package constant

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant/dto"
)

const (
	USER_REGISTRATION_SUCCESS       = "User registered successfully"
	LOGIN_SUCCESS                   = "Login successful"
	DEFAULT_CURRENCY                = "P"
	DEPOSIT                         = "deposit"
	TRANSFER                        = "transfer"
	FUND                            = "fund"
	ADD_FUND                        = "add_fund"
	REMOVE_FUND                     = "remove_fund"
	INTERNAL_TRANSACTION            = "internal_transaction"
	WITHDRAWAL                      = "withdrawal"
	REAL_MONEY                      = "real_money"
	BONUS_MONEY                     = "bonus_money"
	POINTS                          = "points"
	BALANCE_SUCCESS                 = "Balance updated successfully."
	UPDATE_PASSWORD_SUCCESS         = "Password updated successfully."
	BLOCK_USER_SUCCESS              = "user successfully blocked."
	CANCEL_SUCCESS                  = "bet round successfully canceled."
	SUCCESS                         = "success"
	ACTIVE                          = "ACTIVE"
	PENDING                         = "PENDING"
	CLOSED                          = "CLOSED"
	CANCELED                        = "CANCELED"
	INVALID_USER_INPUT              = "invalid otp"
	COMPLTE                         = "COMPLTED"
	CONFIG_POINT_MULTIPLIER         = "point_multiplier"
	CONFIG_POINT_MULTIPLIER_DEFAULT = "1"
	REFERAL_POINT                   = "referal_point"
	POINT_CURRENCY                  = "P"
	NGN_CURRENCY                    = "NGN"
	WON                             = "WON"
	LOSE                            = "LOSE"
)
const (
	BET_STATUS_COMPLETED = "completed"
	BET_STATUS_FAILED    = "failed"
)

const (
	BET_INPROGRESS              = "in_progress"
	BET_OPEN                    = "open"
	BET_CLOSED                  = "closed"
	BET_CRASHPOINT_REACHED      = "crash point is reached"
	BET_NO_OPEN_ROUND_AVAILABLE = "no open round available"
	BET_SUCCESS_MESSAGE         = "Bet placed successfully."
	FORGOT_PASSWORD_RES         = "we will send you otp if the user is exist"
	FORGOT_PASSWORD_SUBJECT     = "Password Reset Request"
	BET_OPERATIONAL_TYPE        = "place_bet"
	BET_CASHOUT                 = "bet_cashout"
	BET_LOTTERY_CASHOUT         = "lottery_cashout"
	BET_LOTTERY_BUY             = "buy_lottery"

	CASHOUT      = "cashout"
	BET_CANCELED = "bet_canceled"
	ROUND_FAILED = "failed"
	REFUND       = "refund"
)

const (
	WS_CURRENT_MULTIPLIER = "current multiplier"
	WS_CASHOUT            = "cashout"
	WS_CRASH              = "crash"
)

var (
	GAME_TUCAKBIT                       = uuid.MustParse("cfb2c688-0d30-46ea-ba7e-6ee2b29a8443")
	GAME_TUCAKBIT_DEFAULT_NAME          = "TucanBIT"
	GAME_PLINKO                         = uuid.MustParse("843495fe-c0b7-451f-b1d2-e68b86d06008")
	GAME_PLINKO_DEFAULT_NAME            = "Plinko"
	GAME_CRYPTO_KINGS                   = uuid.MustParse("e567e3b0-a432-4062-84e5-dca294aa2479")
	GAME_CRYPTO_KINGS_DEFAULT_NAME      = "Crypto_kings"
	GAME_FOOTBALL_FIXTURES              = uuid.MustParse("f144c263-911f-4bf6-b1d9-90b8efb92a9d")
	GAME_FOOTBALL_FIXTURES_DEFAULT_NAME = "Football fixtures"
	GAME_QUICK_HUSTLE                   = uuid.MustParse("8d2ca9f6-1c0e-46ca-975f-72ca3afed060")
	GAME_QUICK_HUSTLE_DEFAULT_NAME      = "Quick hustle"
	GAME_ROLL_DA_DICE                   = uuid.MustParse("b2a8cd89-83a0-40b5-8803-46a079ac245b")
	GAME_ROLL_DA_DICE_DEFAULT_NAME      = "Roll Da Dice"
	GAME_SCRATCH_CARD                   = uuid.MustParse("22ab4676-6657-410b-b030-4344d0ee1937")
	GAME_SCRATCH_CARD_DEFAULT_NAME      = "Scratch Card"
	GAME_SPINNING_WHEEL                 = uuid.MustParse("66f1020e-d9c8-4152-94c0-fc7b812b0016")
	GAME_SPINNING_WHEEL_DEFAULT_NAME    = "Spinning Wheel"
	GAME_STREET_KINGS                   = uuid.MustParse("5a729589-987d-4862-b1a5-3c32831da50d")
	GAME_STREET_KINGS_DEFAULT_NAME      = "Street Kings"
	GAME_LOOT_BOX                       = uuid.MustParse("5a729589-987d-4862-b1a5-3c32831da50d")
	GAME_LOOT_BOX_DEFAULT_NAME          = "Loot Box"
)

var GAMES = []uuid.UUID{GAME_TUCAKBIT, GAME_PLINKO, GAME_CRYPTO_KINGS, GAME_FOOTBALL_FIXTURES, GAME_QUICK_HUSTLE, GAME_ROLL_DA_DICE, GAME_SCRATCH_CARD, GAME_SPINNING_WHEEL, GAME_STREET_KINGS}

var VALID_IMGS = []string{".png", ".jpeg", ".jpg", ".svg"}

func GetOTPBody(otp string) string {

	return `Subject: Password Reset Request
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
  <head>
    <style>
      body {
        font-family: Arial, sans-serif;
        background-color: #f4f4f9;
        color: #333;
        margin: 0;
        padding: 0;
      }
      .container {
        width: 100%;
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
        background-color: #fff;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
      }
      .header {
        text-align: center;
        margin-bottom: 20px;
      }
      h1 {
        color: #4CAF50;
      }
      .otp-code {
        font-size: 24px;
        font-weight: bold;
        color: #333;
        background-color: #f1f1f1;
        padding: 10px;
        border-radius: 5px;
        text-align: center;
      }
      .footer {
        text-align: center;
        margin-top: 20px;
        font-size: 12px;
        color: #777;
      }
      .footer a {
        color: #4CAF50;
        text-decoration: none;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="header">
        <h1>Password Reset Request</h1>
      </div>
      <p>TucanBIT</p>
      <p>You have requested to reset your password. To complete the process, please use the following One-Time Password (OTP):</p>
      <div class="otp-code">` + otp + `</div>
      <p>If you did not request this change, please ignore this email or contact support immediately.</p>
      <div class="footer">
        <p>Thank you,</p>
      </div>
    </div>
  </body>
</html>`
}

func GetUpdateProfileOTPTemplate(otp string) string {

	return `Subject: Profile Update Request
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
  <head>
    <style>
      body {
        font-family: Arial, sans-serif;
        background-color: #f4f4f9;
        color: #333;
        margin: 0;
        padding: 0;
      }
      .container {
        width: 100%;
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
        background-color: #fff;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
      }
      .header {
        text-align: center;
        margin-bottom: 20px;
      }
      h1 {
        color: #4CAF50;
      }
      .otp-code {
        font-size: 24px;
        font-weight: bold;
        color: #333;
        background-color: #f1f1f1;
        padding: 10px;
        border-radius: 5px;
        text-align: center;
      }
      .footer {
        text-align: center;
        margin-top: 20px;
        font-size: 12px;
        color: #777;
      }
      .footer a {
        color: #4CAF50;
        text-decoration: none;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="header">
        <h1>Profile Update Request</h1>
      </div>
      <p>Dear User,</p>
      <p>You have requested to update your profile. To complete the process, please use the following One-Time Password (OTP):</p>
      <div class="otp-code">` + otp + `</div>
      <p>If you did not request this change, please ignore this email or contact support immediately.</p>
      <div class="footer">
        <p>Thank you,</p>
      </div>
    </div>
  </body>
</html>`
}

func GetBlockedUserString(notification dto.NotifyDepartmentsReq) string {
	return `Subject: User Blocked Notification
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
  <head>
    <style>
      body {
        font-family: Arial, sans-serif;
        background-color: #f4f4f9;
        color: #333;
        margin: 0;
        padding: 0;
      }
      .container {
        width: 100%;
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
        background-color: #fff;
        border-radius: 8px;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
      }
      .header {
        text-align: center;
        margin-bottom: 20px;
      }
      h1 {
        color: #FF6347;
      }
      table {
        width: 100%;
        border-collapse: collapse;
        margin-top: 20px;
      }
      table, th, td {
        border: 1px solid #ddd;
      }
      th, td {
        padding: 12px;
        text-align: left;
      }
      th {
        background-color: #f4f4f4;
        color: #333;
      }
      .footer {
        text-align: center;
        margin-top: 20px;
        font-size: 12px;
        color: #777;
      }
      .footer a {
        color: #FF6347;
        text-decoration: none;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="header">
        <h1>User Blocked Notification</h1>
      </div>
      <p>Dear User,</p>
      <p>We are notifying you that a user has been blocked. Below are the details of the blocked user:</p>
      
      <table>
        <tr>
          <th>Block Type</th>
          <td>` + notification.BlockReq.Type + `</td>
        </tr>
        <tr>
          <th>Block Duration</th>
          <td>` + notification.BlockReq.Duration + `</td>
        </tr>
        <tr>
          <th>Blocked User Username</th>
          <td>` + notification.BlockedUser.FirstName + `</td>
        </tr>
        <tr>
          <th>Blocked User Phone</th>
          <td>` + notification.BlockedUser.PhoneNumber + `</td>
        </tr>
        <tr>
          <th>Blocked User Email</th>
          <td>` + notification.BlockedUser.Email + `</td>
        </tr>
        <tr>
          <th>Blocked User First Name</th>
          <td>` + notification.BlockedUser.FirstName + `</td>
        </tr>
        <tr>
          <th>Blocked User Last Name</th>
          <td>` + notification.BlockedUser.LastName + `</td>
        </tr>
        <tr>
          <th>Blocked By</th>
          <td>` + notification.BlockerUser.FirstName + `</td>
        </tr>
        <tr>
          <th>Blocker Email</th>
          <td>` + notification.BlockerUser.Email + `</td>
        </tr>
        <tr>
          <th>Reason</th>
          <td>` + notification.BlockReq.Reason + `</td>
        </tr>
        <tr>
          <th>Note</th>
          <td>` + notification.BlockReq.Note + `</td>
        </tr>
      </table>

      <p>If you did not request this change or have any concerns, please ignore this email or contact our support team immediately.</p>
      
      <div class="footer">
        <p>Thank you,</p>
        <p>Your Support Team</p>
      </div>
    </div>
  </body>
</html>
`
}

var GOOGLE_OAUTH_SCOPES = []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}
var FACEBOOK_OAUTH_SCOPES = []string{"email", "public_profile", "user_friends", "user_photos", "user_birthday"}
var FACEBOOK_OAUTH_RESPONSE_REQ = "https://graph.facebook.com/me?fields=id,name,email,first_name,middle_name,last_name"
var SOURCE_GMAIL = "email"
var SOURCE_FACEBOOK = "facebook"
var SOURCE_PHONE = "facebook"

const (
	BLOCK_DURATION_PERMANENT = "permanent"
	BLOCK_DURATION_TEMPORARY = "temporary"

	BLOCK_TYPE_FINANCIAL = "financial"
	BLOCK_TYPE_GAMING    = "gaming"
	BLOCK_TYPE_LOGIN     = "login"
	BLOCK_TYPE_COMPLETE  = "complete"

	IP_FILTER_TYPE_DENY  = "deny"
	IP_FILTER_TYPE_ALLOW = "allow"
)

const (
	SORT_QUERY_DESC = "desc"
	SORT_QUERY_ASC  = "asc"
)

const (
	PLINKO_MIN_BET             = "plinko_min_bet"
	PLINKO_MAX_BET             = "plinko_max_bet"
	PLINKO_RTP                 = "plinko_rtp"
	PLINKO_MULTIPLIERS         = "plinko_multipliers"
	PLINKO_DEFAULT_MIN_BET     = "0.10"
	PLINKO_DEFAULT_MAX_BET     = "100"
	PLINKO_DEFUALT_RTP         = "97"
	PLINKO_DEFUALT_MULTIPLIERS = "{0.2,0.4,0.6,0.8,1,1.5,3,5,10,30}"
)

const (
	FOOTBALL_MATCH_STATUS_ACTIVE             = "ACTIVE"
	CONFIG_FOOTBALL_MATCH_MULTIPLIER         = "fb_match_multiplier"
	CONFIG_FOOTBALL_MATCH_CARD_PRICE         = "fb_match_price"
	CONFIG_FOOTBALL_DEFAULT_MATCH_CARD_PRICE = "3"
	CONFIG_FOOTBALL_MATCH_MULTIPLIER_DEFAULT = "3"
	FOOTBALL_WON                             = "WON"
	FOOTBALL_LOSE                            = "LOSE"
	FOOTBALL_DRAW                            = "DRAW"
	FOOTBALL_HOME_WON                        = "HOME"
	FOOTBALL_AWAY_WON                        = "AWAY"
)

const SIGNUP_BONUS = "signup_bonus"
const REFERRAL_BONUS = "referral_bonus"

var (
	STREET_KINGS_VERSIONS = map[string]string{
		"v1": "street kings",
		"v2": "street kings 2",
	}
)

const (
	CRYPTO_KINGS_TIME_MAX_VALUE          = "c_k_t_m_v"
	CRYPTO_KINGS_TIME_MAX_VALUE_DEFAULT  = "2500"
	CRYPTO_KINGS_RANGE_MAX_VALUE         = "c_k_r_m_v"
	CRYPTO_KINGS_RANGE_MAX_VALUE_DEFAULT = "1808000"
	CRYPTO_KING_DEFAULT_CURRENT_VALUE    = 30000
	CRYPTO_KING_HIGHER                   = "HIGHER"
	CRYPTO_KING_LOWER                    = "LOWER"
	CRYPTO_KING_RANGE                    = "RANGE"
	CRYPTO_KING_TIME                     = "TIME"
)

var (
	CRYPTO_KINGS_PLACE_BET_TYPES = map[string]bool{"TIME": true, "RANGE": true, "HIGHER": true, "LOWER": true}
)

const (
	QUICK_HUSTLE_HIGHER = "HIGHER"
	QUICK_HUSTLE_LOWER  = "LOWER"
)

const (
	SCRATCH_CARD_PRICE                = "scratch_price"
	SCRATCH_CARD_PRICE_DEFAULT        = 100 // it would be 100 points
	SCRATCH_CAR                string = "scratch_car"
	SCRATCH_CAR_DEFAULT        int    = 25000
	SCRATCH_DOLLAR             string = "scratch_dollar"
	SCRATCH_DOLLAR_DEFAULT     int    = 50000
	SCRATCH_CRAWN              string = "scratch_crawn"
	SCRATCH_CRAWN_DEFAULT      int    = 10000
	SCRATCH_CENT               string = "scratch_cent"
	SCRATCH_CENT_DEFAULT       int    = 1000
	SCRATCH_DIAMOND            string = "scratch_diamond"
	SCRATCH_DIAMOND_DEFAULT    int    = 2500
	SCRATCH_CUP                string = "scratch_cup"
	SCRATCH_CUP_DEFAULT        int    = 5000
)

const (
	SPINNING_WHEEL               = "spinning_wheel"
	SPINNING_WHEEL_DEFAULT_PRICE = 50
	LOOT_BOX                     = "slot_box"
	LOOT_BOX_DEFAULT_PRICE       = 100
)

var (
	SPINNING_WHEEL_VALUES_MAP = make(map[string]decimal.Decimal)
	SPINNING_WHEEL_VALUES     = []string{"25", "500", "better", "mystery", "50", "1000", "100", "mystery", "50", "better", "250", "mystery", "5000"}
	SPINNING_WHEEL_MYSTERY    = []decimal.Decimal{decimal.NewFromInt(50), decimal.NewFromInt(100), decimal.NewFromInt(500), decimal.NewFromInt(1000), decimal.NewFromInt(5000)}
)

const (
	INACTIVE                         = "INACTIVE"
	AIRTIME_SUCCESS_STATUS_CODE      = "1000"
	AIRTIME_LOGIN_ENDPOINT           = "/v1/Authentication/login"
	AIRTIME_GET_UTILITIES_ENDPOINT   = "/v1/utilities/getutilitypackages"
	AIRTIME_PAY_UTILITY              = "/v1/utilities/payutilities"
	AIRTIME_CHECK_TRANSACTION_STATUS = "/v1/utilities/checktransactionstatus?transactionReference="
	AIRTIME_STATUS_ACTIVE            = "ACTIVE"
	AIRTIME_STATUS_INACTIVE          = "INACTIVE"
)

const (
	SQUAD_CREATED_SUCCESS  = "success"
	SQUAD_ADDED_TO_WAITING = "squad waiting for approval"
)

const (
	LEVEL_TYPE_PLAYER = "players"
	LEVEL_TYPE_SQUAD  = "squads"
)

type ContextKey any

const (
	LOTTERY_DRAW_SYNC = "lottery.draw.sync"
)
