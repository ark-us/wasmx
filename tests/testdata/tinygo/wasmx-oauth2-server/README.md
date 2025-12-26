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
