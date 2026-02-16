# Back Office Page Permissions

This document lists all permissions for every page in the back office, organized by feature area.

## Permissions with Value Requirements

Some permissions require a value/limit to be set when assigned to a role:
- **manual fund player** - Requires amount limit (e.g., max $5000 per day)
- **manual withdraw player** - Requires amount limit (e.g., max $3000 per day)
- **approve withdrawal** - Requires amount limit (e.g., max $10000 per approval)

## Dashboard

- **view dashboard** - Access to view the main dashboard

## Reports

- **view analytics dashboard** - Access to view analytics dashboard
- **view game report** - Access to view game reports
- **view transaction report** - Access to view transaction/wallet reports
- **view transaction details** - Access to view transaction details
- **view daily report** - Access to view daily reports
- **view big winners report** - Access to view big winners report
- **view player metrics report** - Access to view player metrics report
- **view country report** - Access to view country report
- **view game performance report** - Access to view game performance report
- **view provider performance report** - Access to view provider performance report
- **export reports** - Permission to export reports to CSV/Excel

## Player Management

- **view players** - Access to view player list
- **view player details** - Access to view individual player details
- **edit player** - Permission to edit player information
- **suspend player** - Permission to suspend player accounts
- **unsuspend player** - Permission to unsuspend player accounts
- **block player** - Permission to block player accounts
- **unblock player** - Permission to unblock player accounts
- **reset player password** - Permission to reset player passwords
- **manual fund player** ⚠️ - Permission to manually fund player accounts (requires value)
- **manual withdraw player** ⚠️ - Permission to manually withdraw from player accounts (requires value)

## Welcome Bonus

- **view welcome bonus** - Access to view welcome bonus settings
- **edit welcome bonus** - Permission to edit welcome bonus settings
- **view welcome bonus channels** - Access to view welcome bonus channel settings
- **create welcome bonus channel** - Permission to create welcome bonus channel rules
- **edit welcome bonus channel** - Permission to edit welcome bonus channel rules
- **delete welcome bonus channel** - Permission to delete welcome bonus channel rules

## Notifications

- **view notifications** - Access to view player notifications
- **create notification** - Permission to create notifications
- **send notification** - Permission to send notifications to players
- **view campaigns** - Access to view notification campaigns
- **create campaign** - Permission to create notification campaigns
- **edit campaign** - Permission to edit notification campaigns
- **delete campaign** - Permission to delete notification campaigns
- **send campaign** - Permission to send notification campaigns

## KYC Management

- **view kyc management** - Access to view KYC management page
- **approve kyc** - Permission to approve KYC requests
- **reject kyc** - Permission to reject KYC requests
- **view kyc risk** - Access to view KYC risk management
- **update kyc risk settings** - Permission to update KYC risk settings

## Rakeback/Cashback

- **view cashback** - Access to view cashback/VIP levels
- **edit cashback** - Permission to edit cashback/VIP level settings
- **view rakéback override** - Access to view rakéback override (Happy Hour)
- **create rakéback override** - Permission to create rakéback override
- **edit rakéback override** - Permission to edit rakéback override
- **delete rakéback override** - Permission to delete rakéback override
- **view rakéback schedules** - Access to view scheduled rakéback overrides
- **create rakéback schedule** - Permission to create scheduled rakéback overrides
- **edit rakéback schedule** - Permission to edit scheduled rakéback overrides
- **delete rakéback schedule** - Permission to delete scheduled rakéback overrides

## Transactions

- **view withdrawals** - Access to view withdrawal requests
- **approve withdrawal** ⚠️ - Permission to approve withdrawal requests (requires value)
- **reject withdrawal** - Permission to reject withdrawal requests
- **view withdrawal dashboard** - Access to view withdrawal dashboard
- **view withdrawal settings** - Access to view withdrawal settings
- **edit withdrawal settings** - Permission to edit withdrawal settings
- **view deposits** - Access to view deposit transactions
- **approve deposit** - Permission to approve deposit transactions
- **reject deposit** - Permission to reject deposit transactions
- **view manual funds** - Access to view manual fund management
- **create manual fund** ⚠️ - Permission to create manual fund transactions (requires value)
- **view fund logs** - Access to view fund logs

## Wallet Management

- **view wallet management** - Access to view wallet management
- **create wallet** - Permission to create wallets
- **edit wallet** - Permission to edit wallet settings
- **delete wallet** - Permission to delete wallets

## Game Management

- **view game management** - Access to view game management
- **add game** - Permission to add new games
- **edit game** - Permission to edit game settings
- **delete game** - Permission to delete games
- **enable game** - Permission to enable games
- **disable game** - Permission to disable games
- **update game status** - Permission to update game status

## Brand Management

- **view brand management** - Access to view brand management
- **create brand** - Permission to create brands
- **edit brand** - Permission to edit brand settings
- **delete brand** - Permission to delete brands

## Falcon Liquidity

- **view falcon liquidity** - Access to view Falcon Liquidity
- **edit falcon liquidity** - Permission to edit Falcon Liquidity settings

## Settings

- **view settings** - Access to view site settings
- **edit settings** - Permission to edit site settings
- **view welcome bonus settings** - Access to view welcome bonus settings in settings page
- **edit welcome bonus settings** - Permission to edit welcome bonus settings
- **view ip filters** - Access to view IP filters in settings
- **add ip filter** - Permission to add IP filters
- **remove ip filter** - Permission to remove IP filters

## Access Control

- **view access control** - Access to view access control page
- **view roles** - Access to view roles
- **create role** - Permission to create roles
- **edit role** - Permission to edit roles
- **delete role** - Permission to delete roles
- **view permissions** - Access to view permissions list
- **assign role** - Permission to assign roles to users
- **revoke role** - Permission to revoke roles from users
- **view admin users** - Access to view admin users
- **create admin user** - Permission to create admin users
- **edit admin user** - Permission to edit admin users
- **delete admin user** - Permission to delete admin users
- **view kyc settings** - Access to view KYC settings in access control
- **edit kyc settings** - Permission to edit KYC settings

## Admin Activity Logs

- **view activity logs** - Access to view admin activity logs
- **export activity logs** - Permission to export activity logs

## Alert Management

- **view alerts** - Access to view alert management
- **create alert** - Permission to create alerts
- **edit alert** - Permission to edit alerts
- **delete alert** - Permission to delete alerts
- **view alert configurations** - Access to view alert configurations
- **create alert configuration** - Permission to create alert configurations
- **edit alert configuration** - Permission to edit alert configurations
- **delete alert configuration** - Permission to delete alert configurations
- **view email groups** - Access to view email groups
- **create email group** - Permission to create email groups
- **edit email group** - Permission to edit email groups
- **delete email group** - Permission to delete email groups

## Total Permissions

**121 permissions** covering all pages and features in the back office.

## Usage

When creating or editing roles, you can assign these permissions. For permissions marked with ⚠️, you must also specify:
- **Value**: The limit amount (e.g., 5000 for $5000)
- **Limit Type**: Daily, Weekly, or Monthly
- **Limit Period**: Number of periods (e.g., 1 for "1 daily", 2 for "2 weekly")

Example: "manual fund player" with value 5000, limit_type "daily", limit_period 1 = "Can fund up to $5000 per 1 day"

