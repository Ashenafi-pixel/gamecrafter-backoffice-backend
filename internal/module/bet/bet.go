package bet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/dto"
	"github.com/tucanbit/internal/constant/errors"
	"github.com/tucanbit/internal/module"
	"github.com/tucanbit/internal/storage"
	"github.com/tucanbit/platform/utils"
	"go.uber.org/zap"
)

type bet struct {
	log                          *zap.Logger
	betStorage                   storage.Bet
	balanceStorage               storage.Balance
	betMax                       decimal.Decimal
	cron                         *cron.Cron
	InProgressRounds             map[uuid.UUID]dto.BetRound
	streaming                    map[uuid.UUID]bool
	betOpenForDuration           time.Duration
	userConn                     map[uuid.UUID]map[*websocket.Conn]bool
	userSingleGameConnection     map[uuid.UUID]map[*websocket.Conn]bool
	betRoundMultiplayerHolder    map[uuid.UUID]decimal.Decimal
	mutex                        sync.Mutex
	balanceLogStorage            storage.BalanceLogs
	locker                       map[uuid.UUID]*sync.Mutex
	operationalGroupStorage      storage.OperationalGroup
	operationalGroupTypeStorage  storage.OperationalGroupType
	bucketName                   string
	broadcastConn                map[*websocket.Conn]bool
	lowerCounter                 int
	streetkingsLocker            *sync.Mutex
	userStorage                  storage.User
	currentRound                 map[string]dto.BetRound
	socketLocker                 map[*websocket.Conn]*sync.Mutex
	ConfigStorage                storage.Config
	activeSingleBet              map[uuid.UUID]decimal.Decimal
	activeSingleBetSync          map[uuid.UUID]*sync.Mutex
	singleBetSocketSync          map[uuid.UUID]*sync.Mutex
	crptoKingsWonLoseMap         map[uuid.UUID]int
	cryptoKingsCurrentUsersValue map[uuid.UUID]decimal.Decimal
	cryptoKingsRangeBets         map[uuid.UUID]int
	cryptoKingsBetLocker         *sync.Mutex
	quickHustleCardsMap          map[string]int
	quickHustelHigherMultiplier  map[string]decimal.Decimal
	quickHustelLowerMultiplier   map[string]decimal.Decimal
	scratchCardsPriceHolder      map[string]decimal.Decimal
	scratchMaxPrice              decimal.Decimal
	scratchBets                  decimal.Decimal
	spinningwheelBets            decimal.Decimal
	spinningwheelFreeSpins       map[uuid.UUID]int
	spinningWheelFreeSpinsLocker sync.Mutex
	squadsStorage                storage.Squads
	playerLevelSocket            map[uuid.UUID]map[*websocket.Conn]bool
	playerLevelSocketLocker      map[*websocket.Conn]*sync.Mutex
	playerProgressSocket         map[uuid.UUID]map[*websocket.Conn]bool
	playerProgressSocketLocker   map[*websocket.Conn]*sync.Mutex
	SquadsProgressSocket         map[uuid.UUID]map[*websocket.Conn]bool
	SquadsProgressSocketLocker   map[*websocket.Conn]*sync.Mutex
	userWS                       utils.UserWS
}

func Init(betStorage storage.Bet,
	balanceStorage storage.Balance,
	log *zap.Logger,
	betMax decimal.Decimal,
	betOpenForDuration time.Duration,
	locker map[uuid.UUID]*sync.Mutex,
	operationalGroupStorage storage.OperationalGroup,
	operationalGroupTypeStorage storage.OperationalGroupType,
	balanceLogStorage storage.BalanceLogs,
	bucketName string,
	userStorage storage.User,
	config storage.Config,
	squadsStorage storage.Squads,
	userWS utils.UserWS,

) module.Bet {
	// initialize crone job
	betPointer := &bet{
		log:                          log,
		betStorage:                   betStorage,
		streaming:                    make(map[uuid.UUID]bool),
		userConn:                     make(map[uuid.UUID]map[*websocket.Conn]bool),
		betMax:                       betMax,
		balanceStorage:               balanceStorage,
		InProgressRounds:             make(map[uuid.UUID]dto.BetRound),
		betRoundMultiplayerHolder:    make(map[uuid.UUID]decimal.Decimal),
		locker:                       locker,
		betOpenForDuration:           betOpenForDuration,
		operationalGroupStorage:      operationalGroupStorage,
		operationalGroupTypeStorage:  operationalGroupTypeStorage,
		balanceLogStorage:            balanceLogStorage,
		bucketName:                   bucketName,
		broadcastConn:                map[*websocket.Conn]bool{},
		userStorage:                  userStorage,
		currentRound:                 make(map[string]dto.BetRound),
		socketLocker:                 make(map[*websocket.Conn]*sync.Mutex),
		ConfigStorage:                config,
		userSingleGameConnection:     map[uuid.UUID]map[*websocket.Conn]bool{},
		activeSingleBet:              make(map[uuid.UUID]decimal.Decimal),
		activeSingleBetSync:          make(map[uuid.UUID]*sync.Mutex),
		crptoKingsWonLoseMap:         map[uuid.UUID]int{},
		cryptoKingsCurrentUsersValue: map[uuid.UUID]decimal.Decimal{},
		cryptoKingsRangeBets:         make(map[uuid.UUID]int),
		cryptoKingsBetLocker:         &sync.Mutex{},
		singleBetSocketSync:          make(map[uuid.UUID]*sync.Mutex),
		quickHustleCardsMap:          map[string]int{},
		scratchCardsPriceHolder:      map[string]decimal.Decimal{},
		scratchMaxPrice:              decimal.Decimal{},
		scratchBets:                  decimal.Decimal{},
		spinningwheelBets:            decimal.Decimal{},
		spinningwheelFreeSpins:       make(map[uuid.UUID]int),
		spinningWheelFreeSpinsLocker: sync.Mutex{},
		quickHustelHigherMultiplier:  make(map[string]decimal.Decimal),
		quickHustelLowerMultiplier:   make(map[string]decimal.Decimal),
		squadsStorage:                squadsStorage,
		playerLevelSocket:            make(map[uuid.UUID]map[*websocket.Conn]bool),
		playerLevelSocketLocker:      make(map[*websocket.Conn]*sync.Mutex),
		playerProgressSocket:         make(map[uuid.UUID]map[*websocket.Conn]bool),
		playerProgressSocketLocker:   make(map[*websocket.Conn]*sync.Mutex),
		SquadsProgressSocket:         make(map[uuid.UUID]map[*websocket.Conn]bool),
		SquadsProgressSocketLocker:   make(map[*websocket.Conn]*sync.Mutex),
		userWS:                       userWS,
		streetkingsLocker:            &sync.Mutex{},
	}
	err := betPointer.initializeBetEngine()
	if err != nil {
		betPointer.log.Error("Failed to initialize bet engine", zap.Error(err))
		betPointer.log.Error("Continuing without bet engine - this is expected for new installations")
	}
	return betPointer
}

