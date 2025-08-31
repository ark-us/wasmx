# Architecture

## Block End

```c++
tendermintp2p -> host.FinalizeBlock
tendermintp2p -> host.EndBlock -> app.EndBlocker
    -> call hooks.RunHook("EndBlock", data)
        -> hooks.RunHook -> call all EndBlock hooks
        -> gov.EndBlock hook -> executeProposal -> role change event
    -> cosmos modules EndBlock hook
    -> app.EndBlocker sets contract hooks events
    -> host.EndBlock returns
tendermintp2p -> host.Commit
```

## Optimistic Execution

- `metainfo` is for storing cross-chain tx results
- right now we only use optimistic execution for atomic transactions - we accept only 1 atomic transaction per block (1 transaction/block)
- but we re-execute at host.FinalizeBlock

```c++
# Consensus contract (tendermint): buildBlockProposal
processResp = host.ProcessProposal(processReq)
-> doOptimisticExecution
-> resbegin = host.BeginBlock(finalizeReq)
-> metainfo = host.OptimisticExecution(processReq, processResp)
    -> respFinalizeBlock, err = bapp.OptimisticExecution(processReq, processResp)
    -> metainfo = mctx.GetExecutionMetaInfoEncoded
-> store metainfo in block data: RequestProcessProposalWithMetaInfo

# Consensus contract: startBlockFinalization
resp = host.FinalizeBlock(finalizeReq, metainfo)
-> host: mctx.SetExecutionMetaInfo - map[string][]byte
-> host: bapp.FinalizeBlockSimple(finalizeReq) !! we actually reexecute the transaction using the crosschain tx metainfo cache if exists
-> host: optimisticExecution.Reset()
```

## crosschain tx & atomic tx

- `ExecuteAtomicTx` , `ExecuteCrossChainTx` in wasmx.network.keeper
- call multichain_registry contract (core consensus) for cross-chain tx
- intermediate results are stored in the `metainfo` field `RequestProcessProposalWithMetaInfo`
- can be done only by nodes who run all the subchains involved
- one subchain is the leader: the chain on the higher level (with longer block time)
- only the leader starts the atomic tx
- all other chains halt and wait for the leader to finish their part
- all chains involved halt until their the atomic transaction is finalized
- if the atomic tx does not leave the first chain - it fails before calling the crosschain tx host API, we will find the error inside MsgExecuteAtomicTxResponse
- atomic transactions have the same hash on both chains; it is the same transaction that is sent on both chains
- TODO explain what happens when a subtx fails while in execution

```c++
```

## SubChains

* Mythos does not create subchains - see `HOOK_NEW_SUBCHAIN` in `DEFAULT_HOOKS_NONC` - the consensus contract does not create subchains. Subchains are created by level0 consensus, and the negotiation is done by the lobby contract between nodes, when they announce themselves as available to create a subchain.
* `<basename>_<level>_<evm_id>-<fork>`, e.g. `ptestp_1_1001-1`, `level2_2_1002-1`, `qtestq_1_1003-1`, `level2_2_1004-1`, `level3_3_1005-1`
* subchains of higher levels should have longer block times, they are used to store checkpoint hashes of lower level chains
