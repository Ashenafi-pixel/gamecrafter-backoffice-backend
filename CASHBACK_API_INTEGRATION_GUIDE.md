# 🎰 TucanBIT Cashback System API Integration Guide

## 📋 Overview

The TucanBIT Cashback System provides a comprehensive reward mechanism for players. This guide covers all API endpoints, request/response formats, authentication, and integration flows for frontend implementation.

## 🔐 Authentication

All cashback APIs require JWT authentication. Include the access token in the Authorization header:

```http
Authorization: Bearer <access_token>
```

### Getting Access Token

```http
POST /login
Content-Type: application/json

{
  "login_id": "user@example.com",
  "password": "user_password"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_profile": {
    "username": "player_username",
    "email": "user@example.com",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "type": "PLAYER"
  }
}
```

---

## 🎯 Cashback System Flow

### 1. Automatic Cashback Calculation
- Cashback is **automatically calculated** after each game result
- No frontend action required for earning cashback
- Players can view available cashback anytime

### 2. Manual Cashback Claiming
- Players must **manually claim** their accumulated cashback
- Cashback is credited to player's balance upon claiming
- Players can claim partial or full amounts

---

## 📊 API Endpoints

## 🎯 Player Profile APIs

### 1. Get User Profile

**Endpoint:** `GET /api/user/profile`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "username": "player_username",
  "phone_number": "+1234567890",
  "email": "user@example.com",
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "profile_picture": "https://example.com/profile.jpg",
  "first_name": "John",
  "last_name": "Doe",
  "type": "PLAYER",
  "referral_code": "REF123456",
  "refered_by_code": "REF789012",
  "referal_type": "PLAYER"
}
```

**Frontend Implementation:**
```javascript
// Fetch user profile
const getUserProfile = async () => {
  try {
    const response = await fetch('/api/user/profile', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display profile information
    displayUserProfile(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching user profile:', error);
  }
};

// Display user profile
const displayUserProfile = (profile) => {
  document.getElementById('username').textContent = profile.username;
  document.getElementById('email').textContent = profile.email;
  document.getElementById('phone').textContent = profile.phone_number;
  document.getElementById('full-name').textContent = `${profile.first_name} ${profile.last_name}`;
  document.getElementById('profile-picture').src = profile.profile_picture || '/default-avatar.png';
  document.getElementById('referral-code').textContent = profile.referral_code;
};
```

---

### 2. Get User Level Information

**Endpoint:** `GET /api/users/level`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "level": 2,
  "next_level": 3,
  "amount_spent_to_reach_level": "2141.5",
  "next_level_requirement": "5000",
  "bucks": "100.50",
  "is_final_level": false,
  "squad_id": "squad-uuid-here"
}
```

**Frontend Implementation:**
```javascript
// Fetch user level
const getUserLevel = async () => {
  try {
    const response = await fetch('/api/users/level', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display level information
    displayUserLevel(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching user level:', error);
  }
};

// Display user level
const displayUserLevel = (levelData) => {
  document.getElementById('current-level').textContent = levelData.level;
  document.getElementById('next-level').textContent = levelData.next_level;
  document.getElementById('amount-spent').textContent = `$${levelData.amount_spent_to_reach_level}`;
  document.getElementById('next-requirement').textContent = `$${levelData.next_level_requirement}`;
  document.getElementById('bucks').textContent = levelData.bucks;
  
  // Calculate progress percentage
  const progress = (parseFloat(levelData.amount_spent_to_reach_level) / parseFloat(levelData.next_level_requirement)) * 100;
  document.getElementById('level-progress').style.width = `${Math.min(progress, 100)}%`;
};
```

---

### 3. Get Level Progression Information

**Endpoint:** `GET /user/cashback/level-progression`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "current_level": 2,
  "current_tier": {
    "id": "302aef59-c303-4835-bfad-8f5ef4e03ffc",
    "tier_name": "Silver",
    "tier_level": 2,
    "min_expected_ggr_required": "1000",
    "cashback_percentage": "1",
    "bonus_multiplier": "1.1",
    "daily_cashback_limit": "100",
    "weekly_cashback_limit": null,
    "monthly_cashback_limit": null,
    "special_benefits": {},
    "is_active": true,
    "created_at": "2025-09-12T14:07:21.477098+03:00",
    "updated_at": "2025-09-12T14:07:21.477098+03:00"
  },
  "next_tier": {
    "id": "next-tier-uuid",
    "tier_name": "Gold",
    "tier_level": 3,
    "min_expected_ggr_required": "5000",
    "cashback_percentage": "2",
    "bonus_multiplier": "1.2",
    "daily_cashback_limit": "200",
    "weekly_cashback_limit": null,
    "monthly_cashback_limit": null,
    "special_benefits": {},
    "is_active": true,
    "created_at": "2025-09-12T14:07:21.477098+03:00",
    "updated_at": "2025-09-12T14:07:21.477098+03:00"
  },
  "total_expected_ggr": "2141.5",
  "progress_to_next": "0.4283",
  "expected_ggr_to_next_level": "2858.5",
  "last_level_up": "2025-09-15T10:30:00Z",
  "level_progress": "0.4283"
}
```

**Frontend Implementation:**
```javascript
// Fetch level progression
const getLevelProgression = async () => {
  try {
    const response = await fetch('/user/cashback/level-progression', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display level progression
    displayLevelProgression(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching level progression:', error);
  }
};

// Display level progression
const displayLevelProgression = (progression) => {
  document.getElementById('current-tier-name').textContent = progression.current_tier.tier_name;
  document.getElementById('current-tier-level').textContent = progression.current_level;
  document.getElementById('cashback-rate').textContent = `${progression.current_tier.cashback_percentage}%`;
  document.getElementById('bonus-multiplier').textContent = progression.current_tier.bonus_multiplier;
  
  if (progression.next_tier) {
    document.getElementById('next-tier-name').textContent = progression.next_tier.tier_name;
    document.getElementById('next-tier-level').textContent = progression.next_tier.tier_level;
    document.getElementById('next-tier-cashback').textContent = `${progression.next_tier.cashback_percentage}%`;
  }
  
  // Progress bar
  const progressPercent = parseFloat(progression.progress_to_next) * 100;
  document.getElementById('tier-progress-bar').style.width = `${progressPercent}%`;
  document.getElementById('progress-text').textContent = `${progressPercent.toFixed(1)}%`;
  
  // GGR information
  document.getElementById('current-ggr').textContent = `$${progression.total_expected_ggr}`;
  document.getElementById('ggr-to-next').textContent = `$${progression.expected_ggr_to_next_level}`;
};
```

---

### 4. Get User Analytics

**Endpoint:** `GET /analytics/users/{user_id}/analytics`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "total_deposits": "5000.00",
  "total_withdrawals": "2000.00",
  "total_bets": "15000.00",
  "total_wins": "12000.00",
  "total_bonuses": "500.00",
  "total_cashback": "150.00",
  "net_loss": "3000.00",
  "transaction_count": 150,
  "unique_games_played": 25,
  "session_count": 45,
  "avg_bet_amount": "100.00",
  "max_bet_amount": "500.00",
  "min_bet_amount": "10.00",
  "last_activity": "2025-09-29T14:30:00Z"
}
```

**Frontend Implementation:**
```javascript
// Fetch user analytics
const getUserAnalytics = async (userId) => {
  try {
    const response = await fetch(`/analytics/users/${userId}/analytics`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display user analytics
    displayUserAnalytics(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching user analytics:', error);
  }
};

// Display user analytics
const displayUserAnalytics = (analytics) => {
  document.getElementById('total-deposits').textContent = `$${analytics.total_deposits}`;
  document.getElementById('total-withdrawals').textContent = `$${analytics.total_withdrawals}`;
  document.getElementById('total-bets').textContent = `$${analytics.total_bets}`;
  document.getElementById('total-wins').textContent = `$${analytics.total_wins}`;
  document.getElementById('total-bonuses').textContent = `$${analytics.total_bonuses}`;
  document.getElementById('total-cashback').textContent = `$${analytics.total_cashback}`;
  document.getElementById('net-loss').textContent = `$${analytics.net_loss}`;
  document.getElementById('transaction-count').textContent = analytics.transaction_count;
  document.getElementById('games-played').textContent = analytics.unique_games_played;
  document.getElementById('session-count').textContent = analytics.session_count;
  document.getElementById('avg-bet').textContent = `$${analytics.avg_bet_amount}`;
  document.getElementById('max-bet').textContent = `$${analytics.max_bet_amount}`;
  document.getElementById('min-bet').textContent = `$${analytics.min_bet_amount}`;
  document.getElementById('last-activity').textContent = new Date(analytics.last_activity).toLocaleDateString();
};
```

---

### 5. Get User Balance History

**Endpoint:** `GET /analytics/users/{user_id}/balance-history`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `date_from` (optional): Start date (YYYY-MM-DD)
- `date_to` (optional): End date (YYYY-MM-DD)
- `limit` (optional): Number of records (default: 50)
- `offset` (optional): Offset for pagination (default: 0)

**Response:**
```json
{
  "balance_history": [
    {
      "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
      "balance": "112865.015",
      "currency": "USD",
      "snapshot_time": "2025-09-29T14:22:16Z",
      "transaction_id": "transaction-uuid",
      "transaction_type": "cashback_claim"
    }
  ],
  "meta": {
    "total": 150,
    "page": 1,
    "page_size": 50,
    "pages": 3
  }
}
```

**Frontend Implementation:**
```javascript
// Fetch balance history
const getBalanceHistory = async (userId, dateFrom = null, dateTo = null, page = 1) => {
  try {
    let url = `/analytics/users/${userId}/balance-history?page=${page}&limit=20`;
    
    if (dateFrom) url += `&date_from=${dateFrom}`;
    if (dateTo) url += `&date_to=${dateTo}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display balance history
    displayBalanceHistory(data.balance_history, data.meta);
    
    return data;
  } catch (error) {
    console.error('Error fetching balance history:', error);
  }
};

// Display balance history
const displayBalanceHistory = (history, meta) => {
  const container = document.getElementById('balance-history');
  container.innerHTML = '';
  
  history.forEach(entry => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${new Date(entry.snapshot_time).toLocaleDateString()}</td>
      <td>$${entry.balance}</td>
      <td>${entry.currency}</td>
      <td>${entry.transaction_type || 'N/A'}</td>
    `;
    container.appendChild(row);
  });
  
  // Update pagination
  updatePagination(meta);
};
```

---

### 6. Get User Referral Information

**Endpoint:** `GET /api/user/referral`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "referral_code": "REF123456",
  "referral_link": "https://tucanbit.com/register?ref=REF123456",
  "total_referrals": 5,
  "total_earnings": "250.00",
  "referral_multiplier": "1.5"
}
```

**Frontend Implementation:**
```javascript
// Fetch referral information
const getReferralInfo = async () => {
  try {
    const response = await fetch('/api/user/referral', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display referral information
    displayReferralInfo(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching referral info:', error);
  }
};

// Display referral information
const displayReferralInfo = (referral) => {
  document.getElementById('referral-code').textContent = referral.referral_code;
  document.getElementById('referral-link').value = referral.referral_link;
  document.getElementById('total-referrals').textContent = referral.total_referrals;
  document.getElementById('referral-earnings').textContent = `$${referral.total_earnings}`;
  document.getElementById('referral-multiplier').textContent = referral.referral_multiplier;
};
```

---

## 🎰 Cashback System APIs

### 1. Get User Cashback Summary

**Endpoint:** `GET /user/cashback`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
  "current_tier": {
    "id": "302aef59-c303-4835-bfad-8f5ef4e03ffc",
    "tier_name": "Silver",
    "tier_level": 2,
    "min_expected_ggr_required": "1000",
    "cashback_percentage": "1",
    "bonus_multiplier": "1.1",
    "daily_cashback_limit": "100",
    "weekly_cashback_limit": null,
    "monthly_cashback_limit": null,
    "special_benefits": {},
    "is_active": true,
    "created_at": "2025-09-12T14:07:21.477098+03:00",
    "updated_at": "2025-09-12T14:07:21.477098+03:00"
  },
  "level_progress": "0.01",
  "total_ggr": "2141.5",
  "available_cashback": "5.015",
  "pending_cashback": "0",
  "total_claimed": "0",
  "next_tier_ggr": "5000",
  "daily_limit": "100",
  "weekly_limit": null,
  "monthly_limit": null,
  "special_benefits": {}
}
```

**Frontend Implementation:**
```javascript
// Fetch user cashback summary
const getCashbackSummary = async () => {
  try {
    const response = await fetch('/user/cashback', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display cashback information
    displayCashbackInfo(data);
    
    return data;
  } catch (error) {
    console.error('Error fetching cashback summary:', error);
  }
};

// Display function
const displayCashbackInfo = (data) => {
  document.getElementById('current-tier').textContent = data.current_tier.tier_name;
  document.getElementById('cashback-rate').textContent = `${data.current_tier.cashback_percentage}%`;
  document.getElementById('available-cashback').textContent = `$${data.available_cashback}`;
  document.getElementById('total-ggr').textContent = `$${data.total_ggr}`;
  document.getElementById('next-tier-progress').textContent = `${data.level_progress}%`;
};
```

---

### 2. Claim Cashback

**Endpoint:** `POST /user/cashback/claim`

**Headers:**
```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "amount": "5.015"
}
```

**Response:**
```json
{
  "claim_id": "9acc22a4-5f72-4190-b3b9-9604088a0e7d",
  "amount": "5.015",
  "net_amount": "5.015",
  "processing_fee": "0",
  "status": "completed",
  "message": "Cashback claim processed successfully"
}
```

**Frontend Implementation:**
```javascript
// Claim cashback
const claimCashback = async (amount) => {
  try {
    const response = await fetch('/user/cashback/claim', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        amount: amount.toString()
      })
    });
    
    const data = await response.json();
    
    if (response.ok) {
      // Show success message
      showSuccessMessage(`Successfully claimed $${data.amount} cashback!`);
      
      // Refresh cashback summary
      await getCashbackSummary();
      
      // Refresh user balance
      await getUserBalance();
      
      return data;
    } else {
      throw new Error(data.error || 'Failed to claim cashback');
    }
  } catch (error) {
    console.error('Error claiming cashback:', error);
    showErrorMessage(error.message);
  }
};

// Claim button handler
const handleClaimCashback = () => {
  const availableAmount = document.getElementById('available-cashback').textContent.replace('$', '');
  
  if (parseFloat(availableAmount) > 0) {
    if (confirm(`Claim $${availableAmount} cashback?`)) {
      claimCashback(availableAmount);
    }
  } else {
    showErrorMessage('No available cashback to claim');
  }
};
```

---

### 3. Get Cashback Earnings History

**Endpoint:** `GET /user/cashback/earnings`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "earnings": [
    {
      "id": "a0daaebc-e98a-400d-9bfd-04b62da85445",
      "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
      "tier_id": "302aef59-c303-4835-bfad-8f5ef4e03ffc",
      "earning_type": "bet",
      "source_bet_id": null,
      "expected_ggr_amount": "0.5",
      "cashback_rate": "1",
      "earned_amount": "0.005",
      "claimed_amount": "0",
      "available_amount": "0.005",
      "status": "available",
      "expires_at": "2025-10-29T13:47:38.830517+03:00",
      "claimed_at": null,
      "created_at": "2025-09-29T13:47:38.838665+03:00",
      "updated_at": "2025-09-29T13:47:38.838665+03:00"
    }
  ],
  "pagination": {
    "limit": 20,
    "page": 1,
    "total": 1,
    "total_pages": 1
  }
}
```

**Frontend Implementation:**
```javascript
// Fetch cashback earnings history
const getCashbackEarnings = async (page = 1, limit = 20) => {
  try {
    const response = await fetch(`/user/cashback/earnings?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    // Display earnings history
    displayEarningsHistory(data.earnings);
    
    return data;
  } catch (error) {
    console.error('Error fetching cashback earnings:', error);
  }
};

// Display earnings history
const displayEarningsHistory = (earnings) => {
  const container = document.getElementById('earnings-history');
  container.innerHTML = '';
  
  earnings.forEach(earning => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${new Date(earning.created_at).toLocaleDateString()}</td>
      <td>$${earning.earned_amount}</td>
      <td>${earning.cashback_rate}%</td>
      <td>$${earning.available_amount}</td>
      <td>
        <span class="status-badge ${earning.status}">
          ${earning.status.charAt(0).toUpperCase() + earning.status.slice(1)}
        </span>
      </td>
      <td>${earning.expires_at ? new Date(earning.expires_at).toLocaleDateString() : 'N/A'}</td>
    `;
    container.appendChild(row);
  });
};
```

---

### 4. Get User Balance

**Endpoint:** `GET /api/balance`

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response:**
```json
[
  {
    "id": "b6a6b5ab-b150-45f3-8de7-8711be2a2f89",
    "user_id": "a5e168fb-168e-4183-84c5-d49038ce00b5",
    "currency_code": "USD",
    "amount_cents": 11286501,
    "amount_units": "112865.015",
    "reserved_cents": 0,
    "reserved_units": "0",
    "updated_at": "2025-09-29T14:22:16.635+03:00"
  }
]
```

**Frontend Implementation:**
```javascript
// Fetch user balance
const getUserBalance = async () => {
  try {
    const response = await fetch('/api/balance', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${access_token}`,
        'Content-Type': 'application/json'
      }
    });
    
    const data = await response.json();
    
    if (data.length > 0) {
      const balance = data[0];
      document.getElementById('user-balance').textContent = `$${balance.amount_units}`;
    }
    
    return data;
  } catch (error) {
    console.error('Error fetching balance:', error);
  }
};
```

---

## 🎮 Game Integration Flow

### Complete Cashback Flow Example

```javascript
// Complete cashback flow demonstration
const demonstrateCashbackFlow = async () => {
  try {
    // 1. Login
    const loginResponse = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        login_id: 'user@example.com',
        password: 'user_password'
      })
    });
    
    const loginData = await loginResponse.json();
    const accessToken = loginData.access_token;
    
    // 2. Check initial cashback status
    const initialSummary = await getCashbackSummary(accessToken);
    console.log('Initial cashback:', initialSummary.available_cashback);
    
    // 3. Launch game (GrooveTech integration)
    const gameResponse = await fetch('/api/groove/launch-game', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        game_id: '82695',
        device_type: 'desktop',
        game_mode: 'real',
        country: 'US',
        currency: 'USD',
        language: 'en_US',
        is_test_account: false
      })
    });
    
    const gameData = await gameResponse.json();
    const sessionId = gameData.session_id;
    
    // 4. Place wager (example)
    const wagerResponse = await fetch(`/groove-official?request=wager&accountid=${userId}&gamesessionid=${sessionId}&device=desktop&gameid=82695&apiversion=1.2&betamount=25.0&roundid=round_${Date.now()}&transactionid=wager_${Date.now()}`);
    
    // 5. Place result (example)
    const resultResponse = await fetch(`/groove-official?request=result&accountid=${userId}&gamesessionid=${sessionId}&device=desktop&gameid=82695&apiversion=1.2&result=15.0&roundid=round_${Date.now()}&transactionid=result_${Date.now()}&gamestatus=completed`);
    
    // 6. Check updated cashback (automatically calculated)
    setTimeout(async () => {
      const updatedSummary = await getCashbackSummary(accessToken);
      console.log('Updated cashback:', updatedSummary.available_cashback);
      
      // 7. Claim cashback if available
      if (parseFloat(updatedSummary.available_cashback) > 0) {
        const claimResponse = await claimCashback(accessToken, updatedSummary.available_cashback);
        console.log('Claim successful:', claimResponse);
      }
    }, 2000);
    
  } catch (error) {
    console.error('Cashback flow error:', error);
  }
};
```

---

## 🎨 Frontend UI Components

### Complete Player Profile Dashboard Component

```html
<!DOCTYPE html>
<html>
<head>
    <title>TucanBIT Player Profile Dashboard</title>
    <style>
        .profile-dashboard {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            font-family: Arial, sans-serif;
            background: #f8f9fa;
        }
        
        .profile-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 15px;
            margin-bottom: 30px;
            display: flex;
            align-items: center;
            gap: 20px;
        }
        
        .profile-avatar {
            width: 80px;
            height: 80px;
            border-radius: 50%;
            border: 3px solid white;
            object-fit: cover;
        }
        
        .profile-info h1 {
            margin: 0 0 10px 0;
            font-size: 28px;
        }
        
        .profile-info p {
            margin: 5px 0;
            opacity: 0.9;
        }
        
        .dashboard-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .dashboard-card {
            background: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        
        .card-title {
            font-size: 18px;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 20px;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        
        .level-progression {
            text-align: center;
            margin-bottom: 20px;
        }
        
        .current-tier {
            font-size: 24px;
            font-weight: bold;
            color: #e74c3c;
            margin-bottom: 10px;
        }
        
        .tier-progress {
            width: 100%;
            height: 25px;
            background: #ecf0f1;
            border-radius: 12px;
            overflow: hidden;
            margin: 15px 0;
        }
        
        .tier-progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #3498db, #2ecc71);
            transition: width 0.3s ease;
            border-radius: 12px;
        }
        
        .progress-text {
            font-size: 14px;
            color: #7f8c8d;
            margin-top: 5px;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        
        .stat-item {
            text-align: center;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        
        .stat-value {
            font-size: 20px;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .stat-label {
            font-size: 12px;
            color: #7f8c8d;
            margin-top: 5px;
        }
        
        .cashback-section {
            background: linear-gradient(135deg, #27ae60, #2ecc71);
            color: white;
        }
        
        .cashback-section .card-title {
            color: white;
            border-bottom-color: white;
        }
        
        .claim-button {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: bold;
            width: 100%;
            margin-top: 15px;
        }
        
        .claim-button:disabled {
            background: #bdc3c7;
            cursor: not-allowed;
        }
        
        .referral-section {
            background: linear-gradient(135deg, #9b59b6, #8e44ad);
            color: white;
        }
        
        .referral-section .card-title {
            color: white;
            border-bottom-color: white;
        }
        
        .referral-link {
            background: rgba(255,255,255,0.2);
            border: 1px solid rgba(255,255,255,0.3);
            color: white;
            padding: 10px;
            border-radius: 5px;
            width: 100%;
            margin: 10px 0;
        }
        
        .copy-button {
            background: rgba(255,255,255,0.2);
            color: white;
            border: 1px solid rgba(255,255,255,0.3);
            padding: 8px 15px;
            border-radius: 5px;
            cursor: pointer;
            margin-left: 10px;
        }
        
        .analytics-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        
        .analytics-table th,
        .analytics-table td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #dee2e6;
        }
        
        .analytics-table th {
            background: #f8f9fa;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .balance-history {
            max-height: 300px;
            overflow-y: auto;
        }
        
        .loading {
            text-align: center;
            padding: 20px;
            color: #7f8c8d;
        }
        
        .error {
            background: #e74c3c;
            color: white;
            padding: 15px;
            border-radius: 8px;
            margin: 10px 0;
        }
        
        .success {
            background: #27ae60;
            color: white;
            padding: 15px;
            border-radius: 8px;
            margin: 10px 0;
        }
        
        @media (max-width: 768px) {
            .profile-header {
                flex-direction: column;
                text-align: center;
            }
            
            .dashboard-grid {
                grid-template-columns: 1fr;
            }
            
            .stats-grid {
                grid-template-columns: repeat(2, 1fr);
            }
        }
    </style>
</head>
<body>
    <div class="profile-dashboard">
        <!-- Profile Header -->
        <div class="profile-header">
            <img id="profile-avatar" class="profile-avatar" src="/default-avatar.png" alt="Profile Picture">
            <div class="profile-info">
                <h1 id="full-name">Loading...</h1>
                <p id="username">@username</p>
                <p id="email">user@example.com</p>
                <p id="user-type">PLAYER</p>
            </div>
        </div>
        
        <!-- Dashboard Grid -->
        <div class="dashboard-grid">
            <!-- Level Progression Card -->
            <div class="dashboard-card">
                <div class="card-title">🎯 Level Progression</div>
                <div class="level-progression">
                    <div class="current-tier" id="current-tier-name">Silver</div>
                    <div>Level <span id="current-tier-level">2</span></div>
                    <div class="tier-progress">
                        <div class="tier-progress-fill" id="tier-progress-bar" style="width: 42.8%"></div>
                    </div>
                    <div class="progress-text" id="progress-text">42.8% to next level</div>
                    <div style="margin-top: 15px;">
                        <div>Next: <span id="next-tier-name">Gold</span> (Level <span id="next-tier-level">3</span>)</div>
                        <div style="font-size: 14px; color: #7f8c8d;">
                            Need $<span id="ggr-to-next">2,858.50</span> more GGR
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Cashback Summary Card -->
            <div class="dashboard-card cashback-section">
                <div class="card-title">💰 Cashback Summary</div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value" id="available-cashback">$5.015</div>
                        <div class="stat-label">Available</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value" id="total-claimed">$0</div>
                        <div class="stat-label">Claimed</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value" id="cashback-rate">1%</div>
                        <div class="stat-label">Rate</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value" id="daily-limit">$100</div>
                        <div class="stat-label">Daily Limit</div>
                    </div>
                </div>
                <button class="claim-button" id="claim-button" onclick="handleClaimCashback()">
                    Claim Available Cashback
                </button>
            </div>
            
            <!-- User Analytics Card -->
            <div class="dashboard-card">
                <div class="card-title">📊 Gaming Statistics</div>
                <table class="analytics-table">
                    <tr>
                        <td><strong>Total Deposits</strong></td>
                        <td id="total-deposits">$5,000.00</td>
                    </tr>
                    <tr>
                        <td><strong>Total Withdrawals</strong></td>
                        <td id="total-withdrawals">$2,000.00</td>
                    </tr>
                    <tr>
                        <td><strong>Total Bets</strong></td>
                        <td id="total-bets">$15,000.00</td>
                    </tr>
                    <tr>
                        <td><strong>Total Wins</strong></td>
                        <td id="total-wins">$12,000.00</td>
                    </tr>
                    <tr>
                        <td><strong>Net Loss</strong></td>
                        <td id="net-loss">$3,000.00</td>
                    </tr>
                    <tr>
                        <td><strong>Games Played</strong></td>
                        <td id="games-played">25</td>
                    </tr>
                    <tr>
                        <td><strong>Avg Bet</strong></td>
                        <td id="avg-bet">$100.00</td>
                    </tr>
                    <tr>
                        <td><strong>Last Activity</strong></td>
                        <td id="last-activity">2025-09-29</td>
                    </tr>
                </table>
            </div>
            
            <!-- Referral Information Card -->
            <div class="dashboard-card referral-section">
                <div class="card-title">👥 Referral Program</div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value" id="total-referrals">5</div>
                        <div class="stat-label">Referrals</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value" id="referral-earnings">$250.00</div>
                        <div class="stat-label">Earnings</div>
                    </div>
                </div>
                <div style="margin-top: 15px;">
                    <label><strong>Your Referral Code:</strong></label>
                    <div style="display: flex; align-items: center; margin-top: 5px;">
                        <input type="text" class="referral-link" id="referral-link" readonly value="REF123456">
                        <button class="copy-button" onclick="copyReferralLink()">Copy</button>
                    </div>
                </div>
            </div>
            
            <!-- Balance Information Card -->
            <div class="dashboard-card">
                <div class="card-title">💳 Account Balance</div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value" id="user-balance">$112,865.015</div>
                        <div class="stat-label">Current Balance</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value" id="bucks">100.50</div>
                        <div class="stat-label">Bucks</div>
                    </div>
                </div>
                <div style="margin-top: 15px;">
                    <button onclick="getBalanceHistory()" style="background: #3498db; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer;">
                        View Balance History
                    </button>
                </div>
            </div>
            
            <!-- Balance History Card -->
            <div class="dashboard-card">
                <div class="card-title">📈 Balance History</div>
                <div class="balance-history">
                    <table class="analytics-table">
                        <thead>
                            <tr>
                                <th>Date</th>
                                <th>Balance</th>
                                <th>Currency</th>
                                <th>Type</th>
                            </tr>
                        </thead>
                        <tbody id="balance-history">
                            <tr>
                                <td colspan="4" class="loading">Loading balance history...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- Messages -->
        <div id="messages"></div>
    </div>

    <script>
        // Global variables
        let accessToken = '';
        let currentUserId = '';
        
        // Initialize dashboard
        document.addEventListener('DOMContentLoaded', async () => {
            // Get access token from localStorage or login
            accessToken = localStorage.getItem('access_token');
            if (!accessToken) {
                showError('Please login first');
                return;
            }
            
            // Load all dashboard data
            await loadDashboardData();
        });
        
        // Load all dashboard data
        const loadDashboardData = async () => {
            try {
                showLoading('Loading dashboard data...');
                
                // Load user profile first to get user ID
                const profile = await getUserProfile();
                if (profile) {
                    currentUserId = profile.user_id;
                    
                    // Load all other data in parallel
                    await Promise.all([
                        getUserLevel(),
                        getLevelProgression(),
                        getCashbackSummary(),
                        getUserAnalytics(currentUserId),
                        getReferralInfo(),
                        getUserBalance(),
                        getBalanceHistory(currentUserId)
                    ]);
                }
                
                hideLoading();
            } catch (error) {
                console.error('Error loading dashboard:', error);
                showError('Failed to load dashboard data');
            }
        };
        
        // Include all the API functions from the examples above
        // ... (all the functions from the previous examples)
        
        // Utility functions
        const showLoading = (message) => {
            const messagesDiv = document.getElementById('messages');
            messagesDiv.innerHTML = `<div class="loading">${message}</div>`;
        };
        
        const hideLoading = () => {
            const messagesDiv = document.getElementById('messages');
            messagesDiv.innerHTML = '';
        };
        
        const showError = (message) => {
            const messagesDiv = document.getElementById('messages');
            messagesDiv.innerHTML = `<div class="error">${message}</div>`;
        };
        
        const showSuccess = (message) => {
            const messagesDiv = document.getElementById('messages');
            messagesDiv.innerHTML = `<div class="success">${message}</div>`;
            setTimeout(() => hideLoading(), 3000);
        };
        
        const copyReferralLink = () => {
            const referralLink = document.getElementById('referral-link');
            referralLink.select();
            document.execCommand('copy');
            showSuccess('Referral link copied to clipboard!');
        };
        
        // Refresh dashboard data
        const refreshDashboard = () => {
            loadDashboardData();
        };
        
        // Auto-refresh every 5 minutes
        setInterval(refreshDashboard, 300000);
    </script>
</body>
</html>
```

### Cashback Dashboard Component

```html
<!DOCTYPE html>
<html>
<head>
    <title>TucanBIT Cashback Dashboard</title>
    <style>
        .cashback-dashboard {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            font-family: Arial, sans-serif;
        }
        
        .tier-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        
        .stat-value {
            font-size: 24px;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .stat-label {
            color: #7f8c8d;
            margin-top: 5px;
        }
        
        .claim-button {
            background: #27ae60;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 16px;
            margin: 10px 0;
        }
        
        .claim-button:disabled {
            background: #bdc3c7;
            cursor: not-allowed;
        }
        
        .status-badge {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
        }
        
        .status-badge.available {
            background: #d5f4e6;
            color: #27ae60;
        }
        
        .status-badge.claimed {
            background: #e8f4fd;
            color: #3498db;
        }
        
        .progress-bar {
            width: 100%;
            height: 20px;
            background: #ecf0f1;
            border-radius: 10px;
            overflow: hidden;
            margin: 10px 0;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #3498db, #2ecc71);
            transition: width 0.3s ease;
        }
    </style>
</head>
<body>
    <div class="cashback-dashboard">
        <h1>🎰 TucanBIT Cashback Dashboard</h1>
        
        <!-- Current Tier Card -->
        <div class="tier-card">
            <h2 id="current-tier">Silver</h2>
            <p>Cashback Rate: <span id="cashback-rate">1%</span></p>
            <div class="progress-bar">
                <div class="progress-fill" id="tier-progress" style="width: 1%"></div>
            </div>
            <p>Progress to next tier: <span id="next-tier-progress">1%</span></p>
        </div>
        
        <!-- Stats Grid -->
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value" id="available-cashback">$5.015</div>
                <div class="stat-label">Available Cashback</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="total-ggr">$2,141.5</div>
                <div class="stat-label">Total GGR</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="total-claimed">$0</div>
                <div class="stat-label">Total Claimed</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="daily-limit">$100</div>
                <div class="stat-label">Daily Limit</div>
            </div>
        </div>
        
        <!-- Claim Button -->
        <button class="claim-button" id="claim-button" onclick="handleClaimCashback()">
            Claim Available Cashback
        </button>
        
        <!-- Earnings History -->
        <h3>Cashback Earnings History</h3>
        <table style="width: 100%; border-collapse: collapse;">
            <thead>
                <tr style="background: #f8f9fa;">
                    <th style="padding: 10px; text-align: left;">Date</th>
                    <th style="padding: 10px; text-align: left;">Earned</th>
                    <th style="padding: 10px; text-align: left;">Rate</th>
                    <th style="padding: 10px; text-align: left;">Available</th>
                    <th style="padding: 10px; text-align: left;">Status</th>
                    <th style="padding: 10px; text-align: left;">Expires</th>
                </tr>
            </thead>
            <tbody id="earnings-history">
                <!-- Dynamic content -->
            </tbody>
        </table>
        
        <!-- User Balance -->
        <div style="margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 8px;">
            <h3>Current Balance</h3>
            <div class="stat-value" id="user-balance">$112,865.015</div>
        </div>
    </div>

    <script>
        // Include all the JavaScript functions from above
        // ... (all the functions from the examples above)
        
        // Initialize dashboard on page load
        document.addEventListener('DOMContentLoaded', async () => {
            await getCashbackSummary();
            await getCashbackEarnings();
            await getUserBalance();
        });
    </script>
</body>
</html>
```

---

## 🔧 Error Handling

### Common Error Responses

```javascript
// Error handling utility
const handleApiError = (response, data) => {
  switch (response.status) {
    case 400:
      throw new Error(data.error || 'Invalid request');
    case 401:
      throw new Error('Authentication required');
    case 403:
      throw new Error('Insufficient permissions');
    case 404:
      throw new Error('Resource not found');
    case 500:
      throw new Error('Server error');
    default:
      throw new Error(data.error || 'Unknown error');
  }
};

// Enhanced API calls with error handling
const apiCall = async (url, options = {}) => {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      handleApiError(response, data);
    }
    
    return data;
  } catch (error) {
    console.error('API call failed:', error);
    throw error;
  }
};
```

---

## 📱 Mobile Integration

### React Native Example

```javascript
// React Native cashback integration
import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';

const CashbackDashboard = ({ accessToken }) => {
  const [cashbackData, setCashbackData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCashbackSummary();
  }, []);

  const fetchCashbackSummary = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/cashback', {
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        }
      });
      
      const data = await response.json();
      setCashbackData(data);
    } catch (error) {
      Alert.alert('Error', 'Failed to fetch cashback data');
    } finally {
      setLoading(false);
    }
  };

  const claimCashback = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/cashback/claim', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          amount: cashbackData.available_cashback
        })
      });
      
      const data = await response.json();
      
      if (response.ok) {
        Alert.alert('Success', `Claimed $${data.amount} cashback!`);
        fetchCashbackSummary(); // Refresh data
      } else {
        Alert.alert('Error', data.error || 'Failed to claim cashback');
      }
    } catch (error) {
      Alert.alert('Error', 'Network error');
    }
  };

  if (loading) {
    return <Text>Loading...</Text>;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Cashback Dashboard</Text>
      
      <View style={styles.tierCard}>
        <Text style={styles.tierName}>{cashbackData.current_tier.tier_name}</Text>
        <Text style={styles.cashbackRate}>{cashbackData.current_tier.cashback_percentage}% Cashback</Text>
      </View>
      
      <View style={styles.statsContainer}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>${cashbackData.available_cashback}</Text>
          <Text style={styles.statLabel}>Available</Text>
        </View>
        
        <View style={styles.statItem}>
          <Text style={styles.statValue}>${cashbackData.total_ggr}</Text>
          <Text style={styles.statLabel}>Total GGR</Text>
        </View>
      </View>
      
      <TouchableOpacity 
        style={[
          styles.claimButton,
          parseFloat(cashbackData.available_cashback) === 0 && styles.disabledButton
        ]}
        onPress={claimCashback}
        disabled={parseFloat(cashbackData.available_cashback) === 0}
      >
        <Text style={styles.claimButtonText}>Claim Cashback</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
    backgroundColor: '#f5f5f5'
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 20,
    textAlign: 'center'
  },
  tierCard: {
    backgroundColor: '#667eea',
    padding: 20,
    borderRadius: 10,
    marginBottom: 20
  },
  tierName: {
    color: 'white',
    fontSize: 20,
    fontWeight: 'bold'
  },
  cashbackRate: {
    color: 'white',
    fontSize: 16,
    marginTop: 5
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginBottom: 20
  },
  statItem: {
    alignItems: 'center'
  },
  statValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#2c3e50'
  },
  statLabel: {
    color: '#7f8c8d',
    marginTop: 5
  },
  claimButton: {
    backgroundColor: '#27ae60',
    padding: 15,
    borderRadius: 8,
    alignItems: 'center'
  },
  disabledButton: {
    backgroundColor: '#bdc3c7'
  },
  claimButtonText: {
    color: 'white',
    fontSize: 16,
    fontWeight: 'bold'
  }
});

export default CashbackDashboard;
```

---

## 📋 Complete API Integration Summary

### 🎯 Player Profile APIs (6 endpoints)
1. **GET /api/user/profile** - User profile information
2. **GET /api/users/level** - User level and progression
3. **GET /user/cashback/level-progression** - Detailed level progression
4. **GET /analytics/users/{user_id}/analytics** - User gaming statistics
5. **GET /analytics/users/{user_id}/balance-history** - Balance history
6. **GET /api/user/referral** - Referral information

### 🎰 Cashback System APIs (4 endpoints)
1. **GET /user/cashback** - Cashback summary
2. **POST /user/cashback/claim** - Claim cashback
3. **GET /user/cashback/earnings** - Cashback earnings history
4. **GET /user/cashback/claims** - Cashback claims history

### 💳 Balance & Transaction APIs (2 endpoints)
1. **GET /api/balance** - Current user balance
2. **GET /analytics/users/{user_id}/transactions** - Transaction history

### 🔐 Authentication APIs (2 endpoints)
1. **POST /login** - User login
2. **POST /register** - User registration

---

## 🚀 Deployment Checklist

### Frontend Integration Checklist

#### 🔐 Authentication & Security
- [ ] **JWT Token Management**: Implement secure token storage and refresh
- [ ] **Login Flow**: Complete user authentication with profile loading
- [ ] **Token Validation**: Handle token expiration and refresh
- [ ] **Input Validation**: Validate all user inputs on frontend
- [ ] **Error Handling**: Comprehensive error handling for all API calls

#### 📊 Player Profile Integration
- [ ] **User Profile Display**: Show user information, avatar, and basic details
- [ ] **Level Progression**: Display current level, tier, and progress to next level
- [ ] **Gaming Statistics**: Show total deposits, withdrawals, bets, wins, and net loss
- [ ] **Balance Information**: Display current balance and bucks
- [ ] **Referral System**: Show referral code, link, and earnings
- [ ] **Balance History**: Display transaction history with pagination

#### 🎰 Cashback System Integration
- [ ] **Cashback Summary**: Show available, claimed, and pending cashback
- [ ] **Level Progression**: Display tier information and cashback rates
- [ ] **Claim Functionality**: Implement cashback claiming with validation
- [ ] **Earnings History**: Show cashback earnings with status and expiration
- [ ] **Claims History**: Display past cashback claims

#### 🎨 UI Components
- [ ] **Profile Dashboard**: Complete player profile dashboard
- [ ] **Cashback Dashboard**: Dedicated cashback management interface
- [ ] **Level Progression UI**: Visual progress bars and tier displays
- [ ] **Responsive Design**: Mobile-optimized layouts
- [ ] **Loading States**: Proper loading indicators for all API calls
- [ ] **Error Messages**: User-friendly error handling

#### 🔄 Real-time Features
- [ ] **Auto-refresh**: Periodic data updates (every 5 minutes)
- [ ] **Balance Updates**: Real-time balance updates after transactions
- [ ] **Cashback Updates**: Live cashback status updates
- [ ] **Level Progression**: Real-time level and tier updates

#### 📱 Mobile Support
- [ ] **React Native Integration**: Mobile app implementation
- [ ] **Responsive Design**: Tablet and mobile layouts
- [ ] **Touch Interactions**: Mobile-friendly buttons and gestures
- [ ] **Offline Handling**: Graceful offline/online transitions

#### 🧪 Testing & Quality
- [ ] **API Testing**: Test all endpoints with various scenarios
- [ ] **Error Scenarios**: Test network failures, invalid tokens, etc.
- [ ] **User Flows**: Complete user journey testing
- [ ] **Performance Testing**: Optimize API calls and caching
- [ ] **Cross-browser Testing**: Ensure compatibility across browsers

#### 🚀 Performance & Optimization
- [ ] **API Caching**: Implement appropriate caching strategies
- [ ] **Lazy Loading**: Load data as needed
- [ ] **Bundle Optimization**: Minimize JavaScript bundle size
- [ ] **Image Optimization**: Optimize profile pictures and assets
- [ ] **CDN Integration**: Use CDN for static assets

#### 📚 Documentation & Maintenance
- [ ] **API Documentation**: Document all custom implementations
- [ ] **Code Comments**: Add comprehensive code documentation
- [ ] **Error Logging**: Implement proper error logging
- [ ] **Analytics**: Track user interactions and API usage
- [ ] **Monitoring**: Set up API monitoring and alerts

### Testing Scenarios

1. **Login Flow**: Test authentication and token management
2. **Cashback Summary**: Verify correct data display
3. **Claim Process**: Test partial and full cashback claims
4. **Error Handling**: Test various error scenarios
5. **Balance Updates**: Verify balance updates after claims
6. **Mobile Experience**: Test on various devices
7. **Network Issues**: Test offline/online scenarios

---

## 📞 Support

For technical support or questions about the cashback API integration:

- **Email**: support@tucanbit.com
- **Documentation**: This guide and inline code comments
- **API Testing**: Use the provided Postman collection
- **Logs**: Check application logs for debugging

---

## 🔄 Version History

- **v1.0** (2025-09-29): Initial cashback system implementation
- **v1.1** (2025-09-29): Fixed claim validation and storage methods
- **v1.2** (2025-09-29): Added comprehensive frontend integration guide

---

*This documentation is maintained by the TucanBIT development team. Last updated: September 29, 2025*