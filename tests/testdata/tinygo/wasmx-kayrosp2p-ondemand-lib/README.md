# Kayros Consensus

A new type of consensus based on our Kayros indexer. Each node will be initialized with a Kayros gateway url. The Kayros indexer is the one who orders transactions and gives them a timestamp. So, the node/validator sends the user transaction to Kayros and gets a Kayros record that is should also include in the block. And will use this Kayros order to order the transactions in a block.

Nodes therefore produce blocks continuously (they finalized blocks continuously). There is no Leader, because Kayros is the only source of truth. We will only check that nodes are in sync by propagating the block hashes and then comparing them per block. If a node sees > x% (e.g. 50%) of nodes having a different block hash, it will revert its blocks and take them from the validators with the majority. And then continue the protocol.

## Protocol

User sends the transaction to the node. The node adds the transaction to the mempool, registers it with Kayros if it was not registered. And forwards it to other nodes.

Block production:
- get max `x` records with data_type `wasmx_<chain_id>` since the previous record hash
- match records with mempool transactions
- if a record does not have a transaction, ask the other validators for the missing txs
- if tx does not appear in x time, produce block without it
- after block is produced, the block hash and txhash list is sent to the other nodes

Block check:
- block hashes are coming in our KAYROS_BLOCKHASH_CHATROOM and we keep them in a mapping block_number => []BlockHash , with hash and validator address; when we have > 50% we check if the hash matches our hash, if not, we rollback the blocks until that block and ask the validators with correct hash for the block.
- only after this check do we consider the block as stable (cannot be rolled back), even though the probability of producing a bad block is low (only when txs don't arrive in the mempool due to a bad actor)

## agents info

FinalizeBlock is part of the process to commit a block. we run this as part of block production. so we run this automatically, for every block, regardless of other nodes/validators. but in parallel, we take messages from the other nodes/validators who send their block hashes and check if those block hashes match ours; this is where our thresholds come in. if we see other validators sending a different blockhash (the message needs to be signed by the validator and we need to verify the signature), then if > CTX_THRESHOLD_COMMIT, we will rollback all the blocks until the offending block (including it) and ask validators with the correct hash for the blocks; these receiveCommit messages, with block hashes keep comming, so when we get a > CTX_THRESHOLD_FINALIZE it just means that the block for sure will not rollback, so we will no longer be concerned with receiveCommit messages about that block; but FinalizeBlock has been already called when the block was produced in the first place. it is different from RAFT, because RAFT and tendermint have an entire voting process and usually split block preparation, dissemination and finalization into multiple phases. but here, we fully produce and finalize the block locally, without voting. and we correct any issues post-block finalization, because they should be very rare. it is fast consensus, because ordering is done by kayros.
