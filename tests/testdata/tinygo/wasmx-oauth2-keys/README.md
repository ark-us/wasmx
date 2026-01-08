# wasmx-oauth2-keys

Ephemeral key management contract for WasmX OAuth2 authentication. Generates, stores, and uses secp256k1 key pairs linked to OAuth tokens.

## Purpose

- Generate ephemeral secp256k1 key pairs for OAuth2-authenticated users
- Store encrypted private keys in non-deterministic storage
- Fund new ephemeral accounts with native coins
- Sign and broadcast transactions on behalf of users

## Architecture

This is a **non-deterministic** contract. It:
- Uses cryptographic operations (key generation, encryption)
- Stores encrypted private keys
- Broadcasts transactions asynchronously

**Important**: Non-deterministic contracts cannot call deterministic contracts directly. State changes to deterministic contracts must be done via signed transactions (`PrepareTx` + `BroadcastTxAsync`).

## Flow

```
1. User authenticates via OAuth2 → gets OAuth token
2. wasmx-oauth2-keys generates ephemeral key pair
3. Private key encrypted with server_secret + oauth_token
4. InitAccount funds the new address (async broadcast)
5. User receives public key + private key + address
6. User signs their own registration with wasmx-identity
```

## Data Structures

### EphemeralKeyPair
```go
type EphemeralKeyPair struct {
    PublicKey     []byte // secp256k1 compressed public key
    PrivateKey    []byte // Encrypted with derived key
    UserID        string
    ServiceDomain string
    CreatedAt     int64
    ExpiresAt     int64
    OAuthToken    string
}
```

## Entry Points

### init_genesis
Initialize the contract with configuration.

```json
{
  "init_genesis": {
    "server_secret": "optional-secret",
    "funder_priv_key": "hex-encoded-private-key",
    "init_account_amt": {"denom": "amyt", "amount": "1000000"},
    "gas_price": {"denom": "amyt", "amount": "10"},
    "route_prefix": "/auth"
  }
}
```

### generate_ephemeral_key
Generate a new ephemeral key pair and fund the account.

```json
{
  "generate_ephemeral_key": {
    "oauth_token": "token-from-oauth2-server",
    "user_id": "user_123",
    "service_domain": "app.example.com",
    "expires_at": 1735689600
  }
}
```

Response:
```json
{
  "public_key": "<base64>",
  "private_key": "<base64>",
  "address": "mythos1...",
  "success": true
}
```

### register_external_key
Register a key pair generated externally (e.g., in browser).

```json
{
  "register_external_key": {
    "oauth_token": "...",
    "user_id": "user_123",
    "public_key": "<base64>",
    "private_key": "<base64>",
    "service_domain": "app.example.com",
    "expires_at": 1735689600
  }
}
```

### sign_and_broadcast_tx
Sign and broadcast a transaction using the ephemeral key.

```json
{
  "sign_and_broadcast_tx": {
    "oauth_token": "...",
    "target_contract": "mythos1...",
    "calldata": "<base64>",
    "gas_limit": 500000,
    "gas_price": {"denom": "amyt", "amount": "10"}
  }
}
```

### init_account
Fund a new account with native coins. Internal use only.

```json
{
  "init_account": {
    "address": "mythos1..."
  }
}
```

### revoke_key
Revoke an ephemeral key.

```json
{
  "revoke_key": {
    "oauth_token": "..."
  }
}
```

## Query Operations

### query_get_public_key
Get public key for an OAuth token.

### query_get_key_info
Get key metadata (without private key).

### query_validate_and_get_key
Validate OAuth token and return full key info including decrypted private key.

## Storage

Non-deterministic storage (not part of blockchain state):
- `key.<public_key_hex>` -> EphemeralKeyPair
- `token.<oauth_token>` -> public_key (reverse lookup)
- `server_secret` -> Master secret for encryption
- `funder_privkey` -> Private key for funding accounts

## Security Model

1. **Private keys encrypted at rest** using AES-256-GCM
2. **Encryption key derived** from server_secret + oauth_token via HKDF
3. **OAuth token required** to decrypt private keys
4. **OnlyInternal** check on sensitive operations
5. **Funder key isolated** - only used for InitAccount funding

## Integration with wasmx-identity

This contract does NOT register users with wasmx-identity directly. The correct flow:

1. `generate_ephemeral_key` creates key pair and funds account
2. User receives private key
3. User signs a `register_user` transaction to wasmx-identity
4. User broadcasts the registration transaction

This maintains the security invariant that only users can register their own addresses.

## Dependencies

- `wasmx-env`: Core WASM environment
- `wasmx-env-core`: Transaction preparation and broadcast (`PrepareTx`, `BroadcastTxAsync`)
- secp256k1 cryptography
