# Frontend WebSocket Integration Guide

## 🎯 TucanBIT Real-Time Notifications Integration

This guide provides complete frontend integration instructions for implementing real-time balance updates, cashback notifications, and winner notifications using WebSocket connections.

## 📋 Table of Contents

1. [Overview](#overview)
2. [WebSocket Connection Setup](#websocket-connection-setup)
3. [Authentication](#authentication)
4. [Message Types & Handling](#message-types--handling)
5. [Complete Implementation](#complete-implementation)
6. [Testing Reference](#testing-reference)
7. [Production Considerations](#production-considerations)
8. [Troubleshooting](#troubleshooting)

## 🎯 Overview

The TucanBIT platform provides real-time notifications through WebSocket connections for:

- **Balance Updates**: Real-time balance changes on every transaction
- **Cashback Notifications**: Automatic cashback earnings and tier information
- **Winner Notifications**: Big win celebrations and jackpot announcements

### Key Features
- ✅ Single WebSocket connection for all notification types
- ✅ JWT-based authentication
- ✅ Automatic reconnection handling
- ✅ Real-time updates with detailed transaction information
- ✅ Mobile-responsive notifications

## 🔌 WebSocket Connection Setup

### Connection Endpoint

**Development:**
```
ws://localhost:8080/ws/balance/player
```

**Production:**
```
wss://your-domain.com/ws/balance/player
```

### Basic Connection Code

```javascript
class TucanBITWebSocket {
    constructor(accessToken, baseUrl = 'ws://localhost:8080') {
        this.accessToken = accessToken;
        this.wsUrl = `${baseUrl}/ws/balance/player`;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 3000;
        this.isManualClose = false;
    }

    connect() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            console.log('Already connected');
            return;
        }

        try {
            this.ws = new WebSocket(this.wsUrl);
            this.setupEventHandlers();
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.handleReconnection();
        }
    }

    setupEventHandlers() {
        this.ws.onopen = (event) => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.sendAuthentication();
            this.onConnectionOpen?.(event);
        };

        this.ws.onmessage = (event) => {
            this.handleMessage(event);
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
            this.onConnectionClose?.(event);
            
            if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
                this.handleReconnection();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.onConnectionError?.(error);
        };
    }

    sendAuthentication() {
        const authMessage = {
            type: 'auth',
            access_token: this.accessToken
        };
        this.ws.send(JSON.stringify(authMessage));
        console.log('Authentication token sent');
    }

    disconnect() {
        this.isManualClose = true;
        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect');
            this.ws = null;
        }
    }

    handleReconnection() {
        this.reconnectAttempts++;
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${this.reconnectDelay/1000}s...`);
        setTimeout(() => {
            this.connect();
        }, this.reconnectDelay);
    }

    // Event handlers (to be implemented by frontend)
    onConnectionOpen = null;
    onConnectionClose = null;
    onConnectionError = null;
    onBalanceUpdate = null;
    onCashbackUpdate = null;
    onWinnerNotification = null;
}
```

## 🔐 Authentication

### JWT Token Requirements

The WebSocket connection requires a valid JWT token with the following claims:

```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "is_verified": false,
  "email_verified": false,
  "phone_verified": false,
  "exp": 1760206847,
  "iat": 1760120447,
  "iss": "tucanbit",
  "sub": "a5e168fb-168e-4183-84c5-d49038ce00b5"
}
```

### Authentication Message Format

```javascript
const authMessage = {
    type: 'auth',
    access_token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
};
```

### Connection Confirmation

After successful authentication, you'll receive:
```
"Connected to user balance socket"
```

## 📨 Message Types & Handling

### 1. Balance Updates

**Message Format:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "balance": "36388.88",
  "balance_formatted": "$36,388.88",
  "currency": "USD"
}
```

**Handler Implementation:**
```javascript
handleBalanceUpdate(data) {
    // Update balance display
    const balanceElement = document.getElementById('balance');
    if (data.balance_formatted) {
        balanceElement.textContent = data.balance_formatted;
    } else {
        balanceElement.textContent = `$${data.balance}`;
    }
    
    // Show balance change animation
    this.showBalanceChangeAnimation(data.balance);
    
    // Trigger custom event
    this.onBalanceUpdate?.(data);
}
```

### 2. Cashback Notifications

**Message Format:**
```json
{
  "type": "cashback_update",
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "data": {
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "current_tier": {
      "id": "d38233b1-920b-462e-bc90-ca805218eaf0",
      "tier_name": "Gold",
      "tier_level": 3,
      "min_expected_ggr_required": "5000",
      "cashback_percentage": "1.5",
      "bonus_multiplier": "1.25",
      "daily_cashback_limit": "250",
      "weekly_cashback_limit": null,
      "monthly_cashback_limit": null,
      "special_benefits": {},
      "is_active": true,
      "created_at": "2025-09-15T11:41:40.757764Z",
      "updated_at": "2025-09-15T11:41:40.757764Z"
    },
    "level_progress": "0",
    "total_ggr": "8260.14",
    "available_cashback": "463.2",
    "pending_cashback": "0",
    "total_claimed": "88.9502",
    "next_tier_ggr": "15000",
    "daily_limit": "250",
    "weekly_limit": null,
    "monthly_limit": null,
    "special_benefits": {},
    "last_game_info": {
      "game_id": "130300002",
      "game_name": "Groovetech Game 130300002",
      "game_type": "groovetech",
      "game_variant": "130300002",
      "house_edge": "0.02",
      "house_edge_percent": "2.00%",
      "cashback_rate": "2",
      "cashback_percent": "2.0%",
      "expected_ggr": "0.6",
      "earned_cashback": "0.6",
      "bet_amount": "30",
      "transaction_id": "winner_combined_tx_001",
      "timestamp": "2025-10-11T13:07:38.34964075Z"
    }
  },
  "message": "Cashback availability updated"
}
```

**Handler Implementation:**
```javascript
handleCashbackUpdate(data) {
    const cashbackData = data.data;
    
    // Update cashback display
    document.getElementById('available-cashback').textContent = 
        `$${cashbackData.available_cashback}`;
    
    // Update tier information
    document.getElementById('current-tier').textContent = 
        cashbackData.current_tier.tier_name;
    
    // Show cashback notification
    this.showCashbackNotification({
        available: cashbackData.available_cashback,
        tier: cashbackData.current_tier.tier_name,
        earned: cashbackData.last_game_info?.earned_cashback || '0',
        game: cashbackData.last_game_info?.game_name || 'Game',
        gameId: cashbackData.last_game_info?.game_id || '',
        houseEdge: cashbackData.last_game_info?.house_edge_percent || '0%'
    });
    
    // Trigger custom event
    this.onCashbackUpdate?.(cashbackData);
}
```

### 3. Winner Notifications

**Message Format:**
```json
{
  "type": "winner_notification",
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "data": {
    "username": "P-uBKtmkyl5LPo",
    "email": "ashenafialemu9898@gmail.com",
    "game_name": "",
    "game_id": "130300002",
    "bet_amount": "30",
    "win_amount": "75",
    "net_winnings": "45",
    "currency": "USD",
    "timestamp": "2025-10-11T13:07:38.354639245Z",
    "session_id": "3818_2ca3c33b-1b76-4e1b-b854-76a41f181336",
    "round_id": "winner_combined_test_001",
    "transaction_id": "winner_combined_tx_001"
  },
  "message": "Congratulations! You won!"
}
```

**Handler Implementation:**
```javascript
handleWinnerNotification(data) {
    const winnerData = data.data;
    
    // Show winner notification popup/toast
    this.showWinnerNotification({
        username: winnerData.username,
        gameName: winnerData.game_name || `Game ${winnerData.game_id}`,
        gameId: winnerData.game_id,
        winAmount: winnerData.win_amount,
        betAmount: winnerData.bet_amount,
        netWinnings: winnerData.net_winnings,
        currency: winnerData.currency,
        timestamp: winnerData.timestamp,
        sessionId: winnerData.session_id,
        roundId: winnerData.round_id,
        transactionId: winnerData.transaction_id
    });
    
    // Play celebration sound
    this.playCelebrationSound();
    
    // Show confetti animation
    this.showConfettiAnimation();
    
    // Trigger custom event
    this.onWinnerNotification?.(winnerData);
}
```

## 🎨 Complete Implementation

### HTML Structure

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TucanBIT Casino - Real-time Notifications</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <div class="casino-interface">
        <!-- Connection Status -->
        <div class="connection-status">
            <div id="ws-status" class="status-indicator">Disconnected</div>
        </div>
        
        <!-- Balance Display -->
        <div class="balance-section">
            <h3>Account Balance</h3>
            <div id="balance" class="balance-display">$0.00</div>
        </div>
        
        <!-- Cashback Display -->
        <div class="cashback-section">
            <h3>Available Cashback</h3>
            <div id="available-cashback" class="cashback-amount">$0.00</div>
            <div id="current-tier" class="tier-info">Bronze</div>
        </div>
        
        <!-- Control Buttons -->
        <div class="controls">
            <button id="connect-btn" onclick="wsClient.connect()">Connect</button>
            <button id="disconnect-btn" onclick="wsClient.disconnect()" disabled>Disconnect</button>
            <button id="clear-btn" onclick="clearMessages()">Clear Messages</button>
        </div>
        
        <!-- Message Log -->
        <div class="message-log">
            <h3>Real-time Messages</h3>
            <div id="messages" class="messages-container"></div>
        </div>
    </div>

    <script src="tucanbit-websocket.js"></script>
    <script src="app.js"></script>
</body>
</html>
```

### CSS Styling

```css
/* styles.css */
body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    margin: 0;
    padding: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
}

.casino-interface {
    max-width: 800px;
    margin: 0 auto;
    background: white;
    border-radius: 12px;
    padding: 30px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.3);
}

.connection-status {
    text-align: center;
    margin-bottom: 20px;
}

.status-indicator {
    display: inline-block;
    padding: 8px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 14px;
}

.status-indicator.connected {
    background: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
}

.status-indicator.disconnected {
    background: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
}

.status-indicator.connecting {
    background: #fff3cd;
    color: #856404;
    border: 1px solid #ffeaa7;
}

.balance-section, .cashback-section {
    text-align: center;
    margin: 20px 0;
    padding: 20px;
    background: #f8f9fa;
    border-radius: 8px;
}

.balance-display {
    font-size: 32px;
    font-weight: bold;
    color: #28a745;
    margin: 10px 0;
}

.cashback-amount {
    font-size: 24px;
    font-weight: bold;
    color: #17a2b8;
    margin: 10px 0;
}

.tier-info {
    font-size: 16px;
    color: #6c757d;
    font-weight: 500;
}

.controls {
    text-align: center;
    margin: 20px 0;
}

.controls button {
    background: #007bff;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 6px;
    cursor: pointer;
    margin: 0 10px;
    font-size: 14px;
    font-weight: 500;
    transition: background-color 0.3s;
}

.controls button:hover {
    background: #0056b3;
}

.controls button:disabled {
    background: #6c757d;
    cursor: not-allowed;
}

.message-log {
    margin-top: 30px;
}

.messages-container {
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid #dee2e6;
    border-radius: 6px;
    padding: 15px;
    background: #f8f9fa;
}

.message {
    margin: 8px 0;
    padding: 10px;
    border-radius: 4px;
    font-family: 'Courier New', monospace;
    font-size: 12px;
    white-space: pre-wrap;
}

.message.success {
    background: #d4edda;
    color: #155724;
    border-left: 4px solid #28a745;
}

.message.info {
    background: #d1ecf1;
    color: #0c5460;
    border-left: 4px solid #17a2b8;
}

.message.error {
    background: #f8d7da;
    color: #721c24;
    border-left: 4px solid #dc3545;
}

/* Notification Styles */
.cashback-toast {
    position: fixed;
    top: 20px;
    right: 20px;
    background: linear-gradient(135deg, #28a745, #20c997);
    color: white;
    padding: 15px 20px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    z-index: 1000;
    animation: slideInRight 0.3s ease-out;
    max-width: 300px;
}

.winner-modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.8);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 2000;
    animation: fadeIn 0.3s ease-out;
}

.modal-content {
    background: white;
    padding: 30px;
    border-radius: 12px;
    text-align: center;
    max-width: 400px;
    animation: scaleIn 0.3s ease-out;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
}

.winner-icon {
    font-size: 48px;
    margin-bottom: 15px;
}

/* Animations */
@keyframes slideInRight {
    from { transform: translateX(100%); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

@keyframes scaleIn {
    from { transform: scale(0.8); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
}

@keyframes pulse {
    0% { transform: scale(1); }
    50% { transform: scale(1.05); }
    100% { transform: scale(1); }
}

.balance-display.updated {
    animation: pulse 0.5s ease-in-out;
}

/* Mobile Responsive */
@media (max-width: 768px) {
    .casino-interface {
        margin: 10px;
        padding: 20px;
    }
    
    .balance-display {
        font-size: 24px;
    }
    
    .cashback-amount {
        font-size: 20px;
    }
    
    .controls button {
        display: block;
        width: 100%;
        margin: 5px 0;
    }
    
    .cashback-toast {
        right: 10px;
        left: 10px;
        max-width: none;
    }
}
```

### JavaScript Implementation

```javascript
// app.js
class TucanBITApp {
    constructor() {
        this.accessToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1IiwiaXNfdmVyaWZpZWQiOmZhbHNlLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInBob25lX3ZlcmlmaWVkIjpmYWxzZSwiZXhwIjoxNzYwMjA2ODQ3LCJpYXQiOjE3NjAxMjA0NDcsImlzcyI6InR1Y2FuYml0Iiwic3ViIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1In0.11AvC497Av56HqLqlVQb0xUkR36Rq21V79aIRMWDEP0';
        this.wsClient = new TucanBITWebSocket(this.accessToken);
        this.setupEventHandlers();
        this.setupUI();
    }

    setupEventHandlers() {
        this.wsClient.onConnectionOpen = (event) => {
            this.updateConnectionStatus('connected');
            this.addMessage('WebSocket connection established', 'success');
            this.enableDisconnectButton();
        };

        this.wsClient.onConnectionClose = (event) => {
            this.updateConnectionStatus('disconnected');
            this.addMessage(`Connection closed. Code: ${event.code}, Reason: ${event.reason}`, 'error');
            this.enableConnectButton();
        };

        this.wsClient.onConnectionError = (error) => {
            this.addMessage(`WebSocket error: ${error}`, 'error');
        };

        this.wsClient.onBalanceUpdate = (data) => {
            this.handleBalanceUpdate(data);
        };

        this.wsClient.onCashbackUpdate = (data) => {
            this.handleCashbackUpdate(data);
        };

        this.wsClient.onWinnerNotification = (data) => {
            this.handleWinnerNotification(data);
        };
    }

    setupUI() {
        // Auto-connect on page load
        this.addMessage('Page loaded. Click Connect to start WebSocket connection.', 'info');
    }

    updateConnectionStatus(status) {
        const statusEl = document.getElementById('ws-status');
        statusEl.textContent = status.charAt(0).toUpperCase() + status.slice(1);
        statusEl.className = `status-indicator ${status}`;
    }

    enableConnectButton() {
        document.getElementById('connect-btn').disabled = false;
        document.getElementById('disconnect-btn').disabled = true;
    }

    enableDisconnectButton() {
        document.getElementById('connect-btn').disabled = true;
        document.getElementById('disconnect-btn').disabled = false;
    }

    addMessage(message, type = 'info') {
        const messagesEl = document.getElementById('messages');
        const messageEl = document.createElement('div');
        messageEl.className = `message ${type}`;
        
        const timestamp = new Date().toLocaleTimeString();
        messageEl.textContent = `[${timestamp}] ${message}`;
        
        messagesEl.appendChild(messageEl);
        messagesEl.scrollTop = messagesEl.scrollHeight;
    }

    handleBalanceUpdate(data) {
        const balanceEl = document.getElementById('balance');
        if (data.balance_formatted) {
            balanceEl.textContent = data.balance_formatted;
        } else {
            balanceEl.textContent = `$${data.balance}`;
        }
        
        // Add pulse animation
        balanceEl.classList.add('updated');
        setTimeout(() => balanceEl.classList.remove('updated'), 500);
        
        this.addMessage(`Balance updated: ${data.balance_formatted || `$${data.balance}`}`, 'success');
    }

    handleCashbackUpdate(data) {
        document.getElementById('available-cashback').textContent = `$${data.available_cashback}`;
        document.getElementById('current-tier').textContent = data.current_tier.tier_name;
        
        this.showCashbackNotification({
            available: data.available_cashback,
            tier: data.current_tier.tier_name,
            earned: data.last_game_info?.earned_cashback || '0',
            game: data.last_game_info?.game_name || 'Game'
        });
        
        this.addMessage(`🎁 CASHBACK UPDATE: Available: $${data.available_cashback}, Tier: ${data.current_tier.tier_name}`, 'success');
    }

    handleWinnerNotification(data) {
        this.showWinnerNotification({
            username: data.username,
            gameName: data.game_name || `Game ${data.game_id}`,
            winAmount: data.win_amount,
            betAmount: data.bet_amount,
            netWinnings: data.net_winnings,
            currency: data.currency
        });
        
        this.addMessage(`🏆 WINNER NOTIFICATION: ${data.username} won $${data.win_amount} on ${data.game_name || `Game ${data.game_id}`}!`, 'success');
    }

    showCashbackNotification(data) {
        const toast = document.createElement('div');
        toast.className = 'cashback-toast';
        toast.innerHTML = `
            <div class="toast-content">
                <div class="toast-icon">🎁</div>
                <div class="toast-text">
                    <strong>Cashback Earned!</strong><br>
                    +$${data.earned} from ${data.game}<br>
                    Total Available: $${data.available}<br>
                    Tier: ${data.tier}
                </div>
            </div>
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 5000);
    }

    showWinnerNotification(data) {
        const modal = document.createElement('div');
        modal.className = 'winner-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="winner-icon">🏆</div>
                <h2>Congratulations!</h2>
                <p><strong>${data.username}</strong> won <strong>$${data.winAmount}</strong>!</p>
                <p>Game: ${data.gameName}</p>
                <p>Bet: $${data.betAmount} → Win: $${data.winAmount}</p>
                <p>Net Profit: <strong>$${data.netWinnings}</strong></p>
                <button onclick="this.parentElement.parentElement.remove()">Close</button>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        setTimeout(() => {
            modal.remove();
        }, 10000);
    }

    playCelebrationSound() {
        // Implement celebration sound
        console.log('🎵 Playing celebration sound');
    }

    showConfettiAnimation() {
        // Implement confetti animation
        console.log('🎊 Showing confetti animation');
    }
}

// Global functions
function clearMessages() {
    document.getElementById('messages').innerHTML = '';
}

// Initialize app
let app;
window.onload = function() {
    app = new TucanBITApp();
};
```

## 🧪 Testing Reference

### Test File: `websocket-test.html`

The complete testing implementation is available in `websocket-test.html` which demonstrates:

1. **WebSocket Connection**: Automatic connection and authentication
2. **Message Handling**: All three message types properly handled
3. **UI Updates**: Real-time balance, cashback, and winner displays
4. **Error Handling**: Connection errors and reconnection logic
5. **Visual Feedback**: Status indicators and message logging

### Testing Endpoints

**Game Launch:**
```bash
curl -X POST "http://localhost:8080/api/groove/launch-game" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "game_id": "130300002",
    "device_type": "desktop",
    "game_mode": "real",
    "country": "MX",
    "currency": "USD",
    "language": "en_US"
  }'
```

**Combined Wager & Result (Triggers Winner Notifications):**
```bash
curl -X GET "http://localhost:8080/groove-official?request=wagerAndResult&accountid=USER_ID&gamesessionid=SESSION_ID&device=desktop&gameid=130300002&apiversion=1.2&betamount=30.0&result=75.0&roundid=test_round&transactionid=test_tx&gamestatus=completed"
```

**Separate Wager & Result (No Winner Notifications):**
```bash
# Wager
curl -X GET "http://localhost:8080/groove-official?request=wager&accountid=USER_ID&gamesessionid=SESSION_ID&device=desktop&gameid=130300002&apiversion=1.2&betamount=25.0&roundid=test_round&transactionid=test_tx"

# Result
curl -X GET "http://localhost:8080/groove-official?request=result&accountid=USER_ID&gamesessionid=SESSION_ID&device=desktop&gameid=130300002&apiversion=1.2&result=15.0&roundid=test_round&transactionid=test_tx&gamestatus=completed"
```

## 🚀 Production Considerations

### Security
- **Use WSS**: Always use `wss://` in production
- **Token Validation**: Implement proper JWT token validation
- **Rate Limiting**: Prevent notification spam
- **CORS Configuration**: Proper CORS headers for WebSocket

### Performance
- **Connection Pooling**: Manage multiple WebSocket connections
- **Message Queuing**: Handle high-frequency updates
- **Memory Management**: Clean up old notifications
- **Bandwidth Optimization**: Compress large messages

### Reliability
- **Reconnection Logic**: Exponential backoff for reconnections
- **Heartbeat/Ping**: Keep connections alive
- **Error Recovery**: Graceful handling of connection failures
- **Fallback Mechanisms**: Alternative notification methods

### Monitoring
- **Connection Metrics**: Track connection success/failure rates
- **Message Delivery**: Monitor notification delivery
- **Performance Metrics**: Track response times
- **Error Logging**: Comprehensive error tracking

## 🔧 Troubleshooting

### Common Issues

**1. Connection Failed**
- Check WebSocket URL format
- Verify JWT token validity
- Check network connectivity
- Review CORS configuration

**2. Authentication Failed**
- Verify JWT token format
- Check token expiration
- Ensure proper message format
- Review server logs

**3. No Notifications Received**
- Verify WebSocket connection status
- Check message handler implementation
- Review transaction endpoints used
- Check server-side notification triggers

**4. Winner Notifications Not Working**
- Use `wagerAndResult` endpoint instead of separate `wager` + `result`
- Verify win amount is greater than bet amount
- Check WebSocket connection is active
- Review winner notification thresholds

### Debug Tools

**Browser Developer Tools:**
```javascript
// Enable WebSocket debugging
localStorage.debug = 'websocket';

// Log all WebSocket messages
ws.onmessage = function(event) {
    console.log('WebSocket Message:', event.data);
    // Your existing handler code
};
```

**Server Logs:**
```bash
# Monitor WebSocket connections
docker logs tucanbit-app --tail 100 | grep -i websocket

# Monitor notification triggers
docker logs tucanbit-app --tail 100 | grep -i "triggering\|notification"
```

## 📞 Support

For technical support or questions about WebSocket integration:

- **Documentation**: This guide and `websocket-test.html`
- **Testing**: Use the provided test file as reference
- **Logs**: Check server logs for debugging information
- **API Reference**: TucanBIT API documentation

---

**Last Updated**: October 11, 2025  
**Version**: 1.0  
**Tested With**: TucanBIT WebSocket API v1.2
