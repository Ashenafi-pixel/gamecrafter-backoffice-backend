package bet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"go.uber.org/zap"
)

// CreateLeague  create league.
//	@Summary		CreateLeague
//	@Description	CreateLeague allow admin  to create league
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.League	true	"create league  Request"
//	@Success		200	{object}	dto.League
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/leagues [post]
func (b *bet) CreateLeague(c *gin.Context) {
	var req dto.League
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.CreateLeague(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetLeagues  get league.
//	@Summary		GetLeagues
//	@Description	GetLeagues allow admin  to get leagues
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetLeagueRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/leagues [get]
func (b *bet) GetLeagues(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.GetLeagues(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateClub  create club.
//	@Summary		CreateClub
//	@Description	CreateClub allow admin  to create club
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.Club	true	"create club  Request"
//	@Success		200	{object}	dto.Club
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/clubs [post]
func (b *bet) CreateClub(c *gin.Context) {
	var req dto.Club
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.CreateClub(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetClubs  get club.
//	@Summary		GetClubs
//	@Description	GetClubs allow admin  to get leagues
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetClubRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/clubs [get]
func (b *bet) GetClubs(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.GetClubs(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateFootballCardMultiplierValue  get football bet multiplier.
//	@Summary		UpdateFootballCardMultiplierValue
//	@Description	UpdateFootballCardMultiplierValue allow admin  to get football bet multiplier
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.FootballCardMultiplier	true	"get  football multiplier  Request"
//	@Success		200	{object}	dto.Config
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/configs [put]
func (b *bet) UpdateFootballCardMultiplierValue(c *gin.Context) {
	var req dto.FootballCardMultiplier
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.UpdateFootballCardMultiplierValue(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetFootballCardMultiplier  get football bet multiplier
//	@Summary		GetFootballCardMultiplier
//	@Description	GetFootballCardMultiplier allow admin  to get football multiplier
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.Config
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/configs [get]
func (b *bet) GetFootballCardMultiplier(c *gin.Context) {
	resp, err := b.betModule.GetFootballCardMultiplier(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateFootballMatchRound  football match round.
//	@Summary		CreateFootballMatchRound
//	@Description	CreateFootballMatchRound allow admin  to create  football match round
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.FootballMatchRound
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds [post]
func (b *bet) CreateFootballMatchRound(c *gin.Context) {
	resp, err := b.betModule.CreateFootballMatchRound(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetFootballMatchRounds  get football bet multiplier
//	@Summary		GetFootballCardMultiplier
//	@Description	GetFootballCardMultiplier allow admin  to get football multiplier
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Param			page			query	string	true	"page type (required)"
//	@Param			per_page		query	string	true	"per_page type (required)"
//	@Produce		json
//	@Success		200	{object}	dto.GetRequest
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds [get]
func (b *bet) GetFootballMatchRounds(c *gin.Context) {
	var req dto.GetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.GetFootballMatchRounds(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateFootballMatch create football match for round.
//	@Summary		CreateFootballMatch
//	@Description	CreateFootballMatch allow admin  to create  football match for  round
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.FootballMatchReq	true	"create football match for round  Request"
//	@Success		200	{object}	[]dto.FootballMatchReq
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds [post]
func (b *bet) CreateFootballMatch(c *gin.Context) {
	var req []dto.FootballMatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.CreateFootballMatch(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetFootballRoundMatchs get football round matches.
//	@Summary		GetFootballRoundMatchs
//	@Description	GetFootballRoundMatchs allow admin to  get football round matches.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.GetFootballRoundMatchesReq	true	"get football round matches  Request"
//	@Success		200	{object}	dto.GetFootballRoundMatchesRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds/mathces [post]
func (b *bet) GetFootballRoundMatchs(c *gin.Context) {
	var req dto.GetFootballRoundMatchesReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.GetFootballRoundMatchs(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetCurrentFootballRound  get current active round.
//	@Summary		GetCurrentFootballRound
//	@Description	GetCurrentFootballRound allow user to get active round
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetFootballRoundMatchesRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/football/rounds/current [get]
func (b *bet) GetCurrentFootballRound(c *gin.Context) {
	resp, err := b.betModule.GetCurrentFootballRound(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CloseFootballMatch allow admin to close match.
//	@Summary		CloseFootballMatch
//	@Description	CloseFootballMatch allow admin to  close match.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.CloseFootballMatchReq	true	"get football round matches  Request"
//	@Success		200	{object}	dto.FootballMatch
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds/matches [patch]
func (b *bet) CloseFootballMatch(c *gin.Context) {
	var req dto.CloseFootballMatchReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.CloseFootballMatch(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateFootballRoundPrice allow admin to update football round price.
//	@Summary		UpdateFootballRoundPrice
//	@Description	UpdateFootballRoundPriceallow admin to update football round price.
//	@Tags			Admin
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UpdateFootballBetPriceReq	true	"get football round price  Request"
//	@Success		200	{object}	dto.UpdateFootballBetPriceRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/admin/football/rounds/prices [post]
func (b *bet) UpdateFootballRoundPrice(c *gin.Context) {
	var req dto.UpdateFootballBetPriceReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	resp, err := b.betModule.UpdateFootballRoundPrice(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetFootballRoundPrice  get current round price.
//	@Summary		GetFootballRoundPrice
//	@Description	GetFootballRoundPrice allow user to get get current round price
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.UpdateFootballBetPriceRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/football/rounds/prices [get]
func (b *bet) GetFootballRoundPrice(c *gin.Context) {
	resp, err := b.betModule.GetFootballMatchPrice(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// PleceBetOnFootballRound allow user to place football match bet
//	@Summary		PleceBetOnFootballRound
//	@Description	PleceBetOnFootballRound allow user to place football match bet.
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Accept			json
//	@Produce		json
//	@Param			req	body		dto.UserFootballMatchBetReq	true	"allow users to place football bet"
//	@Success		200	{object}	dto.UserFootballMatchBetRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/football/rounds/bets [post]
func (b *bet) PleceBetOnFootballRound(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.UserFootballMatchBetReq
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = userIDParsed
	resp, err := b.betModule.PleceBetOnFootballRound(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetUserFootballBets  get football bets.
//	@Summary		GetUserFootballBets
//	@Description	GetUserFootballBets allow user to get bet history
//	@Tags			Bet
//	@Param			Authorization	header	string	true	"Bearer <token> "
//	@Produce		json
//	@Success		200	{object}	dto.GetUserFootballBetRes
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/football/rounds/bets [get]
func (b *bet) GetUserFootballBets(c *gin.Context) {
	userID := c.GetString("user-id")

	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("userID", userID))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.GetRequest
	if err := c.ShouldBind(&req); err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := b.betModule.GetUserFootballBets(c, req, userIDParsed)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}
