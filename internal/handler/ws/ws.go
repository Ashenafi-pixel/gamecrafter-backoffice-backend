package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/handler"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ws struct {
	log                *zap.Logger
	betModule          module.Bet
	notificationModule module.Notification
	userWS             utils.UserWS
	userModule         module.User
}

func Init(log *zap.Logger, betModule module.Bet, notificationModule module.Notification, userWS utils.UserWS, userModule module.User) handler.WS {
	return &ws{
		log:                log,
		betModule:          betModule,
		notificationModule: notificationModule,
		userWS:             userWS,
		userModule:         userModule,
	}
}

// HandleWS Establish WebSocket connection for the user.
//
//	@Summary		WebSocket Connection
//	@Description	HandleWS sets up a WebSocket connection for a user to interact with the game in real time.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true					"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"	//	Status	101	indicates	successful	WebSocket	connection
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws [get]
func (w *ws) HandleWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	w.betModule.AddToBroadcastConnection(c, conn)
	for {
		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}
		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error(err.Error())
		}
		//verify user
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			err := fmt.Errorf("invalid or expired access_token ")
			_ = c.Error(err)
			c.Abort()
			return
		}

		// connect the user
		w.betModule.AddConnection(c, dto.BroadCastPayload{
			UserID: claims.UserID,
			Conn:   conn,
		})
		w.betModule.AddToBroadcastConnection(c, conn)

	}
	if conn != nil {
		conn.Close()
	}
}

// SinglePlayerStreamWS Establish WebSocket connection for the user.
//
//	@Summary		WebSocket Connection
//	@Description	SinglePlayerStreamWS sets up a WebSocket connection for a user to interact with the game in real time for single games.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true					"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"	//	Status	101	indicates	successful	WebSocket	connection
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws/single/player [get]
func (w *ws) SinglePlayerStreamWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}
	for {
		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}
		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error(err.Error())
		}
		//verify user
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			err := fmt.Errorf("invalid or expired access_token ")
			_ = c.Error(err)
			c.Abort()
			return
		}

		// connect the user
		w.betModule.AddToSingleGameConnections(c, dto.BroadCastPayload{
			UserID: claims.UserID,
			Conn:   conn,
		})
		conn.WriteJSON(gin.H{"message": "connected"})

	}
	if conn != nil {
		conn.Close()
	}
}

// PlayerLevelWS Establish WebSocket connection for the user to get level updates.
//
//	@Summary		WebSocket Connection for Player Level Updates
//	@Description	PlayerLevelWS sets up a WebSocket connection for a user to receive real-time updates on their level.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true	"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws/level/player [get]
func (w *ws) PlayerLevelWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	w.log.Info("WebSocket connection established for player level")

	for {

		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}

		w.log.Info("Received WebSocket message", zap.String("message", string(p)))

		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error("Failed to unmarshal WebSocket message", zap.Error(err))
			continue
		}

		w.log.Info("WebSocket message parsed",
			zap.String("type", message.Type),
			zap.String("access_token_length", fmt.Sprintf("%d", len(message.AccessToken))))

		//verify user
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			w.log.Error("JWT token validation failed", zap.Error(err))
			err := fmt.Errorf("invalid or expired access_token ")
			_ = c.Error(err)
			c.Abort()
			return
		}
		w.betModule.TriggerLevelResponse(c, claims.UserID)
		// Add debugging logs
		w.log.Info("JWT token parsed successfully",
			zap.String("user_id", claims.UserID.String()),
			zap.Bool("is_nil_uuid", claims.UserID == uuid.Nil))

		// connect the user
		w.betModule.AddPlayerLevelSocketConnection(c, claims.UserID, conn)

	}
	if conn != nil {
		conn.Close()
	}
}

