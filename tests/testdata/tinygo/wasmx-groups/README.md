# wasmx-groups

Groups contract for wasmX that manages membership by `user_id` from the identity
contract. It supports creating groups, adding/removing members, and querying
membership and group metadata.

Core flow:
- Initialize with `identity_contract` (and optional genesis groups) via
  `init_genesis`.
- Create groups with a `protocol.governance_contract` that is authorized to
  manage membership changes.
- Add/remove members by `user_id`, with checks against the identity contract.

Queries:
- `query_is_member`, `query_get_all_members`, `query_get_group`,
  `query_get_groups`, `query_get_user_groups`, `query_get_voting_protocol`,
  `query_get_config`.
