# wasmx-groups

Groups contract for wasmX that owns group membership and coordinates
group-based governance. Membership is based on either:
- group_vote: explicit membership stored in this contract (user_id from identity)
- group_erc20: implicit membership from ERC20 balance (balance > 0)

Core flow:
- Initialize with `identity_contract` via `init_genesis`.
- Create groups with `protocol.governance_contract` and optional token settings
  (`token` or `token_denom`) for ERC20-based membership and voting power.
- Add/remove members by `user_id` (group_vote), with checks against identity.

Governance integration (two supported paths):
1) Gov queries groups: governance contracts call `query_get_voter_power` to
   validate membership and fetch voting power in one call.
2) Group-mediated forwarding: users call the groups contract, which validates
   membership and forwards `submit_group_proposal` / `vote_group_proposal` to
   the configured governance contract.

Membership and power queries:
- `query_is_member` is token-aware (ERC20 balance > 0) when token settings exist.
- `query_get_voter_power` returns `is_member`, `power`, and token info in one call.

Queries:
- `query_is_member`, `query_get_voter_power`, `query_get_all_members`,
  `query_get_group`, `query_get_groups`, `query_get_user_groups`,
  `query_get_voting_protocol`, `query_get_config`.
