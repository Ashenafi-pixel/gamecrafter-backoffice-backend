# Frontend Winner Notification Integration Guide

## Overview

This document provides comprehensive guidance for frontend developers to integrate the real-time winner notification system with TucanBIT's WebSocket infrastructure. The system automatically triggers winner notifications when players win in GrooveTech games.

## Table of Contents

1. [System Architecture](#system-architecture)
2. [WebSocket Connection](#websocket-connection)
3. [Authentication](#authentication)
4. [Message Types](#message-types)
5. [Winner Notification Format](#winner-notification-format)
6. [Frontend Implementation](#frontend-implementation)
7. [Error Handling](#error-handling)
8. [Testing](#testing)
9. [API Endpoints](#api-endpoints)
10. [Examples](#examples)

## System Architecture

The winner notification system operates through WebSocket connections that provide real-time updates for:

- **Balance Updates**: Real-time balance changes
- **Cashback Updates**: Cashback availability notifications
- **Winner Notifications**: Game win celebrations with detailed information

### WebSocket Endpoint

```
ws://localhost:8080/ws/balance/player
```

**Production URL**: Replace `localhost:8080` with your production server URL.

## WebSocket Connection

### Connection Process

1. **Establish WebSocket Connection**: Connect to the WebSocket endpoint
2. **Send Authentication Message**: Send JWT token for authentication
3. **Receive Confirmation**: Get connection confirmation
4. **Listen for Messages**: Handle incoming notifications

### Connection Code Example

```javascript
class TucanBITWebSocket {
    constructor(baseUrl = 'ws://localhost:8080') {
        this.baseUrl = baseUrl;
        this.ws = null;
        this.accessToken = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
    }

    connect(accessToken) {
        this.accessToken = accessToken;
        const wsUrl = `${this.baseUrl}/ws/balance/player`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            this.setupEventHandlers();
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.handleConnectionError();
        }
    }

    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.authenticate();
        };

        this.ws.onmessage = (event) => {
            this.handleMessage(event.data);
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
            this.handleDisconnection();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.handleConnectionError();
        };
    }

    authenticate() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const authMessage = {
                type: 'auth',
                access_token: this.accessToken
            };
            this.ws.send(JSON.stringify(authMessage));
        }
    }

    handleMessage(data) {
        try {
            // Handle initial connection message
            if (data === 'Connected to user balance socket') {
                console.log('Successfully connected to TucanBIT WebSocket');
                return;
            }

            const message = JSON.parse(data);
            this.processMessage(message);
        } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
        }
    }

    processMessage(message) {
        switch (message.type) {
            case 'balance_update':
                this.handleBalanceUpdate(message);
                break;
            case 'cashback_update':
                this.handleCashbackUpdate(message);
                break;
            case 'winner_notification':
                this.handleWinnerNotification(message);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    handleDisconnection() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            setTimeout(() => {
                this.reconnectAttempts++;
                console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
                this.connect(this.accessToken);
            }, this.reconnectDelay * this.reconnectAttempts);
        }
    }

    handleConnectionError() {
        console.error('WebSocket connection failed');
        // Implement fallback or retry logic
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}
```

## Authentication

### JWT Token Requirements

The WebSocket connection requires a valid JWT access token obtained from the login endpoint.

### Authentication Message Format

```json
{
    "type": "auth",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Getting Access Token

```javascript
async function login(email, password) {
    const response = await fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            login_id: email,
            password: password
        })
    });

    if (response.ok) {
        const data = await response.json();
        return data.access_token;
    } else {
        throw new Error('Login failed');
    }
}
```

## Message Types

The WebSocket system sends three types of messages:

### 1. Balance Update Messages

```json
{
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "balance": "20595.83",
    "balance_formatted": "$20,595.83",
    "currency": "USD"
}
```

### 2. Cashback Update Messages

```json
{
    "type": "cashback_update",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "data": {
        "available_cashback": "150.00",
        "current_tier": {
            "tier_name": "Gold",
            "tier_level": 3
        }
    },
    "message": "Cashback availability updated"
}
```

### 3. Winner Notification Messages

```json
{
    "type": "winner_notification",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "data": {
        "username": "P-uBKtmkyl5LPo",
        "email": "ashenafialemu9898@gmail.com",
        "game_name": "GrooveTech Game 82695",
        "game_id": "82695",
        "bet_amount": "10.00",
        "win_amount": "15.00",
        "net_winnings": "5.00",
        "currency": "USD",
        "timestamp": "2025-10-04T07:36:40.816Z",
        "session_id": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14",
        "round_id": "round_1759552600_26083",
        "transaction_id": "winner_test_1759552600_30579"
    },
    "message": "Congratulations! You won!"
}
```

## Winner Notification Format

### Data Structure

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | Player's username |
| `email` | string | Player's email address |
| `game_name` | string | Display name of the game |
| `game_id` | string | GrooveTech game identifier |
| `bet_amount` | string | Amount wagered (decimal string) |
| `win_amount` | string | Total amount won (decimal string) |
| `net_winnings` | string | Net profit (win_amount - bet_amount) |
| `currency` | string | Currency code (e.g., "USD") |
| `timestamp` | string | ISO 8601 timestamp of the win |
| `session_id` | string | Game session identifier |
| `round_id` | string | Game round identifier |
| `transaction_id` | string | Unique transaction identifier |

## Frontend Implementation

### Complete Implementation Example

```javascript
class WinnerNotificationSystem {
    constructor() {
        this.ws = new TucanBITWebSocket();
        this.notificationQueue = [];
        this.isShowingNotification = false;
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Listen for winner notifications
        this.ws.onWinnerNotification = (notification) => {
            this.showWinnerNotification(notification);
        };

        // Listen for balance updates
        this.ws.onBalanceUpdate = (balance) => {
            this.updateBalanceDisplay(balance);
        };

        // Listen for cashback updates
        this.ws.onCashbackUpdate = (cashback) => {
            this.updateCashbackDisplay(cashback);
        };
    }

    async initialize(accessToken) {
        try {
            await this.ws.connect(accessToken);
            console.log('Winner notification system initialized');
        } catch (error) {
            console.error('Failed to initialize winner notification system:', error);
        }
    }

    showWinnerNotification(notification) {
        const { data } = notification;
        
        // Create notification data
        const notificationData = {
            id: data.transaction_id,
            username: data.username,
            gameName: data.game_name,
            gameId: data.game_id,
            betAmount: parseFloat(data.bet_amount),
            winAmount: parseFloat(data.win_amount),
            netWinnings: parseFloat(data.net_winnings),
            currency: data.currency,
            timestamp: new Date(data.timestamp),
            sessionId: data.session_id,
            roundId: data.round_id,
            transactionId: data.transaction_id
        };

        // Add to queue if already showing a notification
        if (this.isShowingNotification) {
            this.notificationQueue.push(notificationData);
            return;
        }

        this.displayNotification(notificationData);
    }

    displayNotification(notification) {
        this.isShowingNotification = true;

        // Create notification element
        const notificationElement = this.createNotificationElement(notification);
        
        // Add to DOM
        document.body.appendChild(notificationElement);

        // Animate in
        setTimeout(() => {
            notificationElement.classList.add('show');
        }, 100);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            this.removeNotification(notificationElement);
        }, 5000);

        // Process next notification in queue
        setTimeout(() => {
            this.processNextNotification();
        }, 6000);
    }

    createNotificationElement(notification) {
        const element = document.createElement('div');
        element.className = 'winner-notification';
        element.innerHTML = `
            <div class="notification-content">
                <div class="celebration-icon">🎉</div>
                <div class="notification-text">
                    <h3>Congratulations ${notification.username}!</h3>
                    <p class="game-info">${notification.gameName}</p>
                    <div class="win-details">
                        <div class="bet-amount">
                            <span class="label">Bet:</span>
                            <span class="amount">${this.formatCurrency(notification.betAmount, notification.currency)}</span>
                        </div>
                        <div class="win-amount">
                            <span class="label">Won:</span>
                            <span class="amount">${this.formatCurrency(notification.winAmount, notification.currency)}</span>
                        </div>
                        <div class="net-winnings">
                            <span class="label">Profit:</span>
                            <span class="amount profit">+${this.formatCurrency(notification.netWinnings, notification.currency)}</span>
                        </div>
                    </div>
                </div>
                <button class="close-btn" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;

        return element;
    }

    removeNotification(element) {
        element.classList.add('hide');
        setTimeout(() => {
            if (element.parentElement) {
                element.parentElement.removeChild(element);
            }
        }, 300);
    }

    processNextNotification() {
        this.isShowingNotification = false;
        if (this.notificationQueue.length > 0) {
            const nextNotification = this.notificationQueue.shift();
            this.displayNotification(nextNotification);
        }
    }

    formatCurrency(amount, currency) {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: currency
        }).format(amount);
    }

    updateBalanceDisplay(balance) {
        // Update balance display in UI
        const balanceElement = document.getElementById('user-balance');
        if (balanceElement) {
            balanceElement.textContent = balance.balance_formatted;
        }
    }

    updateCashbackDisplay(cashback) {
        // Update cashback display in UI
        const cashbackElement = document.getElementById('cashback-amount');
        if (cashbackElement) {
            cashbackElement.textContent = this.formatCurrency(
                parseFloat(cashback.data.available_cashback), 
                'USD'
            );
        }
    }
}

// Usage
const winnerSystem = new WinnerNotificationSystem();

// Initialize when user logs in
async function initializeWinnerNotifications() {
    const accessToken = localStorage.getItem('access_token');
    if (accessToken) {
        await winnerSystem.initialize(accessToken);
    }
}

// Call this after successful login
initializeWinnerNotifications();
```

## CSS Styling

```css
.winner-notification {
    position: fixed;
    top: 20px;
    right: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
    z-index: 10000;
    max-width: 400px;
    transform: translateX(100%);
    transition: transform 0.3s ease-in-out;
    overflow: hidden;
}

.winner-notification.show {
    transform: translateX(0);
}

.winner-notification.hide {
    transform: translateX(100%);
}

.notification-content {
    padding: 20px;
    position: relative;
}

.celebration-icon {
    font-size: 2em;
    text-align: center;
    margin-bottom: 10px;
    animation: bounce 1s infinite;
}

@keyframes bounce {
    0%, 20%, 50%, 80%, 100% {
        transform: translateY(0);
    }
    40% {
        transform: translateY(-10px);
    }
    60% {
        transform: translateY(-5px);
    }
}

.notification-text h3 {
    margin: 0 0 10px 0;
    font-size: 1.2em;
    text-align: center;
}

.game-info {
    text-align: center;
    margin-bottom: 15px;
    font-weight: bold;
    opacity: 0.9;
}

.win-details {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-bottom: 10px;
}

.win-details > div {
    display: flex;
    justify-content: space-between;
    padding: 5px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.net-winnings {
    grid-column: 1 / -1;
    font-weight: bold;
    font-size: 1.1em;
}

.profit {
    color: #4ade80;
}

.close-btn {
    position: absolute;
    top: 10px;
    right: 10px;
    background: none;
    border: none;
    color: white;
    font-size: 1.5em;
    cursor: pointer;
    opacity: 0.7;
    transition: opacity 0.2s;
}

.close-btn:hover {
    opacity: 1;
}

/* Mobile responsiveness */
@media (max-width: 768px) {
    .winner-notification {
        top: 10px;
        right: 10px;
        left: 10px;
        max-width: none;
    }
    
    .win-details {
        grid-template-columns: 1fr;
    }
}
```

## Error Handling

### Connection Errors

```javascript
class WebSocketErrorHandler {
    static handleConnectionError(error) {
        console.error('WebSocket connection error:', error);
        
        // Show user-friendly error message
        this.showErrorMessage('Connection lost. Attempting to reconnect...');
        
        // Implement retry logic
        setTimeout(() => {
            this.attemptReconnection();
        }, 5000);
    }

    static handleAuthenticationError() {
        console.error('WebSocket authentication failed');
        
        // Redirect to login or refresh token
        this.showErrorMessage('Session expired. Please log in again.');
        
        // Clear stored token and redirect
        localStorage.removeItem('access_token');
        window.location.href = '/login';
    }

    static handleMessageError(error) {
        console.error('Failed to process WebSocket message:', error);
        // Log error but don't disrupt user experience
    }

    static showErrorMessage(message) {
        // Implement your error notification system
        const errorElement = document.createElement('div');
        errorElement.className = 'error-notification';
        errorElement.textContent = message;
        document.body.appendChild(errorElement);
        
        setTimeout(() => {
            errorElement.remove();
        }, 5000);
    }
}
```

### Reconnection Strategy

```javascript
class ReconnectionManager {
    constructor(wsClient) {
        this.wsClient = wsClient;
        this.maxAttempts = 5;
        this.baseDelay = 1000;
        this.maxDelay = 30000;
        this.attempts = 0;
    }

    attemptReconnection() {
        if (this.attempts >= this.maxAttempts) {
            console.error('Max reconnection attempts reached');
            this.showPermanentError();
            return;
        }

        this.attempts++;
        const delay = Math.min(
            this.baseDelay * Math.pow(2, this.attempts - 1),
            this.maxDelay
        );

        console.log(`Reconnection attempt ${this.attempts} in ${delay}ms`);

        setTimeout(() => {
            this.wsClient.connect(this.wsClient.accessToken);
        }, delay);
    }

    resetAttempts() {
        this.attempts = 0;
    }

    showPermanentError() {
        // Show permanent error message to user
        const errorElement = document.createElement('div');
        errorElement.className = 'permanent-error';
        errorElement.innerHTML = `
            <h3>Connection Lost</h3>
            <p>Unable to connect to the server. Please refresh the page.</p>
            <button onclick="window.location.reload()">Refresh Page</button>
        `;
        document.body.appendChild(errorElement);
    }
}
```

## Testing

### Test WebSocket Connection

```javascript
// Test function for development
async function testWinnerNotification() {
    const ws = new TucanBITWebSocket();
    
    // Mock winner notification for testing
    ws.onWinnerNotification = (notification) => {
        console.log('Test winner notification received:', notification);
        // Your notification display logic here
    };

    // Connect with test token
    const testToken = 'your_test_jwt_token_here';
    await ws.connect(testToken);
    
    console.log('WebSocket test connection established');
}
```

### Manual Testing

1. **Connect to WebSocket**: Use the test HTML file provided
2. **Simulate Winner**: Use the test script to trigger winner notifications
3. **Verify Display**: Check that notifications appear correctly
4. **Test Edge Cases**: Test with multiple rapid wins, connection drops, etc.

## API Endpoints

### Authentication Endpoint

```http
POST /login
Content-Type: application/json

{
    "login_id": "user@example.com",
    "password": "userpassword"
}
```

**Response:**
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "username": "P-uBKtmkyl5LPo",
    "email": "user@example.com"
}
```

### Game Launch Endpoint

```http
POST /api/groove/launch-game
Authorization: Bearer {access_token}
Content-Type: application/json

{
    "game_id": "82695",
    "device_type": "desktop",
    "game_mode": "real",
    "country": "ET",
    "currency": "USD",
    "language": "en_US"
}
```

**Response:**
```json
{
    "success": true,
    "game_url": "https://routerstg.groovegaming.com/game/?...",
    "session_id": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14"
}
```

### GrooveTech Wager API

```http
GET /groove-official?request=wagerAndResult&accountid={account_id}&gamesessionid={session_id}&device=desktop&gameid={game_id}&apiversion=1.2&betamount={bet_amount}&result={win_amount}&roundid={round_id}&transactionid={transaction_id}&gamestatus=completed
```

**Response:**
```json
{
    "code": 200,
    "status": "Success",
    "success": true,
    "transactionid": "winner_test_1759552600_30579",
    "accountid": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "sessionid": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14",
    "roundid": "round_1759552600_26083",
    "gamestatus": "completed",
    "walletTx": "TXN_winner_test_1759552600_30579_1759552600",
    "balance": "20595.83",
    "bonusWin": "0",
    "realMoneyWin": "15",
    "bonusmoneybet": "0",
    "realmoneybet": "10",
    "bonus_balance": "0",
    "real_balance": "20595.83",
    "game_mode": 1,
    "order": "cash_money",
    "apiversion": "1.2"
}
```

## Examples

### Complete Integration Example

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TucanBIT Winner Notifications</title>
    <style>
        /* Include CSS from above */
    </style>
</head>
<body>
    <div id="app">
        <h1>TucanBIT Gaming Platform</h1>
        <div id="user-balance">$0.00</div>
        <div id="cashback-amount">$0.00</div>
        <button onclick="testWinnerNotification()">Test Winner Notification</button>
    </div>

    <script>
        // Include all JavaScript classes from above
        
        // Initialize the system
        const winnerSystem = new WinnerNotificationSystem();
        
        // Test function
        function testWinnerNotification() {
            const mockNotification = {
                type: 'winner_notification',
                user_id: 'test-user-id',
                data: {
                    username: 'TestUser',
                    email: 'test@example.com',
                    game_name: 'Test Game',
                    game_id: '12345',
                    bet_amount: '10.00',
                    win_amount: '25.00',
                    net_winnings: '15.00',
                    currency: 'USD',
                    timestamp: new Date().toISOString(),
                    session_id: 'test-session',
                    round_id: 'test-round',
                    transaction_id: 'test-transaction'
                },
                message: 'Congratulations! You won!'
            };
            
            winnerSystem.showWinnerNotification(mockNotification);
        }
        
        // Initialize when page loads
        window.addEventListener('load', () => {
            const token = localStorage.getItem('access_token');
            if (token) {
                winnerSystem.initialize(token);
            }
        });
    </script>
</body>
</html>
```

## Best Practices

### 1. Connection Management
- Always handle connection drops gracefully
- Implement exponential backoff for reconnections
- Store connection state for UI updates

### 2. Message Handling
- Validate message structure before processing
- Handle malformed messages gracefully
- Queue notifications if user is busy

### 3. Performance
- Limit notification queue size
- Use efficient DOM manipulation
- Implement notification deduplication

### 4. User Experience
- Show connection status to users
- Provide manual refresh options
- Handle offline scenarios

### 5. Security
- Never log sensitive data
- Validate JWT tokens properly
- Use secure WebSocket connections (WSS) in production

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check server URL and port
2. **Authentication Failed**: Verify JWT token validity
3. **Messages Not Received**: Check WebSocket connection state
4. **Notifications Not Displaying**: Verify CSS and DOM manipulation

### Debug Tools

```javascript
// Enable debug logging
const DEBUG = true;

function debugLog(message, data) {
    if (DEBUG) {
        console.log(`[DEBUG] ${message}`, data);
    }
}

// Add to WebSocket class
this.ws.onmessage = (event) => {
    debugLog('WebSocket message received', event.data);
    this.handleMessage(event.data);
};
```

## Support

For technical support or questions about the winner notification system:

1. Check server logs for WebSocket connection issues
2. Verify JWT token expiration
3. Test with the provided test scripts
4. Review this documentation for implementation details

The winner notification system is designed to be robust and user-friendly, providing real-time celebration of player wins while maintaining system stability and performance.