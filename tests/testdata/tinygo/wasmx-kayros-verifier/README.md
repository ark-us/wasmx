# wasmx-kayros-verifier

TinyGo library that verifies Kayros records for wasmx consensus.

What it does
- Validates record hash: `hash(prev_hash || data_type || data_item) == hash_item`.
- Validates chain linkage: `prev_hash` equals previous record `hash_item`.
- Validates timestamp and UUID consistency.
- Validates level rollups using Kayros level hashes.

Kayros levels
- Level 1 hashes roll up base records in groups of 256.
- Level 2 hashes roll up Level 1 hashes, and so on.
- To verify inclusion, supply the hashes used at each level (per position) and compare
  the computed rollup to the Kayros level hash.
