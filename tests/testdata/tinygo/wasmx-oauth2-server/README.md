# wasmx-oauth2-server

OAuth2 server for WasmX blockchain. Provides standard OAuth2 authentication with blockchain identity integration.

## Overview

WasmX OAuth2 server acts as a standard OAuth2 provider that any application can use. All authentication is delegated to Nomen (identity aggregation app), ensuring users have a unified identity across services.

## Architecture

```
External App → WasmX /oauth/authorize → WasmX /login
                                              ↓
                                    "Continue with Nomen"
                                              ↓
                              Nomen Auth (Google, GitHub, etc.)
                                              ↓
                              Nomen JWT returned to WasmX
                                              ↓
                              WasmX validates JWT, creates user
                                              ↓
                              Browser generates secp256k1 key
                                              ↓
                              Key encrypted with PIN → Supabase Vault
                                              ↓
                              Public key registered on blockchain
                                              ↓
                              OAuth2 tokens → External App
```

## Registration Flow

### Step 1: Nomen Authentication
User clicks "Continue with Nomen" on the WasmX login page. This redirects to Nomen where they authenticate with their preferred provider (Google, GitHub, etc.).

### Step 2: Key Generation (Browser-side)
After Nomen authentication returns a JWT:
1. Browser generates a secp256k1 key pair locally
2. User enters a PIN to encrypt the private key
3. Encrypted key stored in Supabase Vault (via nomen-oauth2 SDK)
4. Only the public key is sent to WasmX

### Step 3: Account Initialization
```
POST /auth/register_init
Body: { "address": "mythos1..." }
```
- WasmX calls `wasmx-oauth2-keys.InitAccount`
- Funder account sends native coins to the new address
- Transaction broadcasted via `BroadcastTxAsync`

### Step 4: Identity Registration (Self-signed)
```
POST /auth/register_tx
Body: { "address": "mythos1...", "signed_tx": "<base64>" }
```
- User signs a `register_user` transaction with their private key
- Transaction sent to `wasmx-identity` contract
- **Self-registration enforced**: sender must equal address being registered
- Creates user identity on-chain

### Step 5: OAuth2 User Creation
```
POST /auth/register
Headers: Authorization: Bearer <nomen_jwt>
Body: { "address": "mythos1...", "public_key": "...", "username": "..." }
```
- WasmX validates Nomen JWT
- Creates OAuth2 user linked to Nomen profile
- Associates blockchain address with user

## Login Flow (Returning Users)

### Step 1: Nomen Authentication
User authenticates via Nomen, receives JWT.

### Step 2: Session Creation
```
POST /auth/login
Headers: Authorization: Bearer <nomen_jwt>
```
- WasmX validates JWT
- Looks up existing user by Nomen profile
- Creates session with blockchain address info

### Step 3: OAuth2 Authorization
Standard OAuth2 flow with PKCE:
1. App redirects to `/oauth/authorize` with client_id, redirect_uri, code_challenge
2. User approves authorization
3. App receives authorization code
4. App exchanges code for tokens at `/oauth/token`

### Step 4: Ephemeral Key Generation
During token exchange:
- `wasmx-oauth2-keys.GenerateEphemeralKey` creates a new key pair
- Key linked to OAuth token for transaction signing
- Ephemeral address funded automatically

## OAuth2 Token Exchange Flow

```
POST /oauth/token
Body: grant_type=authorization_code&code=...&code_verifier=...

Response:
{
  "access_token": "token_...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_..."
}
```

During token exchange:
1. Validates authorization code and PKCE verifier
2. Calls `wasmx-oauth2-keys` to generate ephemeral key
3. Ephemeral key stored encrypted in oauth2-keys contract
4. Returns OAuth2 tokens to application

## POST Request Flow (Transaction Signing)

```
Browser: POST /tools/userdata/set_favorite_color
Headers: Authorization: Bearer <oauth_token>
Body: {"color": "blue"}
         ↓
wasmx-httpserver-registry.handleMutatingRequest()
         ↓
Calls wasmx-oauth2-keys.SignAndBroadcastTx(oauth_token, calldata)
    ├── Validates OAuth token
    ├── Retrieves encrypted ephemeral key
    ├── Decrypts with server_secret + oauth_token
    ├── Signs transaction with ephemeral private key
    └── Broadcasts via BroadcastTxAsync
         ↓
Return {"tx_hash": "...", "address": "..."} to browser
```

The browser never needs to sign transactions directly - it just sends the OAuth token. The `wasmx-oauth2-keys` contract handles all signing with ephemeral keys.

## GET Request Flow (Read-only)

```
Browser: POST /tools/userdata/get_favorite_color  (use_transaction=false)
Headers: Authorization: Bearer <oauth_token>
         ↓
wasmx-httpserver-registry validates token
         ↓
Forwards to target contract with user context
         ↓
Return data directly (no transaction needed)
```

## API Endpoints

### Authentication
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/login` | GET/POST | Login page / process login |
| `/auth/register` | GET/POST | Registration page / create user |
| `/auth/register_init` | POST | Initialize new account (fund it) |
| `/auth/register_tx` | POST | Submit signed registration transaction |
| `/auth/logout` | POST | End session |

### OAuth2
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/oauth/authorize` | GET/POST | Authorization endpoint |
| `/oauth/token` | POST | Token exchange |
| `/oauth/clients/register` | POST | Register OAuth2 client |

### Nomen Integration
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/nomen/validate` | POST | Validate Nomen JWT, get/create WasmX user |
| `/auth/nomen/link` | POST | Link existing WasmX user to Nomen profile |

## Security Model

1. **No password storage** - All authentication through Nomen OAuth2
2. **Browser-generated keys** - Private keys never touch the server
3. **PIN encryption** - Keys encrypted client-side before vault storage
4. **Self-registration** - Only users can register their own addresses
5. **Ephemeral keys** - Short-lived keys for transaction signing
6. **PKCE required** - All OAuth2 flows use S256 code challenge

## Dependencies

- `wasmx-oauth2-keys` - Ephemeral key management and transaction signing
- `wasmx-identity` - On-chain identity registration
- `wasmx-httpserver-registry` - HTTP route registration
- Nomen - External identity provider

## Configuration

Initialize via `init_genesis`:
```json
{
  "init_genesis": {
    "route_prefix": "/auth",
    "jwt_secret": "...",
    "nomen_public_key": "..."
  }
}
```

## Storage

- `user:<user_id>` → User profile
- `session:<session_id>` → Session data
- `client:<client_id>` → OAuth2 client
- `auth_code:<code>` → Authorization code (temporary)
