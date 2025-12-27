# Kayros Consensus

A new type of consensus based on our Kayros indexer. Each node will be initialized with a Kayros gateway url. The Kayros indexer is the one who orders transactions and gives them a timestamp. So, the node/validator sends the user transaction to Kayros and gets a Kayros record that is should also include in the block. And will use this Kayros order to order the transactions in a block.

Nodes therefore can produce blocks continuously. There is no Leader, because Kayros is the only source of truth. We will only check that nodes are in sync by propagating the block hashes and then comparing them per block. If a node sees > x% (e.g. 50%) of nodes having a different block hash, it will revert its blocks and take them from the validators with the majority. And then continue the protocol.

## Protocol

User sends the transaction to the node. The node adds the transaction to the mempool, registers it with Kayros if it was not registered. And forwards it to other nodes.

Block production:
- get max `x` records with data_type `wasmx_<chain_id>` since the previous record hash
- match records with mempool transactions
- if a record does not have a transaction, ask the other validators for the missing txs
- if tx does not appear in x time, produce block without it
- after block is produced, the block hash and txhash list is sent to the other nodes

Block check:
- block hashes are coming in our KAYROS_BLOCKHASH_CHATROOM and we keep them in a mapping block_number => []BlockHash , with hash and validator address; when we have > 2/3 we check if the hash matches our hash, if not, we rollback the blocks until that block and ask the validators with correct hash for the block.
- only after this check do we consider the block truly finalized, even though the probability of producing a bad block is low (only when txs don't arrive in the mempool due to a bad actor)