func (b *bet) initializeBetEngine() error {
	//check for failed games
	failedBets, err := b.betStorage.GetFailedRounds(context.Background())
	if err != nil {
		b.log.Error("Failed to get failed rounds", zap.Error(err))
		b.log.Error("Continuing without failed rounds check - this is expected for new installations")
	} else if len(failedBets) > 0 {
		b.RefundFailedRounds(context.Background(), failedBets)
	}
	b.cron = cron.New(cron.WithSeconds())
	_, err = b.cron.AddFunc("@every 1s", b.CreateRounds)
	if err != nil {
		return err
	}
	b.cron.Start()
	if err := b.InitConfig(context.Background()); err != nil {
		b.log.Error("Failed to initialize bet config", zap.Error(err))
		b.log.Error("Continuing without bet config - this is expected for new installations")
	}
	b.generateQucickHustleCards()
	return nil
}

func (b *bet) generateQucickHustleCards() {

	// Populate the map
	for _, card := range dto.QuickHustelCards {
		b.quickHustleCardsMap[card.Name] = card.Weight
	}
	// Initialize the multipliers for higher and lower guesses
	for _, card := range dto.QuickHustelHigherMultiplier {
		b.quickHustelHigherMultiplier[card.Name] = card.Multiplier
	}
	for _, card := range dto.QuickHustelLowerMultiplier {
		b.quickHustelLowerMultiplier[card.Name] = card.Multiplier
	}
}

func (b *bet) InitConfig(context.Context) error {
	// check plinko game min bet and max bet config are exist
	//minbet
	_, exist, err := b.ConfigStorage.GetConfigByName(context.Background(), constant.PLINKO_MAX_BET)
	if err != nil {
		return err
	}

	if !exist {
		//create new max bet
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.PLINKO_MAX_BET, Value: constant.PLINKO_DEFAULT_MAX_BET})
		if err != nil {
			return err
		}

	}

	//min bet
	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.PLINKO_MIN_BET)
	if err != nil {
		return err
	}

	if !exist {
		// create new min bet
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.PLINKO_MIN_BET, Value: constant.PLINKO_DEFAULT_MIN_BET})
		if err != nil {
			return err
		}
	}

	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.PLINKO_RTP)
	if err != nil {
		return err
	}

	if !exist {
		// create default rpt
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.PLINKO_RTP, Value: constant.PLINKO_DEFUALT_RTP})
		if err != nil {
			return err
		}
	}

	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.PLINKO_MULTIPLIERS)
	if err != nil {
		return err
	}

	if !exist {
		// create default multiplier
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.PLINKO_MULTIPLIERS, Value: constant.PLINKO_DEFUALT_MULTIPLIERS})
		if err != nil {
			return err
		}
	}
	//check for football card multiplier
	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER)
	if err != nil {
		return err
	}

	if !exist {
		// create default multiplier
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER, Value: constant.CONFIG_FOOTBALL_MATCH_MULTIPLIER_DEFAULT})
		if err != nil {
			return err
		}
	}

	// check for football round price
	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.CONFIG_FOOTBALL_MATCH_CARD_PRICE)
	if err != nil {
		return err
	}

	if !exist {
		// create default multiplier
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.CONFIG_FOOTBALL_MATCH_CARD_PRICE, Value: constant.CONFIG_FOOTBALL_DEFAULT_MATCH_CARD_PRICE})
		if err != nil {
			return err
		}
	}

	// check game are saved or not
	//tucanbit
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_TUCAKBIT)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_TUCAKBIT,
			Name: constant.GAME_TUCAKBIT_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//plinko
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_PLINKO)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_PLINKO,
			Name: constant.GAME_PLINKO_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_CRYPTO_KINGS
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_CRYPTO_KINGS)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_CRYPTO_KINGS,
			Name: constant.GAME_CRYPTO_KINGS_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_FOOTBALL_FIXTURES
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_FOOTBALL_FIXTURES)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_FOOTBALL_FIXTURES,
			Name: constant.GAME_FOOTBALL_FIXTURES_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_QUICK_HUSTLE
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_QUICK_HUSTLE)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_QUICK_HUSTLE,
			Name: constant.GAME_QUICK_HUSTLE_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_ROLL_DA_DICE
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_ROLL_DA_DICE)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_ROLL_DA_DICE,
			Name: constant.GAME_ROLL_DA_DICE_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_SCRATCH_CARD
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_SCRATCH_CARD)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_SCRATCH_CARD,
			Name: constant.GAME_SCRATCH_CARD_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_SPINNING_WHEEL
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_SPINNING_WHEEL)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_SPINNING_WHEEL,
			Name: constant.GAME_SPINNING_WHEEL_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//GAME_STREET_KINGS
	_, err = b.betStorage.GetGameByID(context.Background(), constant.GAME_STREET_KINGS)
	if err != nil {
		if _, err := b.betStorage.CreateGame(context.Background(), dto.Game{
			ID:   constant.GAME_STREET_KINGS,
			Name: constant.GAME_STREET_KINGS_DEFAULT_NAME,
		}); err != nil {
			return err
		}
	}

	//SIGNUP_BONUS
	_, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SIGNUP_BONUS)
	if err != nil {
		return err
	}

	if !exist {
		// create default multiplier
		_, err := b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SIGNUP_BONUS, Value: "0"})
		if err != nil {
			return err
		}
	}

	//scratch game cards price
	return b.InitScratchGameCards()
}

