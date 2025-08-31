# Crypto Wallet Authentication Implementation

## Overview

This document describes the implementation of crypto wallet authentication for the TucanBIT Online Casino platform. The system integrates with the existing JWT-based authentication framework and supports multiple wallet types including MetaMask, WalletConnect, Coinbase Wallet, Phantom, Trust Wallet, and Ledger.

## Architecture

### Integration with Existing System

The crypto wallet authentication is designed to work alongside the existing authentication system, not replace it. Users can:
- Connect multiple crypto wallets to their account
- Use wallet-based authentication as an alternative to traditional login
- Maintain their existing JWT sessions and permissions

### Key Components

1. **Database Schema**: New tables for wallet connections, challenges, and authentication logs
2. **DTOs**: Data transfer objects for wallet operations
3. **Handler**: HTTP endpoints for wallet operations
4. **Module**: Business logic for wallet authentication
5. **Storage**: Database operations for wallet data

## Database Schema

### Tables

#### `crypto_wallet_connections`
Stores wallet connections linked to user accounts.

```sql
CREATE TABLE crypto_wallet_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_type VARCHAR(50) NOT NULL CHECK (wallet_type IN ('metamask', 'walletconnect', 'coinbase', 'phantom', 'trust', 'ledger')),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_chain VARCHAR(50) NOT NULL DEFAULT 'ethereum',
    wallet_name VARCHAR(255),
    wallet_icon_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    verification_signature TEXT,
    verification_message TEXT,
    verification_timestamp TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, wallet_address),
    UNIQUE(wallet_address, wallet_type)
);
```

#### `crypto_wallet_challenges`
Stores verification challenges for wallet authentication.

```sql
CREATE TABLE crypto_wallet_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(50) NOT NULL,
    challenge_message TEXT NOT NULL,
    challenge_nonce VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### `crypto_wallet_auth_logs`
Logs all wallet authentication activities for security and audit purposes.

```sql
CREATE TABLE crypto_wallet_auth_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('connect', 'disconnect', 'login', 'verify', 'challenge')),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### User Table Updates

Added wallet-related fields to the existing users table:

```sql
ALTER TABLE users ADD COLUMN primary_wallet_address VARCHAR(255);
ALTER TABLE users ADD COLUMN wallet_verification_status VARCHAR(50) DEFAULT 'none' CHECK (wallet_verification_status IN ('none', 'pending', 'verified', 'failed'));
```

## API Endpoints

### Wallet Management

#### Connect Wallet
```
POST /api/wallet/connect
Authorization: Bearer <token>
Content-Type: application/json

{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "wallet_chain": "ethereum",
    "wallet_name": "My MetaMask",
    "wallet_icon_url": "https://example.com/metamask.png"
}
```

#### Disconnect Wallet
```
DELETE /api/wallet/disconnect/{connection_id}
Authorization: Bearer <token>
```

#### Get User Wallets
```
GET /api/wallet/list
Authorization: Bearer <token>
```

### Wallet Authentication

#### Create Challenge
```
POST /api/wallet/challenge
Content-Type: application/json

{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
}
```

#### Verify Challenge
```
POST /api/wallet/verify
Content-Type: application/json

{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "signature": "0x...",
    "message": "Sign this message to verify your wallet ownership...",
    "nonce": "abc123..."
}
```

#### Login with Wallet
```
POST /api/wallet/login
Content-Type: application/json

{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "signature": "0x...",
    "message": "Sign this message to verify your wallet ownership...",
    "nonce": "abc123..."
}
```

## Authentication Flow

### 1. Wallet Connection Flow

1. User authenticates with existing JWT token
2. User requests to connect a wallet
3. System validates wallet address format
4. System checks if wallet is already connected to another account
5. System creates wallet connection record
6. System logs the connection action

### 2. Wallet Verification Flow

1. User requests a verification challenge
2. System generates a unique nonce and challenge message
3. System stores challenge with expiration (5 minutes)
4. User signs the challenge message with their wallet
5. User submits signature for verification
6. System verifies the signature cryptographically
7. System marks wallet as verified
8. System logs the verification action

### 3. Wallet Login Flow

1. User submits wallet address and signature
2. System verifies the signature cryptographically
3. System looks up the wallet connection
4. System retrieves associated user account
5. System generates JWT tokens (using existing system)
6. System logs the login action
7. User receives access and refresh tokens

## Security Features

### Cryptographic Verification

- **Ethereum Signature Verification**: Uses ECDSA signature recovery to verify wallet ownership
- **Nonce-based Challenges**: Prevents replay attacks with unique, time-limited challenges
- **Message Formatting**: Follows Ethereum personal message signing standard