// NotificationWS Notification WebSocket endpoint.
//
//	@Summary		Notification WebSocket
//	@Description	Establishes a WebSocket connection for receiving notification messages from internal services.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Success		101	{string}	string					"Switching Protocols"
//	@Failure		400	{object}	response.ErrorResponse	"Bad Request"
//	@Failure		500	{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws/notify [get]
func (w *ws) NotificationWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		return
	}
	defer conn.Close()

	_, p, err := conn.ReadMessage()
	if err != nil {
		w.log.Warn("Failed to read auth message", zap.Error(err))
		return
	}
	var msg dto.WSMessageRequest
	if err := json.Unmarshal(p, &msg); err != nil || msg.AccessToken == "" {
		conn.WriteJSON(gin.H{"status": "error", "message": "Missing or invalid access_token"})
		return
	}

	claims := &dto.Claim{}
	token, err := jwt.ParseWithClaims(msg.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		conn.WriteJSON(gin.H{"status": "error", "message": "Invalid or expired access_token"})
		return
	}
	userID := claims.UserID

	w.notificationModule.AddNotificationSocketConnection(userID, conn)
	conn.WriteJSON(gin.H{"status": "ok", "message": "Connected to notification socket"})

	// Keep the connection open for real-time notifications
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			w.notificationModule.RemoveNotificationSocketConnection(userID, conn)
			break
		}
	}
}

// PlayerProgressBarWS Establish WebSocket connection for the user to get level updates.
//
//	@Summary		WebSocket Connection for Player Progress Bar Updates
//	@Description	PlayerProgressBarWS sets up a WebSocket connection for a user to receive real-time updates on their level.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true	"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws/player/level/progress [get]
func (w *ws) PlayerProgressBarWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	w.log.Info("WebSocket connection established for player level")

	for {

		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}

		w.log.Info("Received WebSocket message", zap.String("message", string(p)))

		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error("Failed to unmarshal WebSocket message", zap.Error(err))
			continue
		}

		w.log.Info("WebSocket message parsed",
			zap.String("type", message.Type),
			zap.String("access_token_length", fmt.Sprintf("%d", len(message.AccessToken))))

		//verify user
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			w.log.Error("JWT token validation failed", zap.Error(err))
			err := fmt.Errorf("invalid or expired access_token ")
			_ = c.Error(err)
			c.Abort()
			return
		}
		w.betModule.TriggerPlayerProgressBar(c, claims.UserID)
		// Add debugging logs
		w.log.Info("JWT token parsed successfully",
			zap.String("user_id", claims.UserID.String()),
			zap.Bool("is_nil_uuid", claims.UserID == uuid.Nil))

		// connect the user
		w.betModule.AddPlayerProgressBarConnection(c, claims.UserID, conn)

	}
	if conn != nil {
		conn.Close()
	}
}

// InitiateTriggerSquadsProgressBar Establish WebSocket connection for the user to get squads progress bar updates.
//
//	@Summary		WebSocket Connection for Squads Progress Bar Updates
//	@Description	InitiateTriggerSquadsProgressBar sets up a WebSocket connection for a user to receive real-time updates on their squads progress bar.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true	"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error
//	@Router			/ws/squads/progress [get]
func (w *ws) InitiateTriggerSquadsProgressBar(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	w.log.Info("WebSocket connection established for squads progress bar")

	for {

		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}

		w.log.Info("Received WebSocket message", zap.String("message", string(p)))

		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error("Failed to unmarshal WebSocket message", zap.Error(err))
			continue
		}

		w.log.Info("WebSocket message parsed",
			zap.String("type", message.Type),
			zap.String("access_token_length", fmt.Sprintf("%d", len(message.AccessToken))))

		//verify user
		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			w.log.Error("JWT token validation failed", zap.Error(err))
			err := fmt.Errorf("invalid or expired access_token ")
			_ = c.Error(err)
			c.Abort()
			return
		}
		// Add debugging logs
		w.log.Info("JWT token parsed successfully",
			zap.String("user_id", claims.UserID.String()),
			zap.Bool("is_nil_uuid", claims.UserID == uuid.Nil))

		// connect the user
		w.betModule.AddSquadsProgressBarConnection(c, claims.UserID, conn)
		w.betModule.InitiateTriggerSquadsProgressBar(c, claims.UserID)

	}
	if conn != nil {
		conn.Close()
	}
}