func (b *bet) InitScratchGameCards() error {
	// car
	var (
		resp  dto.Config
		exist bool
		err   error
	)
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_CAR)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_CAR, Value: fmt.Sprintf("%d", constant.SCRATCH_CAR_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// dollar
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_DOLLAR)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_DOLLAR, Value: fmt.Sprintf("%d", constant.SCRATCH_DOLLAR_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// crawn
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_CRAWN)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_CRAWN, Value: fmt.Sprintf("%d", constant.SCRATCH_CRAWN_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// cent
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_CENT)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_CENT, Value: fmt.Sprintf("%d", constant.SCRATCH_CENT_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// DIAMOND
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_DIAMOND)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_DIAMOND, Value: fmt.Sprintf("%d", constant.SCRATCH_DIAMOND_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// CUP
	resp, exist, err = b.ConfigStorage.GetConfigByName(context.Background(), constant.SCRATCH_CUP)
	if err != nil {
		return err
	}

	if !exist {
		// create car
		resp, err = b.ConfigStorage.CreateConfig(context.Background(), dto.Config{Name: constant.SCRATCH_CUP, Value: fmt.Sprintf("%d", constant.SCRATCH_CUP_DEFAULT)})
		if err != nil {
			return err
		}
	}

	// add to the map
	b.scratchCardsPriceHolder[resp.Name] = decimal.RequireFromString(resp.Value)

	// top price
	for _, v := range b.scratchCardsPriceHolder {
		if v.GreaterThan(b.scratchMaxPrice) {
			b.scratchMaxPrice = v
		}
	}

	return nil
}

func (b *bet) AddConnection(ctx context.Context, connReq dto.BroadCastPayload) {

	// Check if the user already has a connection
	if _, exists := b.userConn[connReq.UserID]; exists {
		for existingConn := range b.userConn[connReq.UserID] {
			if existingConn != nil {
				err := existingConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
				if err == nil {
					continue
				}
				b.RemoveConnection(connReq.UserID, existingConn)
			}
		}
	}

	// Initialize the map for the user if it doesn't exist
	if _, exists := b.userConn[connReq.UserID]; !exists {
		b.userConn[connReq.UserID] = make(map[*websocket.Conn]bool)
	}

	// Add the new connection
	b.userConn[connReq.UserID][connReq.Conn] = true
	b.log.Info("New connection added for the user", zap.String("userID", connReq.UserID.String()))
}

func (b *bet) AddToSingleGameConnections(ctx context.Context, connReq dto.BroadCastPayload) {

	// Check if the user already has a connection
	if _, exists := b.userSingleGameConnection[connReq.UserID]; exists {
		for existingConn := range b.userConn[connReq.UserID] {
			if existingConn != nil {
				err := existingConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
				if err == nil {
					continue
				}
				b.RemoveConnection(connReq.UserID, existingConn)
			}
		}
	}

	// Initialize the map for the user if it doesn't exist
	if _, exists := b.userSingleGameConnection[connReq.UserID]; !exists {
		b.userSingleGameConnection[connReq.UserID] = make(map[*websocket.Conn]bool)
	}

	// Add the new connection
	b.userSingleGameConnection[connReq.UserID][connReq.Conn] = true
	b.log.Info("New connection added for the user", zap.String("userID", connReq.UserID.String()))
}

