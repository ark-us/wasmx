# wasmx-gov-group

Group-aware governance contract (stake-style). It enforces voting eligibility
and power by querying a groups contract once per vote.

Key points:
- Proposals include `group_id` and `group_contract`.
- Voter power is fetched from `query_get_voter_power` on the groups contract.
- Deposits are real bank deposits; denom is enforced by the group token settings
  (or the default gov denom when no token is set).

Designed to work with wasmx-groups (group_vote and group_erc20).