// UserBalanceWS Establish WebSocket connection for the user to get balance updates.
//
//	@Summary		WebSocket Connection for User Balance Updates
//	@Description	UserBalanceWS sets up a WebSocket connection for a user to receive real-time updates on their balance.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Param			wsMessageReq	body		dto.WSMessageRequest	true	"WebSocket message request"
//	@Success		101				{object}	string					"Switching Protocols"
//	@Failure		400				{object}	response.ErrorResponse	"Bad Request"
//	@Failure		401				{object}	response.ErrorResponse	"Unauthorized"
//	@Failure		500				{object}	response.ErrorResponse	"Internal Server Error
//	@Router			/ws/balance/player [get]
func (w *ws) UserBalanceWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		_ = c.Error(err)
		return
	}

	w.log.Info("WebSocket connection established for player level")

	for {

		var message dto.WSMessageRequest
		_, p, err := conn.ReadMessage()
		if err != nil {
			w.log.Warn(err.Error())
			break
		}

		w.log.Info("Received WebSocket message", zap.String("message", string(p)))

		// bind message from websocket
		if err := json.Unmarshal(p, &message); err != nil {
			w.log.Error("Failed to unmarshal WebSocket message", zap.Error(err))
			continue
		}

		w.log.Info("WebSocket message parsed",
			zap.String("type", message.Type),
			zap.String("access_token_length", fmt.Sprintf("%d", len(message.AccessToken))))

		// Handle ping messages without requiring authentication
		if message.Type == "ping" {
			w.log.Info("Received ping message, responding with pong")
			pongMsg := map[string]string{"type": "pong"}
			if err := conn.WriteJSON(pongMsg); err != nil {
				w.log.Warn("Failed to send pong response", zap.Error(err))
			}
			continue
		}

		// For other message types, verify user
		if message.AccessToken == "" {
			w.log.Error("Missing access token for authenticated message", zap.String("type", message.Type))
			continue
		}

		claims := &dto.Claim{}
		token, err := jwt.ParseWithClaims(message.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			w.log.Error("JWT token validation failed", zap.Error(err))
			continue
		}
		// Add debugging logs
		w.log.Info("JWT token parsed successfully",
			zap.String("user_id", claims.UserID.String()),
			zap.Bool("is_nil_uuid", claims.UserID == uuid.Nil))

		// connect the user
		w.userWS.AddToPlayerBalanceWS(c, claims.UserID, conn)
		w.userWS.TriggerBalanceWS(c, claims.UserID)

	}
	if conn != nil {
		conn.Close()
	}
}

// SessionWS establishes a WebSocket connection for session monitoring.
//
//	@Summary		Session Monitoring WebSocket
//	@Description	Establishes a WebSocket connection for monitoring user session events.
//	@Tags			websocket
//	@Accept			json
//	@Produce		json
//	@Success		101	{string}	string					"Switching Protocols"
//	@Failure		400	{object}	response.ErrorResponse	"Bad Request"
//	@Failure		500	{object}	response.ErrorResponse	"Internal Server Error"
//	@Router			/ws/session [get]
func (w *ws) SessionWS(c *gin.Context) {
	key := viper.GetString("auth.jwt_secret")
	jwtKey := []byte(key)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		w.log.Error(err.Error())
		return
	}
	defer conn.Close()

	_, p, err := conn.ReadMessage()
	if err != nil {
		w.log.Warn("Failed to read auth message", zap.Error(err))
		return
	}
	var msg dto.WSMessageRequest
	if err := json.Unmarshal(p, &msg); err != nil || msg.AccessToken == "" {
		conn.WriteJSON(gin.H{"status": "error", "message": "Missing or invalid access_token"})
		return
	}

	claims := &dto.Claim{}
	token, err := jwt.ParseWithClaims(msg.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		conn.WriteJSON(gin.H{"status": "error", "message": "Invalid or expired access_token"})
		return
	}
	userID := claims.UserID

	w.userModule.AddSessionSocketConnection(userID, conn)

	// Keep the connection open for real-time session events
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			w.userModule.RemoveSessionSocketConnection(userID, conn)
			break
		}
	}
}
