# wasmx-erc20i

ERC20-style token that uses wasmx-identity user IDs instead of addresses.

- Balances and allowances are keyed by `user_id`.
- Caller identity is resolved via `ROLE_ACCOUNT_IDENTITY`.
- Mint/burn/transfer operate on user IDs, not bech32 addresses.
