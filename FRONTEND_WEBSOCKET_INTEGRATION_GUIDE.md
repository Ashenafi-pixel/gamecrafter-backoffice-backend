# TucanBIT WebSocket Integration Guide for Frontend Developer

## 🎯 **Objective**
Implement real-time WebSocket notifications for cashback updates and balance changes in the TucanBIT frontend application.

## 🔌 **WebSocket Connection Details**

### **Connection URL:**
```
wss://api.tucanbit.tv/ws/balance/player
```

### **Authentication:**
Send JWT token in the initial message after connection:
```javascript
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 📨 **Message Types & Formats**

### **1. Balance Updates**
Triggered when user balance changes (deposits, withdrawals, cashback claims, etc.)

```javascript
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "balance": "2430.76",
  "balance_formatted": "$2,430.76",
  "currency": "USD"
}
```

### **2. Cashback Updates**
Triggered when cashback is earned or claimed

```javascript
{
  "type": "cashback_update",
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "data": {
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "current_tier": {
      "id": "5e5cbd7f-7b93-489a-9a01-5a9c0f1e2c94",
      "tier_name": "Bronze",
      "tier_level": 1,
      "cashback_percentage": "0.5",
      "daily_cashback_limit": "50",
      "is_active": true
    },
    "level_progress": "0",
    "total_ggr": "2",
    "available_cashback": "0.01",
    "pending_cashback": "0",
    "total_claimed": "0",
    "next_tier_ggr": "1000",
    "daily_limit": "50",
    "weekly_limit": null,
    "monthly_limit": null,
    "special_benefits": {},
    "last_game_info": {
      "game_id": "82695",
      "game_name": "Sweet Bonanza",
      "house_edge": "0.02",
      "house_edge_percent": "2%",
      "cashback_rate": "0.5",
      "cashback_percent": "0.5%",
      "expected_ggr": "2.00",
      "earned_cashback": "0.01",
      "bet_amount": "100.00",
      "transaction_id": "result-1759185679",
      "timestamp": "2025-09-29T21:51:30.16835Z"
    }
  },
  "message": "Cashback availability updated"
}
```

## 🛠 **Implementation Requirements**

### **1. WebSocket Connection Setup**
```javascript
class TucanBitWebSocket {
  constructor(jwtToken) {
    this.token = jwtToken;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }

  connect() {
    try {
      this.ws = new WebSocket('wss://api.tucanbit.tv/ws/balance/player');
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.authenticate();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.handleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

    } catch (error) {
      console.error('Failed to connect:', error);
    }
  }

