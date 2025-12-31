# OAuth2 Server

This is an OAuth2 server for blockchain. The user registers an account on /register and a secp256k1 key pair is generated and stored in the browser.
The address is then sent with /register_init and the wasmx-oauth2-keys contract initiates a transaction to fund the account.
Then, a transaction is signed with the user's private key, to register the account with wasmx-identity.
At this point, the user has an account and any time the user logs with this OAuth2 server, a new ephemeral key is generated and kept by wasmx-oauth2-keys for creating transactions on behalf of the user (for POST requests), with the address added to his identity.

The OAuth2 Server registers its http apis with wasmx-httpserver-registry


Current Login/Token Flow

Login → Authorize → Token Exchange:

1. User logs in (/login) with username/password → session created
2. App requests authorization (/oauth/authorize) with client_id and redirect_uri
3. User approves → authorization code created
4. App exchanges code for token (/oauth/token) with client_id and client_secret
5. During token exchange (ExchangeCodeForToken in oauth_flow.go:206-214):
- Calls generateEphemeralKey() in users.go:485
- oauth2-server calls wasmx-oauth2-keys contract with generate_ephemeral_key message
- wasmx-oauth2-keys generates key pair (actions.go:64-129)
- Stores ephemeral key in oauth2-keys contract storage


POST Request Flow with OAuth2:

Browser sends POST with "Authorization: Bearer <oauth_token>"
                    ↓
wasmx-httpserver-registry.handleMutatingRequest()
                    ↓
Calls wasmx-oauth2-keys.QueryValidateAndGetKey(oauth_token)
    ├── Lookup: token.<token> → public_key
    ├── Lookup: key.<pubkey_hex> → encrypted key pair
    ├── Derive encryption key: HKDF(server_secret, oauth_token)
    ├── Decrypt private key with AES-256-GCM
    └── Return plaintext private key + address
                    ↓
httpserver-registry builds transaction:
    ├── Target contract + HttpRequestHandler calldata
    ├── Signs with ephemeral private key
    └── wasmxcore.PrepareTx(..., keyResponse.PrivateKey)
                    ↓
Broadcast transaction: wasmxcore.BroadcastTxAsync(txBytes)
                    ↓
Return {"tx_hash": "..."} to browser

The wasmx-oauth2-keys contract stores encrypted ephemeral private keys and decrypts them on demand. The wasmx-httpserver-registry contract signs transactions with these ephemeral keys for POST/PUT/DELETE requests when UseTransaction=true on the route. The browser never needs to sign - it just sends the OAuth token.
