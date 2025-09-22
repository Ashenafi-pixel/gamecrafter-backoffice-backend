# Frontend Developer: Balance WebSocket API Integration

## Summary

The TucanBIT platform provides a **WebSocket API for real-time balance updates** that allows frontend applications to receive instant balance changes without polling. This is essential for providing a seamless user experience where balance updates appear immediately after transactions, bets, wins, or deposits.

## API Details

### Endpoint
```
ws://localhost:8080/ws/balance/player
```

### Authentication
- **Method:** JWT Token sent in WebSocket message
- **Required:** Valid access token from login response

### Response Format
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "balance": "255.50",
  "currency": "USD"
}
```

## Frontend Integration Prompt

### Task: Implement Real-time Balance Updates

**Objective:** Create a WebSocket service to receive real-time balance updates and update the UI accordingly.

### Requirements

1. **WebSocket Connection Management**
   - Establish WebSocket connection to `/ws/balance/player`
   - Implement automatic reconnection on connection loss
   - Handle connection errors gracefully

2. **Authentication**
   - Send JWT token in WebSocket message after connection
   - Handle authentication failures
   - Refresh token when needed

3. **Real-time Updates**
   - Listen for balance update messages
   - Update UI components with new balance data
   - Maintain connection for the duration of user session

4. **Error Handling**
   - Handle WebSocket connection errors
   - Handle authentication errors
   - Implement fallback to REST API if WebSocket fails

### Implementation Guidelines

#### 1. WebSocket Service Class

```javascript
class BalanceWebSocketService {
  constructor(baseUrl, token) {
    this.baseUrl = baseUrl;
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.listeners = new Set();
  }

  connect() {
    try {
      this.ws = new WebSocket(`${this.baseUrl}/ws/balance/player`);
      
      this.ws.onopen = () => {
        console.log('Balance WebSocket connected');
        this.reconnectAttempts = 0;
        this.authenticate();
      };
      
      this.ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        this.notifyListeners(data);
      };
      
      this.ws.onerror = (error) => {
        console.error('Balance WebSocket error:', error);
      };
      
      this.ws.onclose = () => {
        console.log('Balance WebSocket disconnected');
        this.handleReconnection();
      };
      
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  }

  authenticate() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: "auth",
        access_token: this.token,
        data: {}
      }));
    }
  }

  addListener(callback) {
    this.listeners.add(callback);
  }

  removeListener(callback) {
    this.listeners.delete(callback);
  }

  notifyListeners(data) {
    this.listeners.forEach(callback => callback(data));
  }

  handleReconnection() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
      // Fallback to REST API polling
      this.fallbackToRestAPI();
    }
  }

  fallbackToRestAPI() {
    // Implement REST API polling as fallback
    console.log('Falling back to REST API for balance updates');
    // Your REST API polling implementation here
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  updateToken(newToken) {
    this.token = newToken;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.authenticate();
    }
  }
}
```

#### 2. React Hook Implementation

```javascript
import { useEffect, useState, useCallback } from 'react';

export const useBalanceWebSocket = (baseUrl, token) => {
  const [balance, setBalance] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState(null);
  const [wsService, setWsService] = useState(null);

  useEffect(() => {
    if (!token) return;

    const service = new BalanceWebSocketService(baseUrl, token);
    setWsService(service);

    const handleBalanceUpdate = (data) => {
      setBalance(data);
      setError(null);
    };

    const handleError = (err) => {
      setError(err);
    };

    service.addListener(handleBalanceUpdate);
    service.connect();

    return () => {
      service.removeListener(handleBalanceUpdate);
      service.disconnect();
    };
  }, [baseUrl, token]);

  const updateToken = useCallback((newToken) => {
    if (wsService) {
      wsService.updateToken(newToken);
    }
  }, [wsService]);

  return {
    balance,
    isConnected,
    error,
    updateToken
  };
};
```

#### 3. React Component Usage

```jsx
import React from 'react';
import { useBalanceWebSocket } from './hooks/useBalanceWebSocket';

const BalanceDisplay = ({ userToken }) => {
  const { balance, isConnected, error } = useBalanceWebSocket(
    'ws://localhost:8080',
    userToken
  );

  if (error) {
    return (
      <div className="balance-error">
        <p>Error loading balance: {error.message}</p>
        <button onClick={() => window.location.reload()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="balance-display">
      <div className="balance-amount">
        {balance ? `$${balance.balance}` : 'Loading...'}
      </div>
      <div className="balance-currency">
        {balance?.currency || 'USD'}
      </div>
      <div className="connection-status">
        {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
      </div>
    </div>
  );
};

export default BalanceDisplay;
```

#### 4. Vue.js Implementation

```javascript
// balance-websocket.js
export class BalanceWebSocketService {
  // ... (same implementation as above)
}

// useBalanceWebSocket.js
import { ref, onMounted, onUnmounted } from 'vue';

export function useBalanceWebSocket(baseUrl, token) {
  const balance = ref(null);
  const isConnected = ref(false);
  const error = ref(null);
  let wsService = null;

  onMounted(() => {
    if (!token) return;

    wsService = new BalanceWebSocketService(baseUrl, token);
    
    wsService.addListener((data) => {
      balance.value = data;
      error.value = null;
    });
    
    wsService.connect();
  });

  onUnmounted(() => {
    if (wsService) {
      wsService.disconnect();
    }
  });

  const updateToken = (newToken) => {
    if (wsService) {
      wsService.updateToken(newToken);
    }
  };

  return {
    balance,
    isConnected,
    error,
    updateToken
  };
}
```

### Testing Checklist

- [ ] WebSocket connection establishes successfully
- [ ] Authentication message is sent correctly
- [ ] Balance updates are received and displayed
- [ ] Connection errors are handled gracefully
- [ ] Reconnection works after connection loss
- [ ] Multiple components can listen to the same WebSocket
- [ ] Token refresh updates the WebSocket authentication
- [ ] Fallback to REST API works when WebSocket fails

### Error Scenarios to Handle

1. **Invalid Token:** Show authentication error, redirect to login
2. **Connection Lost:** Attempt reconnection, show connection status
3. **Server Error:** Fallback to REST API polling
4. **Network Issues:** Show offline indicator, retry when online
5. **Token Expired:** Refresh token and re-authenticate

### Performance Considerations

- Use a single WebSocket connection per user session
- Implement connection pooling if multiple components need balance updates
- Debounce rapid balance updates to avoid UI flickering
- Cache the last known balance for offline scenarios
- Implement proper cleanup on component unmount

### Security Notes

- Always validate balance data before updating UI
- Never trust WebSocket data without server-side validation
- Implement rate limiting for reconnection attempts
- Use secure WebSocket connections (wss://) in production
- Log balance updates for audit purposes

This implementation will provide real-time balance updates that enhance the user experience by showing immediate feedback for all balance-related transactions.