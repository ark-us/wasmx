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
