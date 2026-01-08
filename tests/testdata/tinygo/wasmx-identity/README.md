# wasmx-identity

Identity aggregation contract for WasmX. Maps multiple blockchain addresses to a single user identity.

## Purpose

- Allow users to have multiple addresses (keys) associated with one identity
- Support ephemeral keys with expiration and limited permissions
- Enable features like account recovery, delegation, and multi-device access

## Security Model

**Self-registration only**: The transaction sender must be the address being registered or modified.

- `register_user`: Sender must equal the address being registered
- `add_address`: Sender must be an existing address of the user
- `remove_address`: Sender must be an existing address of the user

This ensures no contract or third party can register addresses on behalf of users.

## Data Structures

### UserIdentity
```go
type UserIdentity struct {
    UserID         string   // Unique user ID (e.g., "user_1")
    PrimaryAddress string   // Main address for the user
    Addresses      []string // All associated addresses
    CreatedAt      int64
    UpdatedAt      int64
}
```

### AddressInfo
```go
type AddressInfo struct {
    Address           string
    PublicKey         []byte
    ServiceDomain     string       // Empty for regular, domain for ephemeral
    Permissions       []Permission // What this address can do
    ExpiresAt         int64        // 0 = never expires
    DefaultAllowances []wasmx.Coin // ERC20 allowances from primary
}
```

## Entry Points

### register_user
Register a new user with their first address. **Self-registration only**.

```json
{
  "register_user": {
    "address": "mythos1...",
    "public_key": "<base64>",
    "service_domain": "",
    "permissions": [],
    "expires_at": 0
  }
}
```

### add_address
Add a new address to an existing user. Caller must own an existing address of this user.

```json
{
  "add_address": {
    "user_id": "user_1",
    "address": "mythos1...",
    "public_key": "<base64>",
    "service_domain": "app.example.com",
    "expires_at": 1735689600,
    "default_allowances": [{"denom": "amyt", "amount": "1000000"}]
  }
}
```

### remove_address
Remove an address from a user. Cannot remove the primary address.

### Query Operations
- `query_user_by_id`: Get user by user ID
- `query_user_by_address`: Get user by any of their addresses
- `query_address_info`: Get detailed info for a specific address

## Storage

- `user:<user_id>` -> UserIdentity
- `addr:<address>` -> user_id (reverse lookup)
- `addrinfo:<user_id>:<address>` -> AddressInfo

## Dependencies

- `wasmx-env`: Core WASM environment
- Deterministic contract (no async operations)

## Architecture Notes

This is a **deterministic** contract. It cannot be called directly by non-deterministic contracts like `wasmx-oauth2-keys`. Instead:

1. User generates key pair
2. `wasmx-oauth2-keys` funds the new address (async broadcast)
3. User signs a `register_user` transaction with their own key
4. User broadcasts the registration transaction

This maintains the security invariant that only users can register their own addresses.