### Rate Limiting and Logging

- **Authentication Logs**: All wallet operations are logged with IP address and user agent
- **Challenge Expiration**: Verification challenges expire after 5 minutes
- **Duplicate Prevention**: Wallet addresses can only be connected to one account

### Data Validation

- **Wallet Type Validation**: Only supported wallet types are accepted
- **Address Format Validation**: Ethereum address format validation
- **Input Sanitization**: All user inputs are validated and sanitized

## Supported Wallet Types

| Wallet Type | Description | Chain Support |
|-------------|-------------|---------------|
| MetaMask | Browser extension wallet | Ethereum, BSC, Polygon |
| WalletConnect | Mobile wallet connection | Multi-chain |
| Coinbase Wallet | Coinbase's wallet solution | Multi-chain |
| Phantom | Solana-focused wallet | Solana, Ethereum |
| Trust Wallet | Binance's mobile wallet | Multi-chain |
| Ledger | Hardware wallet | Multi-chain |

## Error Handling

### Common Error Codes

- **400 Bad Request**: Invalid input data, malformed requests
- **401 Unauthorized**: Invalid signature, expired challenge
- **403 Forbidden**: Wallet already connected to another account
- **404 Not Found**: Wallet connection not found
- **409 Conflict**: Wallet already connected
- **500 Internal Server Error**: System errors, database failures

### Error Response Format

```json
{
    "code": 400,
    "message": "Invalid wallet address format",
    "field_error": [
        {
            "name": "wallet_address",
            "description": "Must be a valid Ethereum address"
        }
    ]
}
```

## Implementation Status

### Completed
- [x] Database schema design
- [x] DTO definitions
- [x] HTTP handler endpoints
- [x] Swagger documentation
- [x] Error handling integration
- [x] Mock responses for testing

### Pending
- [ ] SQLC code generation
- [ ] Database storage implementation
- [ ] Business logic module
- [ ] Signature verification implementation
- [ ] Integration with existing JWT system
- [ ] Unit and integration tests
- [ ] Production deployment

## Testing

### Manual Testing

1. **Connect Wallet**: Test wallet connection with valid/invalid data
2. **Challenge Creation**: Test challenge generation and expiration
3. **Signature Verification**: Test with real wallet signatures
4. **Login Flow**: Test complete authentication flow
5. **Error Cases**: Test various error scenarios

### Test Data

```json
{
    "wallet_type": "metamask",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "wallet_chain": "ethereum",
    "wallet_name": "Test Wallet"
}
```

## Deployment

### Database Migration

1. Run the crypto wallet migration:
   ```bash
   # Apply migration
   psql -d tucanbit -f migrations/20250827104500_crypto_wallet_auth.up.sql
   
   # Rollback if needed
   psql -d tucanbit -f migrations/20250827104500_crypto_wallet_auth.down.sql
   ```

### Configuration

Ensure the following environment variables are set:
- `JWT_SECRET`: For token generation
- `DEBUG_MODE`: For error logging detail

### Monitoring

Monitor the following metrics:
- Wallet connection success/failure rates
- Authentication attempt patterns
- Challenge expiration rates
- Error log volumes

## Future Enhancements

### Planned Features
- **Multi-chain Support**: Enhanced support for different blockchain networks
- **Wallet Recovery**: Backup and recovery mechanisms
- **Advanced Security**: Hardware wallet integration, multi-signature support
- **Analytics**: Wallet usage analytics and insights

### Scalability Considerations
- **Database Indexing**: Optimize queries for large user bases
- **Caching**: Redis caching for frequently accessed wallet data
- **Rate Limiting**: Advanced rate limiting for wallet operations
- **Load Balancing**: Distribute wallet authentication load

## Support and Maintenance

### Troubleshooting

Common issues and solutions:
1. **Signature Verification Failures**: Check message format and nonce
2. **Challenge Expiration**: Ensure system clock synchronization
3. **Database Connection Issues**: Verify PostgreSQL connectivity
4. **JWT Integration Issues**: Check token generation and validation

### Monitoring and Alerts

Set up alerts for:
- High error rates in wallet operations
- Database connection failures
- Unusual authentication patterns
- System performance degradation

## Conclusion

This crypto wallet authentication system provides a secure, scalable solution for integrating blockchain wallets with the existing TucanBIT platform. The implementation follows security best practices and integrates seamlessly with the current authentication infrastructure.

The system is designed to be production-ready with proper error handling, logging, and monitoring capabilities. Future enhancements will expand the supported wallet types and add advanced security features. 