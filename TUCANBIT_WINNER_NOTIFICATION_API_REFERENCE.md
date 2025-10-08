# TucanBIT Winner Notification API Reference

## Overview

This document provides detailed API specifications for the TucanBIT Winner Notification System, including request/response formats, WebSocket message structures, and integration examples for frontend developers.

## Table of Contents

1. [Authentication API](#authentication-api)
2. [Game Session Management](#game-session-management)
3. [WebSocket Connection](#websocket-connection)
4. [Winner Notification Messages](#winner-notification-messages)
5. [GrooveTech Integration](#groovetech-integration)
6. [Error Responses](#error-responses)
7. [Complete Integration Flow](#complete-integration-flow)

## Authentication API

### Login Request

**Endpoint:** `POST /login`

**Request Headers:**
```http
Content-Type: application/json
```

**Request Body:**
```json
{
    "login_id": "ashenafialemu9898@gmail.com",
    "password": "Secure!Pass123"
}
```

**Success Response (200 OK):**
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1IiwiaXNfdmVyaWZpZWQiOmZhbHNlLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInBob25lX3ZlcmlmaWVkIjpmYWxzZSwiZXhwIjoxNzU5NjM5MDAwLCJpYXQiOjE3NTk1NTI2MDAsImlzcyI6InR1Y2FuYml0Iiwic3ViIjoiYTVlMTY4ZmItMTY4ZS00MTgzLTg0YzUtZDQ5MDM4Y2UwMGI1In0.y-WReWXbMie6c2SJCd7RGiBcpy0SYIyUd7vy0pHPpFQ",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "username": "P-uBKtmkyl5LPo",
    "email": "ashenafialemu9898@gmail.com",
    "is_verified": false,
    "email_verified": false,
    "phone_verified": false
}
```

**Error Response (400 Bad Request):**
```json
{
    "code": 400,
    "message": "Invalid credentials"
}
```

## Game Session Management

### Launch Game Request

**Endpoint:** `POST /api/groove/launch-game`

**Request Headers:**
```http
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**
```json
{
    "game_id": "82695",
    "device_type": "desktop",
    "game_mode": "real",
    "country": "ET",
    "currency": "USD",
    "language": "en_US",
    "is_test_account": false,
    "reality_check_elapsed": 0,
    "reality_check_interval": 60
}
```

**Success Response (200 OK):**
```json
{
    "success": true,
    "game_url": "https://routerstg.groovegaming.com/game/?accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&country=ET&nogsgameid=82695&nogslang=en_US&nogsmode=real&nogsoperatorid=3818&nogscurrency=USD&sessionid=3818_1104eb29-5ecb-4727-9044-f9b06dec8d14&homeurl=https://tucanbit.tv&license=Curacao&is_test_account=false&device_type=desktop&realityCheckElapsed=0&realityCheckInterval=60&historyUrl=https://tucanbit.tv/history&exitUrl=https://tucanbit.tv",
    "session_id": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14"
}
```

**Error Response (401 Unauthorized):**
```json
{
    "code": 401,
    "message": "Invalid or missing authentication token"
}
```

### Validate Game Session

**Endpoint:** `GET /api/groove/validate-session/{session_id}`

**Request Headers:**
```http
Authorization: Bearer {access_token}
```

**Success Response (200 OK):**
```json
{
    "id": "f8e01c72-3833-4f6d-b145-e437a5bd9df7",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "session_id": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14",
    "game_id": "82695",
    "device_type": "desktop",
    "game_mode": "real",
    "groove_url": "",
    "home_url": "https://tucanbit.tv",
    "exit_url": "https://tucanbit.tv",
    "history_url": "https://tucanbit.tv/history",
    "license_type": "Curacao",
    "is_test_account": false,
    "reality_check_elapsed": 0,
    "reality_check_interval": 60,
    "created_at": "2025-10-04T04:36:40.730984Z",
    "expires_at": "2025-10-04T06:36:40.730984Z",
    "is_active": true,
    "last_activity": "2025-10-04T04:36:40.737651Z"
}
```

**Error Response (404 Not Found):**
```json
{
    "code": 404,
    "message": "Game session not found or expired"
}
```

## WebSocket Connection

### Connection Endpoint

**WebSocket URL:** `ws://localhost:8080/ws/balance/player`

**Production URL:** `wss://your-domain.com/ws/balance/player`

### Authentication Message

**Message Format:**
```json
{
    "type": "auth",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Connection Confirmation:**
```
"Connected to user balance socket"
```

### Initial Balance Message

**Message Format:**
```json
{
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "balance": "20590.83",
    "balance_formatted": "$20,590.83",
    "currency": "USD"
}
```

## Winner Notification Messages

### Winner Notification Message Structure

**Message Type:** `winner_notification`

**Complete Message:**
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

### Winner Notification Data Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `username` | string | Player's display name | "P-uBKtmkyl5LPo" |
| `email` | string | Player's email address | "ashenafialemu9898@gmail.com" |
| `game_name` | string | Display name of the game | "GrooveTech Game 82695" |
| `game_id` | string | GrooveTech game identifier | "82695" |
| `bet_amount` | string | Amount wagered (decimal) | "10.00" |
| `win_amount` | string | Total amount won (decimal) | "15.00" |
| `net_winnings` | string | Net profit (win - bet) | "5.00" |
| `currency` | string | Currency code | "USD" |
| `timestamp` | string | ISO 8601 timestamp | "2025-10-04T07:36:40.816Z" |
| `session_id` | string | Game session ID | "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14" |
| `round_id` | string | Game round identifier | "round_1759552600_26083" |
| `transaction_id` | string | Unique transaction ID | "winner_test_1759552600_30579" |

### Balance Update Message

**Message Type:** `balance_update`

**Message Format:**
```json
{
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "balance": "20595.83",
    "balance_formatted": "$20,595.83",
    "currency": "USD"
}
```

### Cashback Update Message

**Message Type:** `cashback_update`

**Message Format:**
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

## GrooveTech Integration

### Get Account Request

**Endpoint:** `GET /groove-official`

**Query Parameters:**
```
request=getaccount
accountid={user_id}
gamesessionid={session_id}
device=desktop
apiversion=1.2
```

**Example Request:**
```http
GET /groove-official?request=getaccount&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=3818_1104eb29-5ecb-4727-9044-f9b06dec8d14&device=desktop&apiversion=1.2
```

**Success Response (200 OK):**
```json
{
    "accountid": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "apiversion": "1.2",
    "bonus_balance": 0,
    "city": "Addis Ababa",
    "code": 200,
    "country": "ET",
    "currency": "USD",
    "game_mode": 1,
    "gamesessionid": "3818_1104eb29-5ecb-4727-9044-f9b06dec8d14",
    "order": "cash_money",
    "real_balance": 20590.83,
    "status": "Success"
}
```

### Wager and Result Request

**Endpoint:** `GET /groove-official`

**Query Parameters:**
```
request=wagerAndResult
accountid={user_id}
gamesessionid={session_id}
device=desktop
gameid={game_id}
apiversion=1.2
betamount={bet_amount}
result={win_amount}
roundid={round_id}
transactionid={transaction_id}
gamestatus=completed
```

**Example Request:**
```http
GET /groove-official?request=wagerAndResult&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid=3818_1104eb29-5ecb-4727-9044-f9b06dec8d14&device=desktop&gameid=82695&apiversion=1.2&betamount=10.0&result=15.0&roundid=round_1759552600_26083&transactionid=winner_test_1759552600_30579&gamestatus=completed
```

**Success Response (200 OK):**
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

**Error Response (200 OK with Error Code):**
```json
{
    "code": 1000,
    "status": "Not logged on",
    "error": "Player session is invalid or expired"
}
```

## Error Responses

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| 200 | Success | Request successful |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Invalid or missing authentication |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Server error |
| 1000 | Not logged on | Invalid game session |

### Authentication Errors

**401 Unauthorized:**
```json
{
    "code": 401,
    "message": "Invalid or missing authentication token"
}
```

**400 Bad Request (Login):**
```json
{
    "code": 400,
    "message": "Invalid credentials"
}
```

### Session Errors

**404 Not Found (Session):**
```json
{
    "code": 404,
    "message": "Game session not found or expired"
}
```

**1000 Not logged on (GrooveTech):**
```json
{
    "code": 1000,
    "status": "Not logged on",
    "error": "Player session is invalid or expired"
}
```

### WebSocket Errors

**Connection Error:**
```json
{
    "status": "error",
    "message": "Missing or invalid access_token"
}
```

**Authentication Error:**
```json
{
    "status": "error",
    "message": "Invalid or expired access_token"
}
```

## Complete Integration Flow

### Step-by-Step Integration

1. **User Login**
   ```javascript
   const loginResponse = await fetch('/login', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       body: JSON.stringify({
           login_id: 'user@example.com',
           password: 'password'
       })
   });
   const { access_token } = await loginResponse.json();
   ```

2. **Launch Game Session**
   ```javascript
   const launchResponse = await fetch('/api/groove/launch-game', {
       method: 'POST',
       headers: {
           'Authorization': `Bearer ${access_token}`,
           'Content-Type': 'application/json'
       },
       body: JSON.stringify({
           game_id: '82695',
           device_type: 'desktop',
           game_mode: 'real'
       })
   });
   const { session_id } = await launchResponse.json();
   ```

3. **Connect to WebSocket**
   ```javascript
   const ws = new WebSocket('ws://localhost:8080/ws/balance/player');
   
   ws.onopen = () => {
       ws.send(JSON.stringify({
           type: 'auth',
           access_token: access_token
       }));
   };
   ```

4. **Handle Winner Notifications**
   ```javascript
   ws.onmessage = (event) => {
       if (event.data === 'Connected to user balance socket') {
           console.log('Connected successfully');
           return;
       }
       
       const message = JSON.parse(event.data);
       
       if (message.type === 'winner_notification') {
           showWinnerNotification(message.data);
       }
   };
   ```

5. **Display Winner Notification**
   ```javascript
   function showWinnerNotification(data) {
       const notification = {
           username: data.username,
           gameName: data.game_name,
           betAmount: parseFloat(data.bet_amount),
           winAmount: parseFloat(data.win_amount),
           netWinnings: parseFloat(data.net_winnings),
           currency: data.currency
       };
       
       // Display notification in UI
       displayNotification(notification);
   }
   ```

### Complete Frontend Implementation

```javascript
class TucanBITWinnerSystem {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.wsUrl = baseUrl.replace('http', 'ws');
        this.accessToken = null;
        this.sessionId = null;
        this.ws = null;
    }

    async login(email, password) {
        const response = await fetch(`${this.baseUrl}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ login_id: email, password })
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const data = await response.json();
        this.accessToken = data.access_token;
        return data;
    }

    async launchGame(gameId = '82695') {
        const response = await fetch(`${this.baseUrl}/api/groove/launch-game`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.accessToken}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                game_id: gameId,
                device_type: 'desktop',
                game_mode: 'real'
            })
        });

        if (!response.ok) {
            throw new Error('Game launch failed');
        }

        const data = await response.json();
        this.sessionId = data.session_id;
        return data;
    }

    connectWebSocket() {
        this.ws = new WebSocket(`${this.wsUrl}/ws/balance/player`);
        
        this.ws.onopen = () => {
            this.ws.send(JSON.stringify({
                type: 'auth',
                access_token: this.accessToken
            }));
        };

        this.ws.onmessage = (event) => {
            this.handleMessage(event.data);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
        };
    }

    handleMessage(data) {
        if (data === 'Connected to user balance socket') {
            console.log('WebSocket connected successfully');
            return;
        }

        try {
            const message = JSON.parse(data);
            
            switch (message.type) {
                case 'winner_notification':
                    this.onWinnerNotification(message.data);
                    break;
                case 'balance_update':
                    this.onBalanceUpdate(message);
                    break;
                case 'cashback_update':
                    this.onCashbackUpdate(message);
                    break;
                default:
                    console.log('Unknown message type:', message.type);
            }
        } catch (error) {
            console.error('Failed to parse message:', error);
        }
    }

    onWinnerNotification(data) {
        console.log('Winner notification received:', data);
        
        // Trigger custom event for UI components
        const event = new CustomEvent('winnerNotification', {
            detail: {
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
            }
        });
        
        document.dispatchEvent(event);
    }

    onBalanceUpdate(data) {
        console.log('Balance update received:', data);
        
        const event = new CustomEvent('balanceUpdate', {
            detail: {
                balance: parseFloat(data.balance),
                balanceFormatted: data.balance_formatted,
                currency: data.currency
            }
        });
        
        document.dispatchEvent(event);
    }

    onCashbackUpdate(data) {
        console.log('Cashback update received:', data);
        
        const event = new CustomEvent('cashbackUpdate', {
            detail: data.data
        });
        
        document.dispatchEvent(event);
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

// Usage Example
const winnerSystem = new TucanBITWinnerSystem();

// Listen for winner notifications
document.addEventListener('winnerNotification', (event) => {
    const { username, gameName, betAmount, winAmount, netWinnings, currency } = event.detail;
    
    // Display notification
    showNotification(`${username} won ${netWinnings} ${currency} in ${gameName}!`);
});

// Initialize system
async function initializeSystem() {
    try {
        await winnerSystem.login('user@example.com', 'password');
        await winnerSystem.launchGame();
        winnerSystem.connectWebSocket();
        console.log('Winner notification system initialized');
    } catch (error) {
        console.error('Failed to initialize system:', error);
    }
}
```

## Testing

### Test Data

**Test User Credentials:**
- Email: `ashenafialemu9898@gmail.com`
- Password: `Secure!Pass123`
- User ID: `a5e168fb-168e-4183-84c5-d49038ce00b5`

**Test Game Configuration:**
- Game ID: `82695`
- Device Type: `desktop`
- Game Mode: `real`

### Manual Testing Commands

```bash
# Test login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"login_id":"ashenafialemu9898@gmail.com","password":"Secure!Pass123"}'

# Test game launch
curl -X POST http://localhost:8080/api/groove/launch-game \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{"game_id":"82695","device_type":"desktop","game_mode":"real"}'

# Test winner notification trigger
curl "http://localhost:8080/groove-official?request=wagerAndResult&accountid=a5e168fb-168e-4183-84c5-d49038ce00b5&gamesessionid={session_id}&device=desktop&gameid=82695&apiversion=1.2&betamount=10.0&result=25.0&roundid=test_round&transactionid=test_txn&gamestatus=completed"
```

This API reference provides complete specifications for integrating the TucanBIT Winner Notification System into any frontend application. The system automatically triggers winner notifications when players win in GrooveTech games, providing real-time celebration of player achievements.