func (b *bet) getPlayerLevelSocketLocker(conn *websocket.Conn) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if locker, exists := b.playerLevelSocketLocker[conn]; exists {
		return locker
	}

	locker := &sync.Mutex{}
	b.playerLevelSocketLocker[conn] = locker
	return locker
}

func (b *bet) getPlayerProgressBarSocketLocker(conn *websocket.Conn) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if locker, exists := b.playerProgressSocketLocker[conn]; exists {
		return locker
	}

	locker := &sync.Mutex{}
	b.playerProgressSocketLocker[conn] = locker
	return locker
}

func (b *bet) getSquadsProgressBarSocketLocker(conn *websocket.Conn) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if locker, exists := b.SquadsProgressSocketLocker[conn]; exists {
		return locker
	}

	locker := &sync.Mutex{}
	b.SquadsProgressSocketLocker[conn] = locker
	return locker
}
func (b *bet) AddPlayerProgressBarConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn) {
	if _, exists := b.playerProgressSocket[userID]; !exists {
		b.playerProgressSocket[userID] = make(map[*websocket.Conn]bool)
	}
	b.playerProgressSocket[userID][conn] = true
	if _, exists := b.playerProgressSocketLocker[conn]; !exists {
		b.playerProgressSocketLocker[conn] = &sync.Mutex{}
	}

	b.getPlayerProgressBarSocketLocker(conn).Lock()
	defer b.getPlayerProgressBarSocketLocker(conn).Unlock()

	b.log.Info("New player level connection added", zap.String("userID", userID.String()))

	// Send ping
	if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
		b.log.Error("Failed to send ping", zap.Error(err))
		return
	}

	// Send connection message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to player progress bar socket")); err != nil {
		b.log.Error("Failed to send connection message", zap.Error(err))
		return
	}

	// Get and send player level
	playerLevel, err := b.GetUserLevel(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	p := dto.GetUserLevelResp3{
		ID:                      playerLevel.ID,
		Level:                   playerLevel.Level,
		NextLevel:               playerLevel.NextLevel,
		AmountSpentToReachLevel: playerLevel.AmountSpentToReachLevel,
		NextLevelRequirement:    playerLevel.NextLevelRequirement,
	}

	msg, err := json.Marshal(p)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		b.log.Error("Failed to send player level message", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Player level socket connection established", zap.String("userID", userID.String()))
}

func (b *bet) AddSquadsProgressBarConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn) {
	if _, exists := b.SquadsProgressSocket[userID]; !exists {
		b.SquadsProgressSocket[userID] = make(map[*websocket.Conn]bool)
	}
	b.SquadsProgressSocket[userID][conn] = true
	if _, exists := b.SquadsProgressSocketLocker[conn]; !exists {
		b.SquadsProgressSocketLocker[conn] = &sync.Mutex{}
	}

	b.getSquadsProgressBarSocketLocker(conn).Lock()
	defer b.getSquadsProgressBarSocketLocker(conn).Unlock()

	b.log.Info("New player level connection added", zap.String("userID", userID.String()))

	// Send ping
	if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
		b.log.Error("Failed to send ping", zap.Error(err))
		return
	}

	// Send connection message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to squads progress bar socket")); err != nil {
		b.log.Error("Failed to send connection message", zap.Error(err))
		return
	}

	// Get and send player level
	b.InitiateTriggerSquadsProgressBar(ctx, userID)

}

