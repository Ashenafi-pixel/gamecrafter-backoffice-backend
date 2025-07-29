package squads

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joshjones612/egyptkingcrash/internal/constant"
	"github.com/joshjones612/egyptkingcrash/internal/constant/dto"
	"github.com/joshjones612/egyptkingcrash/internal/constant/errors"
	"github.com/joshjones612/egyptkingcrash/internal/constant/model/response"
	"github.com/joshjones612/egyptkingcrash/internal/handler"
	"github.com/joshjones612/egyptkingcrash/internal/module"
	"go.uber.org/zap"
)

type squads struct {
	log          *zap.Logger
	squadsModule module.Squads
}

func Init(log *zap.Logger, squadsModule module.Squads) handler.Squads {
	return &squads{
		log:          log,
		squadsModule: squadsModule,
	}
}

// CreateSquad allows user to create a squad
//
//	@Summary		Create a squad
//	@Description	Creates a new squad for the authenticated user.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateSquadsReq	true	"Create squad request"
//	@Success		201		{object}	dto.CreateSquadsReq
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/squads [post]
func (s *squads) CreateSquad(c *gin.Context) {
	var req dto.CreateSquadsReq
	userID := c.GetString("user-id")
	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	req.Owner = parsedUserID

	resp, err := s.squadsModule.CreateSquads(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetMySquads allows user to get their squads
//
//	@Summary		Get my squads
//	@Description	Retrieves the squads the authenticated user is a member of.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]dto.GetSquadsResp
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/squads [get]
func (s *squads) GetMySquads(c *gin.Context) {
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetMySquads(c, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if len(resp) == 0 {
		response.SendSuccessResponse(c, http.StatusOK, []struct{}{})
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetMyOwnSquads allows user to get their own squads
//
//	@Summary		Get my own squads
//	@Description	Retrieves the squads owned by the authenticated user.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	[]dto.Squad
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/api/squads/own [get]
func (s *squads) GetMyOwnSquads(c *gin.Context) {
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetSquadsByOwnerID(c, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// UpdateSquadHandle allows user to update their own squads
//
//	@Summary		Update a squad
//	@Description	Updates a squad owned by the authenticated user.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.Squad	true	"Update squad request"
//	@Success		200		{object}	dto.Squad
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/squads/own [put]
func (s *squads) UpdateSquadHandle(c *gin.Context) {
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	var req dto.Squad
	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.UpdateSquadHandle(c, req, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteSquad allows user to delete a squad
//
//	@Summary		Delete a squad
//	@Description	Deletes a squad owned by the authenticated user.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			body		body		dto.DeleteSquadReq		true	"Delete squad request"
//	@Success		200			{object}	map[string]string		"Success message"
//	@Failure		400,401,500	{object}	response.ErrorResponse	"Error response"
//	@Router			/api/squads/delete [delete]
func (s *squads) DeleteSquad(c *gin.Context) {
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	var req dto.DeleteSquadReq
	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err = s.squadsModule.DeleteSquad(c, req.ID, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: constant.SUCCESS})
}

// CreateSquadMember allows user to create squad members
//
//	@Summary		Create a squad member
//	@Description	Adds a new member to a squad using user ID.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateSquadMemeberReq	true	"Create squad member request with user_id"
//	@Success		201		{object}	dto.SquadMember
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/squads/members [post]
func (s *squads) CreateSquadMember(c *gin.Context) {
	var req dto.CreateSquadMemeberReq

	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	req.OwnerID = parsedUserID
	resp, err := s.squadsModule.CreateSquadMember(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSquadMembersBySquadID allows user to get squad members by squad ID
//
//	@Summary		Get squad members by squad ID
//	@Description	Retrieves members of a specific squad by squad ID.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	true	"Page number"
//	@Param			per_page	query		string	true	"Items per page"
//	@Param			squad_id	query		string	true	"Squad ID"
//	@Success		200			{object}	dto.GetSquadMemebersRes
//	@Failure		401			{object}	response.ErrorResponse
//	@Router			/api/squads/members/bysquadid [get]
func (s *squads) GetSquadMembersBySquadID(c *gin.Context) {

	var req dto.GetSquadMemebersReq
	if err := c.ShouldBindQuery(&req); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			squadID := c.Query("squad_id")
			id, err := uuid.Parse(squadID)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			req.SquadID = id
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	resp, err := s.squadsModule.GetSquadMembersBySquadID(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if resp.SquadMemebers == nil {
		resp.SquadMemebers = []dto.GetSquadMembersResp{}
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// DeleteSquadMember allows user to delete a squad member
//
//	@Summary		Delete a squad member
//	@Description	Deletes a specific squad member for the authenticated user.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	map[string]string		"Success message"
//	@Param			member_id	query		string					true	"Member ID"
//	@Failure		400,401,500	{object}	response.ErrorResponse	"Error response"
//	@Router			/api/squads/members/:member_id [delete]
func (s *squads) DeleteSquadMember(c *gin.Context) {
	var req dto.DeleteSquadMemberReq
	memberID, _ := c.Params.Get("member_id")
	id, err := uuid.Parse(memberID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "member_id is not a valid UUID")
		s.log.Error(err.Error())
		_ = c.Error(err)
		return
	}
	req.MemberID = id
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err = s.squadsModule.DeleteSquadMember(c, req.MemberID, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: constant.SUCCESS})
}

// DeleteSquadMembersBySquadID allows user to delete all squad members
//
//	@Summary		Delete all squad members
//	@Description	Deletes all members of a specific squad for the authenticated user.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			body		body		dto.DeleteReq			true	"Delete squad members request"
//	@Success		200			{object}	map[string]string		"Success message"
//	@Failure		400,401,500	{object}	response.ErrorResponse	"Error response"
//	@Router			/api/squads/members/all [delete]
func (s *squads) DeleteSquadMembersBySquadID(c *gin.Context) {
	var req dto.DeleteReq
	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err = s.squadsModule.DeleteSquadMembersBySquadID(c, req.ID, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: constant.SUCCESS})
}

// GetSquadEarns allows user to get squad earnings
//
//	@Summary		Get squad earnings
//	@Description	Retrieves earnings for a specific squad.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		string	true	"Page number"
//	@Param			per_page		query		string	true	"Items per page"
//	@Param			squad_id		query		string	true	"Squad ID"
//	@Success		200				{object}	dto.GetSquadEarnsResp
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/earns [get]
func (s *squads) GetSquadEarns(c *gin.Context) {
	var req dto.GetSquadEarnsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			squadID := c.Query("squad_id")
			id, err := uuid.Parse(squadID)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			req.SquadID = id
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	resp, err := s.squadsModule.GetSquadEarns(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if resp.SquadsEarns == nil {
		resp.SquadsEarns = []dto.SquadEarns{}
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetMySquadEarns allows user to get their own squad earnings
//
//	@Summary		Get my squad earnings
//	@Description	Retrieves earnings for the authenticated user's squads.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		string	true	"Page number"
//	@Param			per_page		query		string	true	"Items per page"
//	@Param			squad_id		query		string	true	"Squad ID"
//	@Success		200				{object}	dto.GetSquadEarnsResp
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/my/squads/earns [get]
func (s *squads) GetMySquadEarns(c *gin.Context) {
	var req dto.GetSquadEarnsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			squadID := c.Query("squad_id")
			id, err := uuid.Parse(squadID)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			req.SquadID = id
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetMySquadEarns(c, req, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSquadTotalEarns allows user to get total squad earnings
//
//	@Summary		Get total squad earnings
//	@Description	Retrieves total earnings for a specific squad.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			squad_id		query		string	true	"Squad ID"
//	@Success		200				{object}	dto.GetSquadBalanceRes
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/total/earns [get]
func (s *squads) GetSquadTotalEarns(c *gin.Context) {
	var req dto.GetSquadBalanceReq
	if err := c.ShouldBindQuery(&req); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			squadID := c.Query("squad_id")
			id, err := uuid.Parse(squadID)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			req.SquadID = id
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}
	resp, err := s.squadsModule.GetSquadTotalEarns(c, req.SquadID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, dto.GetSquadBalanceRes{
		Earns: resp,
	})
}

// GetSquadsByType allows user to get squads by type
//
//	@Summary		Get squads by type
//	@Description	Gets all squads of a specific type.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			type	query		string	true	"Squad type"
//	@Success		200		{object}	dto.GetSquadsByTypeResp
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/squads/type [get]
func (s *squads) GetSquadsByType(c *gin.Context) {
	squadType := c.Param("type")
	if squadType == "" {
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("type parameter is required"), "type parameter is required")
		_ = c.Error(err)
		return
	}

	squads, err := s.squadsModule.GetSquadsByType(c, squadType)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, dto.GetSquadsByTypeResp{
		Squads: squads,
	})
}

// GetSquadsByName allows user to get squads by name
//
//	@Summary		Get squads by name
//	@Description	Gets all squads with a specific name.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	true	"Squad name"
//	@Success		200		{object}	dto.GetSquadsResp
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/api/squads/name [get]
func (s *squads) GetSquadByName(c *gin.Context) {
	squadName := c.Query("name")
	if squadName == "" {
		err := errors.ErrInvalidUserInput.Wrap(fmt.Errorf("name parameter is required"), "name parameter is required")
		_ = c.Error(err)
		return
	}

	squads, err := s.squadsModule.GetSquadByName(c, squadName)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, squads)
}

// GetTornamentStyleRanking allows user to get tournament style ranking
//
//	@Summary		Get tournament style ranking
//	@Description	Gets the ranking of squads in a tournament style format.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			per_page	query		string	true	"Items per page"
//	@Param			page		query		string	true	"Page number"
//	@Success		200			{object}	dto.GetTornamentStyleRankingResp
//	@Failure		400,401,500	{object}	response.ErrorResponse
//	@Router			/api/squads/tournament/ranking [get]
func (s *squads) GetTornamentStyleRanking(c *gin.Context) {
	var req dto.GetTornamentStyleRankingReq
	if err := c.ShouldBindQuery(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetTornamentStyleRanking(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// CreateTournaments
// CreateTournaments allows user to create a tournament
//
//	@Summary		Create a tournament
//	@Description	Creates a new tournament for the authenticated user.
//	@Tags			Admins
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.CreateTournamentReq	true	"Create tournament request"
//	@Success		201		{object}	dto.CreateTournamentResp
//	@Failure		401		{object}	response.ErrorResponse
//	@Router			/api/admins/tournaments [post]
func (s *squads) CreateTournaments(c *gin.Context) {
	var req dto.CreateTournamentReq
	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.CreateTournaments(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusCreated, resp)
}

// GetTornamentStyles allows user to get all tournament styles
//
//	@Summary		Get tournament styles
//	@Description	Retrieves all available tournament styles.
//	@Tags			Squads
//	@Param			Authorization	header	string	true	"Bearer <token>"
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	[]dto.Tournament
//	@Failure		400,401,500	{object}	response.ErrorResponse
//	@Router			/api/tournaments [get]
func (s *squads) GetTornamentStyles(c *gin.Context) {
	tournaments, err := s.squadsModule.GetTornamentStyles(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, tournaments)
}

// GetSquadMembersEarnings allows squad owner to get earnings for all squad members
//
//	@Summary		Get squad members earnings
//	@Description	Retrieves earnings for all members of a specific squad. Only squad owner can access this.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			page			query		string	true	"Page number"
//	@Param			per_page		query		string	true	"Items per page"
//	@Param			squad_id		query		string	true	"Squad ID"
//	@Success		200				{object}	dto.GetSquadMembersEarningsResp
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/members/earnings [get]
func (s *squads) GetSquadMembersEarnings(c *gin.Context) {
	var req dto.GetSquadMembersEarningsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		if strings.Contains(err.Error(), "is not valid value for uuid.UUID") {
			squadID := c.Query("squad_id")
			id, err := uuid.Parse(squadID)
			if err != nil {
				err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
				_ = c.Error(err)
				return
			}
			req.SquadID = id
		} else {
			err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
			_ = c.Error(err)
			return
		}
	}

	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetSquadMembersEarnings(c, req, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if resp.Members_Earning == nil {
		resp.Members_Earning = []dto.SquadMemberEarnings{}
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// LeaveSquad allows a user to leave their squad
//
//	@Summary		Leave a squad
//	@Description	User can leave their own squad.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Param			squad_id		query		string	true	"Squad ID"
//	@Success		200				{object}	dto.LeaveSquadResp
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/{squad_id}/members/leave [delete]
func (s *squads) LeaveSquad(c *gin.Context) {

	squaeID := c.Query("squad_id")
	id, err := uuid.Parse(squaeID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "squad id is not a valid UUID")
		s.log.Error(err.Error())
		_ = c.Error(err)
		return
	}
	var req dto.LeaveSquadReq
	req.SquadID = id

	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	req.UserID = parsedUserID

	err = s.squadsModule.LeaveSquad(c, req.UserID, req.SquadID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c,
		http.StatusOK, dto.LeaveSquadResp{
			Message: "You have successfully left the squad",
		})
}

// JoinSquad allows a user to join a squad
//
//	@Summary		Join a squad
//	@Description	User can join a squad by providing the squad ID.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	dto.SquadMemberResp
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/{squad_id}/members/join [post]
func (s *squads) JoinSquad(c *gin.Context) {
	squadID := c.Query("squad_id")
	id, err := uuid.Parse(squadID)
	if err != nil {
		err = errors.ErrInvalidUserInput.Wrap(err, "squad id is not a valid UUID")
		s.log.Error(err.Error())
		_ = c.Error(err)
		return
	}

	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.JoinSquad(c, parsedUserID, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// GetSquadWaitlist allows user to get squad waitlist
//
//	@Summary		Get squad waitlist
//	@Description	Retrieves the waitlist for a specific squad.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer <token>"
//	@Success		200				{object}	[]dto.WaitingSquadMember
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/waitlist [get]
func (s *squads) GetSquadWaitlist(c *gin.Context) {
	userID := c.GetString("user-id")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	resp, err := s.squadsModule.GetWaitingSquadMembers(c, parsedUserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, resp)
}

// RemoveWaitingSquadWaitingUser allows user to remove a waiting user from the squad waitlist
//
//	@Summary		Remove a waiting user from the squad waitlist
//	@Description	Removes a user from the squad waitlist.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token>"
//	@Param			body			body		dto.SquadMemberReq	true	"Remove waiting user request"
//	@Success		200				{object}	map[string]string	"Success message"
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/waitlist/remove [delete]
func (s *squads) RemoveWaitingSquadWaitingUser(c *gin.Context) {
	var req dto.SquadMemberReq

	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err := s.squadsModule.RemoveWaitingSquadWaitingUser(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: constant.SUCCESS})
}

// ApproveWaitingSquadMember allows user to approve a waiting squad member
//
//	@Summary		Approve a waiting squad member
//	@Description	Approves a user from the squad waitlist to join the squad.
//	@Tags			Squads
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer <token>"
//	@Param			body			body		dto.SquadMemberReq	true	"Approve waiting user request"
//	@Success		200				{object}	map[string]string	"Success message"
//	@Failure		400,401,500		{object}	response.ErrorResponse
//	@Router			/api/squads/waitlist/approve [post]
func (s *squads) ApproveWaitingSquadMember(c *gin.Context) {
	var req dto.SquadMemberReq

	if err := c.ShouldBind(&req); err != nil {
		s.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	err := s.squadsModule.ApproveWaitingSquadMember(c, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.SendSuccessResponse(c, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: constant.SUCCESS})
}
