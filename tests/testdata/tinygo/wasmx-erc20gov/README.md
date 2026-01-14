# wasmx-erc20gov

Governance-controlled ERC20-style token that uses wasmx-identity user IDs.

- `mint`, `burn`, and `transfer` are restricted to the configured governance contract.
- `transferFrom` is restricted to admin user IDs (no allowance flow).
- Governance address can be updated by the current governance contract.