func (b *bet) AddPlayerLevelSocketConnection(ctx context.Context, userID uuid.UUID, conn *websocket.Conn) {
	if _, exists := b.playerLevelSocket[userID]; !exists {
		b.playerLevelSocket[userID] = make(map[*websocket.Conn]bool)
	}
	b.playerLevelSocket[userID][conn] = true
	if _, exists := b.playerLevelSocketLocker[conn]; !exists {
		b.playerLevelSocketLocker[conn] = &sync.Mutex{}
	}

	b.getPlayerLevelSocketLocker(conn).Lock()
	defer b.getPlayerLevelSocketLocker(conn).Unlock()

	b.log.Info("New player level connection added", zap.String("userID", userID.String()))

	// Send ping
	if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
		b.log.Error("Failed to send ping", zap.Error(err))
		return
	}

	// Send connection message
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to player level socket")); err != nil {
		b.log.Error("Failed to send connection message", zap.Error(err))
		return
	}

	// Get and send player level
	playerLevel, err := b.GetUserLevel(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	p := dto.GetUserLevelResp{
		ID:    playerLevel.ID,
		Level: playerLevel.Level,
		Bucks: playerLevel.Bucks,
	}

	msg, err := json.Marshal(p)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		b.log.Error("Failed to send player level message", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Player level socket connection established", zap.String("userID", userID.String()))
}

func (b *bet) TriggerLevelResponse(ctx context.Context, userID uuid.UUID) {

	playerLevel, err := b.GetUserLevel(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	pl := dto.GetUserLevelResp{
		ID:           playerLevel.ID,
		Level:        playerLevel.Level,
		Bucks:        playerLevel.Bucks,
		IsFinalLevel: playerLevel.IsFinalLevel,
	}

	msg, err := json.Marshal(pl)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	b.log.Info("Triggering player level response for user", zap.Any("userID", userID), zap.Any("sockets", b.playerLevelSocket[userID]))

	if conns, exists := b.playerLevelSocket[userID]; exists {

		for conn := range conns {
			b.getPlayerLevelSocketLocker(conn).Lock()
			if conn == nil {
				if b.getPlayerLevelSocketLocker(conn) != nil {
					b.getPlayerLevelSocketLocker(conn).Unlock()
				}
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send player level response", zap.Error(err), zap.String("userID", userID.String()))
				if b.getPlayerLevelSocketLocker(conn) != nil {
					b.getPlayerLevelSocketLocker(conn).Unlock()
				}
				b.RemoveConnection(userID, conn)
				continue
			}
			if b.getPlayerLevelSocketLocker(conn) != nil {
				b.getPlayerLevelSocketLocker(conn).Unlock()
			}

			log.Printf("Player level response sent to user %s: %s", userID.String(), msg)
		}
	} else {
		b.log.Info("No player level socket connections found for user", zap.String("userID", userID.String()))
	}
}

func (b *bet) TriggerPlayerProgressBar(ctx context.Context, userID uuid.UUID) {

	playerLevel, err := b.GetUserLevel(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	pl := dto.GetUserLevelResp3{
		ID:                      playerLevel.ID,
		Level:                   playerLevel.Level,
		NextLevel:               playerLevel.NextLevel,
		AmountSpentToReachLevel: playerLevel.AmountSpentToReachLevel,
		NextLevelRequirement:    playerLevel.NextLevelRequirement,
		IsFinalLevel:            playerLevel.IsFinalLevel,
	}

	msg, err := json.Marshal(pl)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	log.Println("Triggering player level response for user", b.playerProgressSocket[userID])

	if conns, exists := b.playerProgressSocket[userID]; exists {

		for conn := range conns {
			b.getPlayerProgressBarSocketLocker(conn).Lock()
			if conn == nil {
				if b.getPlayerProgressBarSocketLocker(conn) != nil {
					b.getPlayerProgressBarSocketLocker(conn).Unlock()
				}
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send player level response", zap.Error(err), zap.String("userID", userID.String()))
				if b.getPlayerProgressBarSocketLocker(conn) != nil {
					b.getPlayerProgressBarSocketLocker(conn).Unlock()
				}
				b.RemoveConnection(userID, conn)
				continue
			}
			if b.getPlayerProgressBarSocketLocker(conn) != nil {
				b.getPlayerProgressBarSocketLocker(conn).Unlock()
			}

			log.Printf("Player level response sent to user %s: %s", userID.String(), msg)
		}
	} else {
		b.log.Info("No player level socket connections found for user", zap.String("userID", userID.String()))
	}
}

func (b *bet) AddToBroadcastConnection(ctx context.Context, conn *websocket.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.broadcastConn, conn)
	b.broadcastConn[conn] = true
}

func (b *bet) BroadcastMessage(ctx context.Context, message []byte, conns map[uuid.UUID]map[*websocket.Conn]bool) {
	var wg sync.WaitGroup

	for userID, userConns := range conns {
		wg.Add(1)

		// Copy connections to avoid modification during iteration
		userConnsCopy := make(map[*websocket.Conn]bool)

		for conn := range userConns {
			userConnsCopy[conn] = true
		}

		go func(userID uuid.UUID, conns map[*websocket.Conn]bool) {
			defer wg.Done()

			for conn := range conns {
				if conn == nil {
					b.RemoveConnection(userID, conn)
					continue
				}

				locker := b.getSocketLocker(conn)
				locker.Lock()
				// Send PingMessage
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					b.log.Warn("Failed to send PingMessage", zap.Error(err), zap.String("userID", userID.String()))
					locker.Unlock()
					b.RemoveConnection(userID, conn)
					continue
				}

				// Send TextMessage
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					b.log.Warn("Failed to send TextMessage", zap.Error(err), zap.String("userID", userID.String()))
					locker.Unlock()
					b.RemoveConnection(userID, conn)
				}
				locker.Unlock()
			}
		}(userID, userConnsCopy)
	}

	wg.Wait()
}

func (b *bet) BroadcastToAllMessage(ctx context.Context, message []byte, conns map[*websocket.Conn]bool) {
	var wg sync.WaitGroup

	for conn, _ := range conns {
		locker := b.getSocketLocker(conn)
		locker.Lock()
		wg.Add(1)
		go func(conn2 *websocket.Conn) {
			defer wg.Done()
			if conn2 == nil {
				b.RemoveBroadcastConn(conn2)
				return
			}
			// Send PingMessage
			if err := conn2.WriteMessage(websocket.PingMessage, nil); err != nil {
				b.log.Warn("Failed to send PingMessage")
				b.RemoveBroadcastConn(conn2)
				return
			}

			// Send TextMessage
			if err := conn2.WriteMessage(websocket.TextMessage, message); err != nil {
				b.log.Warn("Failed to send TextMessage")
				b.RemoveBroadcastConn(conn)
			}

		}(conn)
		locker.Unlock()
	}

	wg.Wait()
}

func (b *bet) RemoveBroadcastConn(conn *websocket.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if conn != nil {
		conn.Close()
	}
	delete(b.broadcastConn, conn)
}