  authenticate() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const authMessage = {
        type: 'auth',
        access_token: this.token
      };
      this.ws.send(JSON.stringify(authMessage));
    }
  }

  handleMessage(data) {
    try {
      const message = JSON.parse(data);
      
      if (message.type === 'cashback_update') {
        this.onCashbackUpdate(message.data);
      } else if (message.balance !== undefined) {
        this.onBalanceUpdate(message);
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }

  onBalanceUpdate(balanceData) {
    // Update balance display in UI
    console.log('Balance updated:', balanceData);
    // Example: updateBalanceUI(balanceData.balance_formatted);
  }

  onCashbackUpdate(cashbackData) {
    // Update cashback display in UI
    console.log('Cashback updated:', cashbackData);
    // Example: updateCashbackUI(cashbackData);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
```

### **2. UI Integration Points**

#### **Balance Display Update:**
```javascript
function updateBalanceUI(formattedBalance) {
  // Update balance in header/navigation
  const balanceElement = document.querySelector('.user-balance');
  if (balanceElement) {
    balanceElement.textContent = formattedBalance;
  }
  
  // Update balance in any modals or components
  const balanceElements = document.querySelectorAll('.balance-amount');
  balanceElements.forEach(el => el.textContent = formattedBalance);
}
```

#### **Cashback Display Update:**
```javascript
function updateCashbackUI(cashbackData) {
  // Update available cashback amount
  const availableCashbackEl = document.querySelector('.available-cashback');
  if (availableCashbackEl) {
    availableCashbackEl.textContent = `$${cashbackData.available_cashback}`;
  }

  // Update tier information
  const tierEl = document.querySelector('.current-tier');
  if (tierEl) {
    tierEl.textContent = `${cashbackData.current_tier.tier_name} (${cashbackData.current_tier.cashback_percentage}%)`;
  }

  // Update progress bar
  const progressEl = document.querySelector('.tier-progress');
  if (progressEl) {
    const progress = (parseFloat(cashbackData.level_progress) / parseFloat(cashbackData.next_tier_ggr)) * 100;
    progressEl.style.width = `${Math.min(progress, 100)}%`;
  }

  // Show notification for new cashback
  if (parseFloat(cashbackData.available_cashback) > 0) {
    showCashbackNotification(cashbackData);
  }
}
```

#### **Cashback Notification:**
```javascript
function showCashbackNotification(cashbackData) {
  // Create toast notification
  const notification = document.createElement('div');
  notification.className = 'cashback-notification';
  notification.innerHTML = `
    <div class="notification-content">
      <h4>🎉 Cashback Earned!</h4>
      <p>You earned $${cashbackData.available_cashback} cashback from ${cashbackData.last_game_info?.game_name || 'your last game'}!</p>
      <p>Tier: ${cashbackData.current_tier.tier_name} (${cashbackData.current_tier.cashback_percentage}%)</p>
      <button onclick="claimCashback()">Claim Now</button>
    </div>
  `;
  
  document.body.appendChild(notification);
  
  // Auto-remove after 10 seconds
  setTimeout(() => {
    notification.remove();
  }, 10000);
}
```

### **3. React/Vue/Angular Integration**

#### **React Hook Example:**
```javascript
import { useEffect, useState, useRef } from 'react';

export const useTucanBitWebSocket = (jwtToken) => {
  const [balance, setBalance] = useState(null);
  const [cashback, setCashback] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef(null);

  useEffect(() => {
    if (!jwtToken) return;

    const connect = () => {
      wsRef.current = new WebSocket('wss://api.tucanbit.tv/ws/balance/player');
      
      wsRef.current.onopen = () => {
        setIsConnected(true);
        // Send authentication
        wsRef.current.send(JSON.stringify({
          type: 'auth',
          access_token: jwtToken
        }));
      };

      wsRef.current.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        if (data.type === 'cashback_update') {
          setCashback(data.data);
        } else if (data.balance !== undefined) {
          setBalance(data);
        }
      };

      wsRef.current.onclose = () => {
        setIsConnected(false);
        // Attempt to reconnect after 3 seconds
        setTimeout(connect, 3000);
      };
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [jwtToken]);

  return { balance, cashback, isConnected };
};
```

## 🎮 **Cashback Flow Context**

### **When Cashback is Earned:**
1. User launches a game (Sweet Bonanza - ID: 82695)
2. User places a wager ($100)
3. User gets a result (loses $50)
4. System calculates GGR: $100 × 2% house edge = $2
5. System calculates cashback: $2 × 0.5% = $0.01
6. **WebSocket notification sent** with `available_cashback: "0.01"`

### **When Cashback is Claimed:**
1. User clicks "Claim Cashback" button
2. Frontend calls `POST /user/cashback/claim` API
3. Cashback is credited to user balance
4. **WebSocket notification sent** with updated balance and `available_cashback: "0"`

## 🧪 **Testing**

### **Test WebSocket Connection:**
1. Open browser developer tools
2. Connect to `wss://api.tucanbit.tv/ws/balance/player`
3. Send authentication message
4. Play a game to trigger notifications

### **Test Files Available:**
- `test_websocket_connection.html` - Visual WebSocket test interface
- `test_websocket.sh` - Command line WebSocket test

## 🔧 **Error Handling**

### **Connection Issues:**
- Implement exponential backoff for reconnection
- Show connection status indicator in UI
- Handle authentication failures gracefully

### **Message Parsing:**
- Always wrap JSON.parse in try-catch
- Validate message structure before processing
- Log unknown message types for debugging

## 📱 **Mobile Considerations**

### **Background Handling:**
- WebSocket may disconnect when app goes to background
- Implement reconnection logic for when app becomes active
- Consider using push notifications as fallback

### **Battery Optimization:**
- Use heartbeat/ping mechanism to keep connection alive
- Implement smart reconnection based on user activity

## 🎨 **UI/UX Recommendations**

### **Visual Indicators:**
- Show WebSocket connection status (green/red dot)
- Animate balance updates with smooth transitions
- Use toast notifications for cashback alerts
- Add progress bars for tier progression

### **User Feedback:**
- Show "Cashback earned!" animations
- Display tier upgrade notifications
- Provide clear cashback claim buttons
- Show cashback history/earnings

## 🔐 **Security Notes**

- JWT tokens expire - implement token refresh logic
- Validate all incoming WebSocket messages
- Don't expose sensitive data in console logs
- Use HTTPS/WSS only in production

## 📊 **Monitoring & Analytics**

### **Track These Events:**
- WebSocket connection success/failure
- Cashback notifications received
- User interactions with cashback UI
- Tier progression milestones

---

**Ready to implement? Start with the basic WebSocket connection and gradually add the UI updates. The backend is fully functional and ready to send notifications!** 🚀