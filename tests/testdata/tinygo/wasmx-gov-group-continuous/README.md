# wasmx-gov-group-continuous

Group-aware continuous governance contract. It keeps the continuous voting
algorithm, but gates proposals and votes by group membership and power.

Key points:
- Proposals include `group_id` and `group_contract`.
- Voter eligibility and power come from `query_get_voter_power` on the groups
  contract.
- Deposits are real bank deposits; group token denom is enforced when present.

Designed to work with wasmx-groups (group_vote and group_erc20).