func (b *bet) RemoveConnection(userID uuid.UUID, conn *websocket.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.userConn[userID]; ok {
		if _, exists := b.userConn[userID][conn]; exists {
			delete(b.userConn[userID], conn)
			if conn != nil {
				conn.Close()
			}
			b.log.Info("Connection removed", zap.Any("userID", userID), zap.Any("conn", conn))
		}
		if len(b.userConn[userID]) == 0 {
			delete(b.userConn, userID)
		}
	}
}
func (b *bet) PlaceBet(ctx context.Context, placeBetReq dto.PlaceBetReq) (dto.PlaceBetRes, error) {

	// check users balance
	// lock user before making transactions
	userLock := b.getUserLock(placeBetReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	//validate age

	// validate user input
	if placeBetReq.Currency != "" && placeBetReq.Currency != constant.POINT_CURRENCY && !dto.IsValidCurrency(placeBetReq.Currency) {
		err := fmt.Errorf("invalid currency %s", placeBetReq.Currency)
		b.log.Warn(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceBetRes{}, err
	}

	if placeBetReq.Currency == "" {
		placeBetReq.Currency = constant.POINT_CURRENCY
	}

	//check for the users is blocked or not
	if err := b.CheckGameBlocks(ctx, placeBetReq.UserID); err != nil {
		return dto.PlaceBetRes{}, err
	}

	if placeBetReq.Amount.LessThan(decimal.NewFromInt(1)) || placeBetReq.Amount.GreaterThan(decimal.NewFromInt(1000)) {
		err := fmt.Errorf("minimum bet is $1 and maximum bet amount is $1000")
		b.log.Warn(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceBetRes{}, err
	}

	// check for open bet for the given bet round
	betRound, exist, err := b.betStorage.GetRoundByID(ctx, placeBetReq.RoundID)
	if err != nil {
		return dto.PlaceBetRes{}, err
	}
	if !exist || betRound.Status != constant.BET_OPEN {
		err = fmt.Errorf("no open bet round found with this round_id")
		b.log.Warn(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceBetRes{}, err
	}

	userBalance, exist, err := b.balanceStorage.GetUserBalanaceByUserID(ctx, dto.Balance{
		UserId:       placeBetReq.UserID,
		CurrencyCode: placeBetReq.Currency,
	})
	if err != nil {
		return dto.PlaceBetRes{}, err
	}
	if !exist || userBalance.RealMoney.LessThan(placeBetReq.Amount) {
		err = fmt.Errorf("insufficient balance with %s currency", placeBetReq.Currency)
		b.log.Warn(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceBetRes{}, err
	}

	//perform internal transaction
	// check user bets if exist or not
	userBets, exist, _ := b.betStorage.GetUserBetByUserIDAndRoundID(ctx, dto.Bet{
		RoundID: placeBetReq.RoundID,
		UserID:  placeBetReq.UserID,
		Status:  constant.ACTIVE,
	})

	if exist && len(userBets) > 0 {
		err = fmt.Errorf("bet already exist")
		b.log.Error(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.PlaceBetRes{}, err
	}

	newBalance := userBalance.RealMoney.Sub(placeBetReq.Amount)
	transactionID := utils.GenerateTransactionId()
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    placeBetReq.UserID,
		Currency:  placeBetReq.Currency,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.PlaceBetRes{}, err
	}
	// save transaction
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_OPERATIONAL_TYPE)
	if err != nil {
		b.log.Error(err.Error(), zap.Any("placeBetReq", placeBetReq))
		err = errors.ErrInternalServerError.Wrap(err, err.Error())
		b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    placeBetReq.UserID,
			Currency:  placeBetReq.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.RealMoney,
		})
		return dto.PlaceBetRes{}, err
	}

	// save transaction logs
	// save operations logs
	balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             placeBetReq.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           placeBetReq.Currency,
		Description:        fmt.Sprintf("place bet amount %v, new balance is %v and  currency %s", placeBetReq.Amount, userBalance.RealMoney.Sub(placeBetReq.Amount), placeBetReq.Currency),
		ChangeAmount:       placeBetReq.Amount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		return dto.PlaceBetRes{}, err
	}
	//placeBet
	savedBet, err := b.betStorage.SaveUserBet(ctx, dto.Bet{
		UserID:              placeBetReq.UserID,
		RoundID:             placeBetReq.RoundID,
		Amount:              placeBetReq.Amount,
		Currency:            placeBetReq.Currency,
		ClientTransactionID: balanceLogs.ID.String(),
	})
	if err != nil {
		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    placeBetReq.UserID,
			Currency:  placeBetReq.Currency,
			Component: constant.REAL_MONEY,
			Amount:    userBalance.RealMoney,
		})
		if err2 != nil {
			return dto.PlaceBetRes{}, err2
		}
		return dto.PlaceBetRes{}, err
	}

	return dto.PlaceBetRes{
		Status:  constant.SUCCESS,
		Message: constant.BET_SUCCESS_MESSAGE,
		Date: dto.PlaceBetReqData{
			BetID:    savedBet.BetID,
			RoundID:  savedBet.RoundID,
			Amount:   savedBet.Amount,
			Currency: savedBet.Currency,
		},
	}, nil
}

