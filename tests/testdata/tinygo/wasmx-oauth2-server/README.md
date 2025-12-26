# OAuth2 Server

This is an OAuth2 server for blockchain. The user registers an account on /register and a secp256k1 key pair is generated and stored in the browser.
The address is then sent with /register_init and the wasmx-oauth2-keys contract initiates a transaction to fund the account.
Then, a transaction is signed with the user's private key, to register the account with wasmx-identity.
At this point, the user has an account and any time the user logs with this OAuth2 server, a new ephemeral key is generated and kept by wasmx-oauth2-keys for creating transactions on behalf of the user (for POST requests), with the address added to his identity.

The OAuth2 Server registers its http apis with wasmx-httpserver-registry
