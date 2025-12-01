# Seed Data Summary

This document describes the essential data seeded by the `20251201213100_seed_essential_data` migration.

## Overview

The seed migration includes all essential configuration data and the admin user (`kirubel_gizaw32`) needed to start the application.

## Seeded Data

### 1. Brands (4 records)
- TucanBIT (Main)
- TucanBIT Production
- TucanBIT Staging
- TucanBIT Development

### 2. Currency Configuration (9 currencies)
- **Crypto**: BNB, BTC, ETH, LTC, MATIC, SOL, USDC, USDT
- **Fiat**: USD

Each currency includes:
- Currency name
- Currency type (crypto/fiat)
- Decimal places
- Smallest unit name

### 3. Supported Chains (7 chains)
- **Active Mainnets**: 
  - Bitcoin Mainnet (BTC)
  - Ethereum Mainnet (ETH)
  - Solana Mainnet (SOL)
- **Inactive Mainnets**:
  - Binance Smart Chain (BNB)
  - Polygon Mainnet (MATIC)
- **Testnets**:
  - Ethereum Sepolia Testnet
  - Solana Testnet

### 4. Roles (3 roles)
- `super` - Super admin role
- `admin role` - Admin role
- `manager` - Manager role

### 5. Admin Activity Categories (8 categories)
- user_management
- financial
- security
- system
- withdrawal
- game_management
- reports
- notifications

### 6. System Configuration (6 settings)
- `cumulative_kyc_transaction_limit` - KYC transaction limits
- `deposit_margin_percent` - Deposit margin percentage (10%)
- `global_withdrawal_limits` - Global withdrawal limits
- `require_kyc_on_first_withdrawal` - KYC requirement on first withdrawal
- `withdrawal_limit_validation_enabled` - Withdrawal limit validation
- `withdrawal_margin_percent` - Withdrawal margin percentage (10%)

### 7. Cashback Tiers (10 tiers)
1. **Bronze** (Level 1) - 0 GGR required, 10% cashback
2. **Iron** (Level 2) - 1,000 GGR required, 15% cashback
3. **Steel** (Level 3) - 5,000 GGR required, 20% cashback
4. **Gold** (Level 4) - 15,000 GGR required, 25% cashback
5. **Diamond** (Level 5) - 50,000 GGR required, 30% cashback
6. **Crystal** (Level 6) - 5,000,000 GGR required, 35% cashback
7. **Emerald** (Level 7) - 20,000,000 GGR required, 40% cashback
8. **Royal** (Level 8) - 30,000,000 GGR required, 43% cashback
9. **Crown** (Level 9) - 40,000,000 GGR required, 47% cashback
10. **Cosmic** (Level 10) - 50,000,000 GGR required, 50% cashback

### 8. Admin User: kirubel_gizaw32

**User Details:**
- **ID**: `1dba1be4-e7d6-4d99-88cd-604456da0b70`
- **Username**: `kirubel_gizaw32`
- **Email**: `kirubel.tech23@gmail.com`
- **Status**: ACTIVE
- **Is Admin**: true
- **KYC Status**: PENDING
- **Default Currency**: USD
- **Created**: 2025-10-26 21:47:49

**Password:**
- The password hash is stored: `$2a$12$OHjbqF3r4wXcXHfoXEaqMuYP0hlruk.RD.PxEHv7YvD.z14Tfiy06`
- **Note**: The original password is not stored for security reasons. You may need to reset the password or use the password reset functionality.

**Role Assignment:**
- Assigned to `admin role` (ID: `33dbb86c-e306-4d1d-b7df-cdf556e1ae32`)

**Initial Balance:**
- **Currency**: USD
- **Amount**: 5,496.50 USD (549,650 cents)
- **Brand**: TucanBIT Production
- **Reserved**: 0

## Usage

### Apply the seed migration:
```bash
./migrate.sh up
```

### Rollback the seed migration:
```bash
./migrate.sh down 1
```

## Important Notes

1. **Password**: The admin user password hash is included, but the actual password is not known. You may need to:
   - Use the password reset functionality
   - Or update the password hash if you know the original password

2. **Permissions**: The admin user has the `admin role` assigned, but you may need to ensure the role has the necessary permissions assigned via `role_permissions` table.

3. **Brand ID**: The user balance is associated with `TucanBIT Production` brand. Adjust if needed for your environment.

4. **Conflicts**: All INSERT statements use `ON CONFLICT DO NOTHING` to prevent errors if data already exists.

## Additional Data Included (Placeholders)

The migration file includes placeholders for the following data that needs to be generated from the database:

1. **Permissions** (204 permissions) - Section 11
2. **Role Permissions** (203 role-permission mappings for admin role) - Section 12
3. **Pages** (31 pages) - Section 13
4. **Admin Activity Actions** (24 actions) - Section 14

### Generating Additional Seed Data

To populate these sections:

1. **Run the generation script** (when database is available):
   ```bash
   python3 scripts/generate_seed_data.py > /tmp/additional_seed_data.sql
   ```

2. **Extract and add each section** to the migration file, replacing the placeholder comments.

3. **See detailed instructions** in `scripts/README_GENERATE_SEED.md`

The script will:
- Query the database for all required data
- Generate properly formatted INSERT statements
- Handle NULL values and data types correctly
- Use `ON CONFLICT DO NOTHING` to prevent errors

**Note**: The migration file currently has placeholder comments where this data should be inserted. Once the database is available, run the script and add the generated SQL to complete the seed migration.

### Other Data Not Included

- **Games** (3,227 games exist in production) - These can be added in a separate migration if needed, as they are large in volume.