func (b *bet) CancelBet(ctx context.Context, cancelReq dto.CancelBetReq) (dto.CancelBetResp, error) {
	var newBalance decimal.Decimal
	var currentBalance decimal.Decimal
	var userBet dto.Bet
	userLock := b.getUserLock(cancelReq.UserID)
	userLock.Lock()
	defer userLock.Unlock()
	// GET user bet of requested bet
	userBets, exist, err := b.betStorage.GetUserBetByUserIDAndRoundID(ctx, dto.Bet{
		RoundID: cancelReq.RoundID,
		UserID:  cancelReq.UserID,
		Status:  constant.ACTIVE,
	})

	if err != nil {
		return dto.CancelBetResp{}, err
	}
	if !exist || len(userBets) < 1 {
		err := fmt.Errorf("user dose not have bet with this round id %s %s %s", cancelReq.UserID.String(), "roundID", cancelReq.RoundID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CancelBetResp{}, err
	}

	userBet = userBets[0]

	// check for the bet status
	betRound, exist, err := b.betStorage.GetRoundByID(ctx, cancelReq.RoundID)
	if err != nil {
		return dto.CancelBetResp{}, err
	}

	if !exist || betRound.Status != constant.BET_OPEN {
		err := fmt.Errorf("no open round found with this userID %s  and roundID %s ", cancelReq.UserID.String(), cancelReq.RoundID.String())
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.CancelBetResp{}, err
	}

	// get user balance
	userBalances, err := b.balanceStorage.GetBalancesByUserID(ctx, cancelReq.UserID)
	if err != nil {
		return dto.CancelBetResp{}, err
	}

	for _, userBalance := range userBalances {
		if userBalance.CurrencyCode == userBet.Currency {
			currentBalance = userBalance.RealMoney
			newBalance = userBalance.RealMoney.Add(userBet.Amount)
		}
	}
	_, err = b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
		UserID:    cancelReq.UserID,
		Currency:  userBet.Currency,
		Component: constant.REAL_MONEY,
		Amount:    newBalance,
	})
	if err != nil {
		return dto.CancelBetResp{}, err
	}
	operationalGroupAndTypeIDs, err := b.CreateOrGetOperationalGroupAndType(ctx, constant.TRANSFER, constant.BET_CANCELED)
	if err != nil {
		return dto.CancelBetResp{}, err
	}
	transactionID := utils.GenerateTransactionId()
	balanceLogs, err := b.balanceLogStorage.SaveBalanceLogs(ctx, dto.BalanceLogs{
		UserID:             cancelReq.UserID,
		Component:          constant.REAL_MONEY,
		Currency:           userBet.Currency,
		Description:        fmt.Sprintf("cancel bet amount %v, new balance is %v and currency  %s", userBet.Amount, newBalance, userBet.Currency),
		ChangeAmount:       userBet.Amount,
		OperationalGroupID: operationalGroupAndTypeIDs.OperationalGroupID,
		OperationalTypeID:  operationalGroupAndTypeIDs.OperationalTypeID,
		BalanceAfterUpdate: &newBalance,
		TransactionID:      &transactionID,
	})
	if err != nil {
		// revert balance to the previous
		_, err2 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    cancelReq.UserID,
			Currency:  userBet.Currency,
			Component: constant.REAL_MONEY,
			Amount:    currentBalance,
		})
		if err2 != nil {
			return dto.CancelBetResp{}, err2
		}
		return dto.CancelBetResp{}, err
	}
	//cancel bet
	_, err = b.betStorage.UpdateBetStatus(ctx, userBet.BetID, constant.CANCELED)
	if err != nil {
		// revert the balance and balance log
		err2 := b.balanceLogStorage.DeleteBalanceLog(ctx, balanceLogs.ID)
		if err2 != nil {
			return dto.CancelBetResp{}, err2
		}
		//revert balance
		_, err3 := b.balanceStorage.UpdateMoney(ctx, dto.UpdateBalanceReq{
			UserID:    cancelReq.UserID,
			Currency:  userBet.Currency,
			Component: constant.REAL_MONEY,
			Amount:    currentBalance,
		})
		if err3 != nil {
			return dto.CancelBetResp{}, err3
		}
	}

	return dto.CancelBetResp{
		Message: constant.CANCEL_SUCCESS,
		UserID:  cancelReq.UserID,
		RoundID: cancelReq.RoundID,
	}, err
}

func (b *bet) getUserLock(userID uuid.UUID) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.locker[userID]; !exists {
		b.locker[userID] = &sync.Mutex{}
	}
	return b.locker[userID]
}

func (b *bet) GetBetHistory(ctx context.Context, betReq dto.GetBetHistoryReq) (dto.BetHistoryResp, error) {
	if betReq.PerPage <= 0 {
		betReq.PerPage = 10
	}
	if betReq.Page <= 0 {
		betReq.Page = 1
	}
	offset := (betReq.Page - 1) * betReq.PerPage
	betReq.Offset = offset
	betHistory, exist, err := b.betStorage.GetBetHistory(ctx, betReq)
	if err != nil {
		return dto.BetHistoryResp{}, err
	}
	if !exist {
		err := fmt.Errorf("no bet history available")
		b.log.Warn(err.Error(), zap.Any("getBetReq", betReq))
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return dto.BetHistoryResp{}, err
	}
	return betHistory, nil
}

