# TucanBIT WebSocket APIs Documentation

## Overview
This document provides comprehensive documentation for all WebSocket APIs available in the TucanBIT platform. These APIs enable real-time communication between the frontend and backend for various features including balance updates, notifications, game streaming, and session management.

## Table of Contents
- [Authentication](#authentication)
- [WebSocket Endpoints](#websocket-endpoints)
- [Request/Response Formats](#requestresponse-formats)
- [Testing Guide](#testing-guide)
- [Error Handling](#error-handling)
- [Frontend Integration](#frontend-integration)

## Authentication

All WebSocket endpoints require JWT authentication. The authentication is performed by sending a message after establishing the WebSocket connection.

### Authentication Message Format
```json
{
  "type": "auth",
  "access_token": "your_jwt_token_here",
  "data": {}
}
```

## WebSocket Endpoints

### 1. General WebSocket Connection
- **Endpoint:** `ws://localhost:8080/ws`
- **Method:** WebSocket
- **Description:** General WebSocket connection for game interactions and broadcasts

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
- **Connection:** Establishes WebSocket connection
- **Broadcast:** Receives game-related broadcasts and updates

---

### 2. Single Player Game Stream
- **Endpoint:** `ws://localhost:8080/ws/single/player`
- **Method:** WebSocket
- **Description:** WebSocket for single-player game streaming

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
```json
{
  "message": "connected"
}
```

---

### 3. Player Level Updates
- **Endpoint:** `ws://localhost:8080/ws/level/player`
- **Method:** WebSocket
- **Description:** Real-time player level updates

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
- **Connection:** Establishes WebSocket connection
- **Updates:** Receives player level changes in real-time

---

### 4. Notifications WebSocket
- **Endpoint:** `ws://localhost:8080/ws/notify`
- **Method:** WebSocket
- **Description:** Real-time notification delivery

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Connected to notification socket"
}
```

**Notification Message Format:**
```json
{
  "type": "notification",
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "Notification Title",
    "message": "Notification message",
    "type": "INFO|WARNING|ERROR|SUCCESS",
    "is_read": false,
    "created_at": "2025-01-16T10:30:00Z",
    "read_at": null
  }
}
```

---

### 5. Player Progress Bar Updates
- **Endpoint:** `ws://localhost:8080/ws/player/level/progress`
- **Method:** WebSocket
- **Description:** Real-time progress bar updates for player level

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
- **Connection:** Establishes WebSocket connection
- **Updates:** Receives progress bar updates in real-time

---

### 6. Squads Progress Bar Updates
- **Endpoint:** `ws://localhost:8080/ws/squads/progress`
- **Method:** WebSocket
- **Description:** Real-time squads progress bar updates

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
- **Connection:** Establishes WebSocket connection
- **Updates:** Receives squads progress updates in real-time

---

### 7. User Balance Updates ⭐
- **Endpoint:** `ws://localhost:8080/ws/balance/player`
- **Method:** WebSocket
- **Description:** Real-time user balance updates

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "balance": "255.50",
  "currency": "USD"
}
```

---

### 8. Session Monitoring
- **Endpoint:** `ws://localhost:8080/ws/session`
- **Method:** WebSocket
- **Description:** Session monitoring and management

**Request:**
```json
{
  "type": "auth",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "data": {}
}
```

**Response:**
- **Connection:** Establishes WebSocket connection
- **Session Events:** Receives session-related events

**Session Event Message Format:**
```json
{
  "type": "session_warning|session_expired|session_refreshed",
  "message": "Human-readable message",
  "action": "login|refresh|dismiss",
  "timeout": 300,
  "data": {
    "session_id": "uuid",
    "expires_at": "2025-01-16T11:00:00Z"
  }
}
```

## Request/Response Formats

### WebSocket Message Request (WSMessageRequest)
```json
{
  "type": "auth",
  "access_token": "jwt_token_here",
  "data": {}
}
```

### WebSocket Response (WSRes)
```json
{
  "type": "response_type",
  "data": {
    // Response data here
  }
}
```

### Broadcast Payload (BroadCastPayload)
```json
{
  "user_id": "uuid",
  "conn": "websocket_connection_object"
}
```

## Testing Guide

### JavaScript Test Script for All Endpoints

```javascript
const endpoints = [
  { name: 'General WS', url: 'ws://localhost:8080/ws' },
  { name: 'Single Player', url: 'ws://localhost:8080/ws/single/player' },
  { name: 'Player Level', url: 'ws://localhost:8080/ws/level/player' },
  { name: 'Notifications', url: 'ws://localhost:8080/ws/notify' },
  { name: 'Progress Bar', url: 'ws://localhost:8080/ws/player/level/progress' },
  { name: 'Squads Progress', url: 'ws://localhost:8080/ws/squads/progress' },
  { name: 'Balance Updates', url: 'ws://localhost:8080/ws/balance/player' },
  { name: 'Session Monitor', url: 'ws://localhost:8080/ws/session' }
];

const token = "your_jwt_token_here";

endpoints.forEach(endpoint => {
  const ws = new WebSocket(endpoint.url);
  
  ws.onopen = function() {
    console.log(`✅ Connected to ${endpoint.name}`);
    
    const authMessage = {
      type: "auth",
      access_token: token,
      data: {}
    };
    
    ws.send(JSON.stringify(authMessage));
  };
  
  ws.onmessage = function(event) {
    console.log(`${endpoint.name} - Received:`, JSON.parse(event.data));
  };
  
  ws.onerror = function(error) {
    console.error(`❌ ${endpoint.name} - Error:`, error);
  };
  
  ws.onclose = function() {
    console.log(`🔌 ${endpoint.name} - Connection closed`);
  };
});
```

### Postman WebSocket Testing

1. **Create New WebSocket Request**
2. **Set URL:** `ws://localhost:8080/ws/balance/player` (or any endpoint)
3. **Connect**
4. **Send Message:**
   ```json
   {
     "type": "auth",
     "access_token": "your_jwt_token_here",
     "data": {}
   }
   ```

## Error Handling

### Common Error Responses

**Invalid Token:**
```json
{
  "status": "error",
  "message": "Invalid or expired access_token"
}
```

**Missing Token:**
```json
{
  "status": "error",
  "message": "Missing or invalid access_token"
}
```

**Connection Errors:**
- WebSocket connection failures return HTTP 400 with WebSocket protocol error
- Invalid endpoints return HTTP 404
- Server errors return HTTP 500

## Frontend Integration

### React/JavaScript Example

```javascript
class WebSocketService {
  constructor() {
    this.connections = new Map();
  }

  connect(endpoint, token, onMessage, onError) {
    const ws = new WebSocket(`ws://localhost:8080${endpoint}`);
    
    ws.onopen = () => {
      console.log(`Connected to ${endpoint}`);
      
      // Send authentication
      ws.send(JSON.stringify({
        type: "auth",
        access_token: token,
        data: {}
      }));
    };
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };
    
    ws.onerror = (error) => {
      console.error(`WebSocket error on ${endpoint}:`, error);
      onError(error);
    };
    
    ws.onclose = () => {
      console.log(`Disconnected from ${endpoint}`);
    };
    
    this.connections.set(endpoint, ws);
    return ws;
  }

  disconnect(endpoint) {
    const ws = this.connections.get(endpoint);
    if (ws) {
      ws.close();
      this.connections.delete(endpoint);
    }
  }

  disconnectAll() {
    this.connections.forEach(ws => ws.close());
    this.connections.clear();
  }
}

// Usage
const wsService = new WebSocketService();

// Connect to balance updates
wsService.connect('/ws/balance/player', userToken, (data) => {
  console.log('Balance update:', data);
  // Update UI with new balance
}, (error) => {
  console.error('Balance WS error:', error);
});
```

## Key Features

1. **Authentication Required:** All endpoints require JWT token
2. **Real-time Updates:** All provide real-time data streaming
3. **Connection Management:** Automatic cleanup on disconnect
4. **Error Handling:** Proper error responses for invalid tokens
5. **Multiple Connections:** Users can have multiple connections per endpoint
6. **Thread Safety:** Uses mutex locks for concurrent access

## Security Considerations

- Always validate JWT tokens on the frontend before sending
- Implement proper error handling for connection failures
- Use secure WebSocket connections (wss://) in production
- Implement reconnection logic for dropped connections
- Monitor connection status and handle authentication expiration

## Production Notes

- Replace `localhost:8080` with your production domain
- Use `wss://` instead of `ws://` for secure connections
- Implement proper logging and monitoring
- Consider connection pooling and rate limiting
- Test thoroughly with multiple concurrent connections