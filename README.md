# TucanBIT Online Casino

A dynamic platform for TucanBIT online casino, offering robust user registration, authentication, and operational group management.

**This is API documentation for TucanBIT online casino**

---

## Table of Contents
- [Endpoints](#endpoints)
  - [User Management](#user-management)
  - [Operational Group](#operational-group)
  - [Operational Group Type](#operational-group-type)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

---

## Endpoints

### User Management

#### **User Registration**
- **Endpoint:** `/register`
- **Method:** `POST`
- **Description:** Register a new user on the platform.

**Validation Rules:**
- **Username:**
  - Required.
  - 3–20 characters.
  - Alphanumeric (optional: allow underscores or hyphens).
  - Must be unique.
- **Phone Number:**
  - Required.
  - Must follow valid E.164 phone number format.
  - Must be unique.
- **Password:**
  - Required.
  - Minimum 8 characters.
  - Must include:
    - At least one uppercase letter.
    - At least one lowercase letter.
    - At least one digit.
    - At least one special character.

#### **User Login**
- **Endpoint:** `/login`
- **Method:** `POST`
- **Description:** Authenticate and log in users to access the platform.

---

### Operational Group

#### **Create Operational Group**
- **Endpoint:** `/operationalgroup`
- **Method:** `POST`
- **Description:** Allow admins to create a new operational group.

#### **Get Operational Groups**
- **Endpoint:** `/operationalgroup`
- **Method:** `GET`
- **Description:** Retrieve a list of all operational groups.

---

### Operational Group Type

#### **Create Operational Group Type**
- **Endpoint:** `/operationalgrouptype`
- **Method:** `POST`
- **Description:** Allow admins to define new operational group types.

#### **Get Operational Group Type**
- **Endpoint:** `/operationalgrouptype/:groupID`
- **Method:** `GET`
- **Description:** Retrieve operational group types associated with a specific `groupID`.

---

### balance 

#### **update balance**
- **Endpoint:** `/api/balance/update`
- **Method:** `POST`
- **Description:** Allow users and admins to update balance. it should be protected by role.


#### **Get balance**
- **Endpoint:** `/api/balance`
- **Method:** `GET`
- **Description:** Retrieve user balance. user must also provide authorization header, which is Bearer token

---

### balancelogs 

#### **Get balance logs**
- **Endpoint:** `/api/balance/logs`
- **Method:** `GET`
- **Description:** Retrieve balance logs

**Query Parameters:**
- **user-id:**
  - Optional.
  - uuid.
- **operation_group_id**
  - Optional.
  - uuid.- 
 **operation_type_id**
  - Optional.
  - uuid.
 **start-date**
  - Optional.
  - format: YYYY-MM-DD.
 **end-date**
  - Optional.
  - format: YYYY-MM-DD.
 **page**
  - Required.
  - number.
 **per-page**
  - Required.
  - number.
---

### exchange 

#### **Get Conversion rate**
- **Endpoint:** `api/conversion/rate`
- **Method:** `POST`
- **Description:** Retrieve exchange rate

**Parameters:**
- **currency_from:**
  - Required.
  - ISO 4217 currency code
- **currency_to**
  - Required.
  - ISO 4217 currency code


#### **Exchange**
- **Endpoint:** `api/conversion/rate`
- **Method:** `POST`
- **Description:** exchange currency from one currency to another

**Parameters:**
- **currency_from:**
  - Required.
  - ISO 4217 currency code
- **currency_to**
  - Required.
  - ISO 4217 currency code
- **amount**
  - Required
  - Decimal  
---

#### **HandleWS**
- **Endpoint:** `/ws`
- **Method:** `GET`
- **Description:** HandleWS sets up a WebSocket connection for a user to interact with the game in real time

**body:**
after getting initial connection the user has to send auth access token 
- **access_token:**
  - Required.
  - string  
---

#### **GetOpenRound**
- **Endpoint:** `/api/game/round`
- **Method:** `GET`
- **Description:** Get allow users to get open round

---

#### **PlaceBet**
- **Endpoint:** `/api/game/place-bet`
- **Method:** `POST`
- **Description:** PlaceBet allow user to bet for open round
**body:**
- **round_id:**
  - Required.
  - uuid  
- **currency:**
  - Required.
  - string  
- **amount:**
  - Required.
  - decimal (min 1 max 1000)  
  
---

#### **CashOut**
- **Endpoint:** `/api/game/cash-out`
- **Method:** `POST`
- **Description:** CashOut allow user to cashout in progress bets
**body:**
- **round_id:**
  - Required.
  - uuid  

---

#### **GetBetHistory**
- **Endpoint:** `/api/admin/game/history`
- **Method:** `GET`
- **Description:** Retrieve user bets based on user_id (opetional)
**parameter:**
- **user-id:**
  - Optional.
  - uuid  
  
---

#### **GeMytBetHistory**
- **Endpoint:** `/api/user/game/history`
- **Method:** `GET`
- **Description:** Retrieve user to retrieve users bet
**parameter:**
- **user-id:**
  - Optional.
  - uuid  
  
---

#### **Cancel Bet**
- **Endpoint:** `/api/game/cancel`
- **Method:** `POST`
- **Description:** Cancel User Bet
**parameter:**
- **round_id:**
  - required.
  - uuid  
  
---


#### **Change Password**
- **Endpoint:** `/api/user/password`
- **Method:** `PATCH`
- **Description:** change user password
**parameter:**
- **old_password:**
  - required.
  - string   
 
- **old_password:**
  - required.
  - string   

- **confirm_password:**
  - required.
  - string   

  **Validation Rules:**
- **new_password:**
  - Required.
  - 3–20 characters.
  - Alphanumeric (optional: allow underscores or hyphens).
  - Must be unique.
---

#### **Get Leaders**
- **Endpoint:** `/api/game/leaders`
- **Method:** `Get`
- **Description:** Get bet leaders
  
---

#### **Forget Password**
- **Endpoint:** `/api/user/password/forget`
- **Method:** `POST`
- **Description:** Forget password Request
**parameter:**
- **username:**
  - required.
  - email,phone, or username (string)
  
---

#### **Verify OTP For Forget Password**
- **Endpoint:** `/api/user/password/forget/verify`
- **Method:** `POST`
- **Description:** Verify OTP For Forget password Request
**parameter:**
- **username:**
  - required.
  - email,phone, or username  
 - **otp:**
  - required.
  - string
  
---

#### **Verify OTP For Forget Password**
- **Endpoint:** `/api/user/password/reset`
- **Method:** `POST`
- **Description:** Reset Password
**parameter:**
- **token:**
  - required.
  - string 
 - **new_password:**
  - required.
  - string
 - **confirm_password:**
  - required.
  - string  
---


#### **Verify OTP For Forget Password**
- **Endpoint:** `/api/user/password/reset`
- **Method:** `POST`
- **Description:** Reset Password
**parameter:**
- **token:**
  - required.
  - string 
 - **new_password:**
  - required.
  - string
 - **confirm_password:**
  - required.
  - string  
---

#### **update profile**
- **Endpoint:** `/api/user/profile`
- **Method:** `POST`
- **Description:** update profile
**parameter:**
- **token:**
  - required.
  - string 
 - **first_name:**
  - optional.
  - string
 - **last_name:**
  - optional.
  - string  
- **email:**
  - optional.
  - string  
- **date_of_birth**
   - optional 
   - string 
---


#### **update profile**
- **Endpoint:** `/user/oauth/google`
- **Method:** `GET`
- **Description:** sigin with google

---

#### **update profile**
- **Endpoint:** `/user/oauth/facebook`
- **Method:** `GET`
- **Description:** sigin with facebook

---
#### **update profile**
- **Endpoint:** `/api/admin/user/block`
- **Method:** `POST`
- **Description:** block account
**parameter:**
- **token:**
  - required.
  - string 
 - **duration:**
  - required.
  - string (temporary  or permanent)
 - **type:**
  - required. (financial, gaming, login, complete)
  - string  
- **blocked_from:**
  - required for duration temporary. (example 2025-01-15T04:03:00Z)
  - string 
- **blocked_to:**
  - required for duration temporary. (example 2025-01-17T05:03:00Z)
  - string  
- **note**
   - optional 
   - string 

- **reason**
   - optional 
   - string 


---


#### **update profile**
- **Endpoint:** `/api/admin/user/block/accounts`
- **Method:** `POST`
- **Description:** get blocked accounts
**parameter:**
- **token:**
  - required.
  - string 
 - **per_page:**
  - required.
  - int
 - **page:**
  - required
  - int
- **user_id:**
  - optional 
  - uuid 
- **duration:**
  - optional 
  - string  
- **type**
   - optional 
   - string 

---


#### **update profile**
- **Endpoint:** `/api/admin/departments`
- **Method:** `POST`
- **Description:** create department
**parameter:**
- **name:**
  - required.
  - string 
 - **notifications:**
  - optional.
  - []string (which are the list of notification department gets by default)
 

---

#### **update profile**
- **Endpoint:** `/api/admin/departments`
- **Method:** `GET`
- **Description:** create department
**parameter:**
 - **per_page:**
  - required.
  - int
 - **page:**
  - required
  - int

---

#### **update profile**
- **Endpoint:** `/api/admin/departments`
- **Method:** `PATCH`
- **Description:** update department
**parameter:**
- **name:**
  - required.
  - string 
 - **notifications:**
  - optional.
  - []string (which are the list of notification department gets by default)
 
---

#### **update profile**
- **Endpoint:** `/api/admin/departments/assign`
- **Method:** `POST`
- **Description:** assign department
**parameter:**
- **user_id:**
  - required.
  - uuid 
 - **department_id:**
  - required.
  - uuid
 
---

#### **add ip filter**
- **Endpoint:** `/api/admin/ipfilters`
- **Method:** `POST`
- **Description:** assign department
**parameter:**
- **start_ip:**
  - required.
  - string
- **end_ip:**
  - optional.
  - string 
 - **type:**
  - required.
  - string (allow or deny)
 
#### **get financial metrics**
- **Endpoint:** `/api/admin/performance/financial`
- **Method:** `GET`
- **Description:** get finanicial metrics

---
#### **get game metrics**
- **Endpoint:** `/api/admin/performance/game`
- **Method:** `GET`
- **Description:** get game metrics

---

#### **add fund manually**
- **Endpoint:** `/api/admin/balance/add/fund`
- **Method:** `POST`
- **Description:** Add Fund Manually 
**parameter:**
- **amount:**
  - required.
  - decimal
- **reason:**
  - required.
  - decimal 
 - **Currency:**
  - optional.
  - string 
 - **Note:**
  - optional.
  - string 

---

#### **remove fund manually**
- **Endpoint:** `/api/admin/balance/remove/fund`
- **Method:** `POST`
- **Description:** Remove Fund Manually 
**parameter:**
- **amount:**
  - required.
  - decimal
- **reason:**
  - required.
  - decimal 
 - **Currency:**
  - optional.
  - string 
 - **Note:**
  - optional.
  - string 

---


#### **remove fund manually**
- **Endpoint:** `/api/admin/balance/remove/fund`
- **Method:** `POST`
- **Description:** Remove Fund Manually 
**parameter:**
- **amount:**
  - required.
  - decimal
- **reason:**
  - required.
  - decimal 
 - **Currency:**
  - optional.
  - string 
 - **Note:**
  - optional.
  - string 

---

## API Documentation

### Swagger UI
- **URL:** `/swagger/index.html`
- **Description:** View and test API endpoints using Swagger's interactive interface.


---

## License

This project is licensed under the [MIT License](LICENSE).