func (b *bet) getSocketLocker(conn *websocket.Conn) *sync.Mutex {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.socketLocker[conn]; !exists {
		b.socketLocker[conn] = &sync.Mutex{}
	}
	return b.socketLocker[conn]
}

func (b *bet) CheckGameBlocks(ctx context.Context, userID uuid.UUID) error {
	acBlocks, exist, err := b.userStorage.GetBlockedAccountByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if exist {
		// check if account blocked for game
		for _, acBlock := range acBlocks {
			if acBlock.Type == constant.BLOCK_TYPE_GAMING || acBlock.Type == constant.BLOCK_TYPE_COMPLETE {
				// check for  duration and type
				if acBlock.Duration == constant.BLOCK_DURATION_PERMANENT {
					err = fmt.Errorf("user can not play game, account permanently blocked from playing game")
					err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
					return err
				} else {
					//check if the user is still blocked
					if acBlock.BlockedTo != nil {
						blockedUntil := *acBlock.BlockedTo
						if blockedUntil.After(time.Now().In(time.Now().Location()).UTC()) {
							err = fmt.Errorf("user can not play game account temporary blocked for %2f hours ", blockedUntil.Sub(time.Now().In(time.Now().Location()).UTC()).Hours())
							err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (b *bet) StreamToSingleConnection(ctx context.Context, conns map[*websocket.Conn]bool, byteDate []byte, userID uuid.UUID) {
	cons := conns
	if _, ok := b.singleBetSocketSync[userID]; !ok {
		b.singleBetSocketSync[userID] = &sync.Mutex{}
	}
	b.singleBetSocketSync[userID].Lock()
	defer b.singleBetSocketSync[userID].Unlock()

	for conn := range cons {
		if conn != nil {
			conn.WriteMessage(websocket.TextMessage, byteDate)
		}
	}

}

func (b *bet) CheckBetLockStatus(ctx context.Context, ID uuid.UUID) (bool, error) {
	resp, err := b.betStorage.GetGameByID(ctx, ID)
	if err != nil {
		return false, err
	}

	if resp.Status != constant.ACTIVE {
		return false, nil
	}
	return true, nil
}

func (b *bet) InitiateTriggerSquadsProgressBar(ctx context.Context, userID uuid.UUID) {
	userSquads, err := b.betStorage.GetUserSquads(ctx, userID)
	if err != nil {
		b.log.Error("Failed to get user squads", zap.Error(err), zap.String("userID", userID.String()))
		return
	}
	if len(userSquads) == 0 {
		b.log.Info("No squads found for user", zap.String("userID", userID.String()))
		return
	}

	for i, squad := range userSquads {

		if i == 0 {
			go b.TriggerSquadsProgressBar(ctx, userID, squad.ID)
		}

		squadMembers, err := b.betStorage.GetAllSquadMembersBySquadId(ctx, squad.ID)
		if err != nil {
			b.log.Error("Failed to get squad members", zap.Error(err), zap.String("squadID", squad.ID.String()))
			continue
		}
		for _, member := range squadMembers {
			go b.TriggerSquadsProgressBar(ctx, member.UserID, squad.ID)
		}
	}

}

func (b *bet) TriggerSquadsProgressBar(ctx context.Context, userID, squadID uuid.UUID) {

	playerLevel, err := b.GetSquadLevel(ctx, squadID)
	if err != nil {
		b.log.Error("Failed to get user level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	pl := dto.GetUserLevelResp3{
		ID:                      playerLevel.ID,
		Level:                   playerLevel.Level,
		NextLevel:               playerLevel.NextLevel,
		AmountSpentToReachLevel: playerLevel.AmountSpentToReachLevel,
		NextLevelRequirement:    playerLevel.NextLevelRequirement,
		IsFinalLevel:            playerLevel.IsFinalLevel,
		SquadID:                 playerLevel.SquadID,
	}

	msg, err := json.Marshal(pl)
	if err != nil {
		b.log.Error("Failed to marshal player level", zap.Error(err), zap.String("userID", userID.String()))
		return
	}

	log.Println("Triggering player level response for user", b.SquadsProgressSocket[userID])

	if conns, exists := b.SquadsProgressSocket[userID]; exists {

		for conn := range conns {
			b.getSquadsProgressBarSocketLocker(conn).Lock()
			if conn == nil {
				if b.getSquadsProgressBarSocketLocker(conn) != nil {
					b.getSquadsProgressBarSocketLocker(conn).Unlock()
				}
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				b.log.Warn("Failed to send player level response", zap.Error(err), zap.String("userID", userID.String()))
				if b.getSquadsProgressBarSocketLocker(conn) != nil {
					b.getSquadsProgressBarSocketLocker(conn).Unlock()
				}
				b.RemoveConnection(userID, conn)
				continue
			}
			if b.getSquadsProgressBarSocketLocker(conn) != nil {
				b.getSquadsProgressBarSocketLocker(conn).Unlock()
			}

			log.Printf("Player level response sent to user %s: %s", userID.String(), msg)
		}
	} else {
		b.log.Info("No player level socket connections found for user", zap.String("userID", userID.String()))
	}
}
