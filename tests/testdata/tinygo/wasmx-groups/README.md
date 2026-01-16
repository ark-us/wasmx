# wasmx-groups

Groups contract for wasmX that coordinates group-based governance. Membership
is derived from ERC20-style token balances (balance >= min_balance).

Core flow:
- Create groups with `protocol.governance_contract` and token settings
  (`token` plus `min_balance`).

Governance integration (two supported paths):
1) Gov queries groups: governance contracts call `query_get_voter_power` to
   validate membership and fetch voting power in one call.
2) Group-mediated forwarding: users call the groups contract, which validates
   membership and forwards `submit_group_proposal` / `vote_group_proposal` to
   the configured governance contract.

Membership and power queries:
- `query_is_member` checks `balanceOf` against `min_balance`.
- `query_get_voter_power` returns `is_member`, `power`, and token info in one call.

Queries:
- `query_is_member`, `query_get_voter_power`, `query_get_group`,
  `query_get_groups`, `query_get_user_groups`, `query_get_voting_protocol`,
  `query_get_config`.
