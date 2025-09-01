# The System

```
cd mythos-wazero
make build
make install
```

* see [./wasmx/README.md](./wasmx/README.md) and [./wasmx/EXAMPLES.md](./wasmx/EXAMPLES.md) for how to initialize chains and execute transactions.

The chain should be compatible with both Ethereum and Cosmos browser wallets.

## Resources

* docs: https://wasmx.provable.dev
* discord chat: https://discord.gg/8W5jeBke4f

## Core Components

* core contracts: https://github.com/loredanacirstea/wasmx-as-contracts
* consensus state machines:
    * tendermintp2p: https://stately.ai/registry/editor/55492854-7285-4432-a9f4-92e9333dda9b?machineId=535d10d2-da31-4dfe-b6f7-df09942cda1d
    * raftp2p: https://stately.ai/registry/editor/28cb5c6b-ea62-4ee0-9438-fdf903777162?machineId=9a11eb42-c6a3-4db2-a9fd-148cda512bc8
    * avalanche: https://stately.ai/registry/editor/d23cf175-c676-4f72-9c1a-4a0229f6528f?machineId=8b82eeb5-b437-456c-a6f5-f30f36c27c1e
    * levels0: https://stately.ai/registry/editor/beca3faa-e368-4938-b48d-346fb45d7561?machineId=fec428b1-aacb-40fe-85f2-457d0581eef4&mode=design
    * levels0 block on demand: https://stately.ai/registry/editor/2e5c6d52-e197-48af-886d-7c81d9d16cc9?mode=design&machineId=c73273b3-e54f-411d-9267-490108b21f80
    * subchain lobby consensus: https://stately.ai/registry/editor/beca3faa-e368-4938-b48d-346fb45d7561?mode=design&machineId=c9e41142-4275-490c-8dbd-d1c1093479bd
* explorer: https://github.com/loredanacirstea/explorer-pingpub/tree/mythos-wasmx

## Contribute

Read [./CONTRIBUTE.md](./CONTRIBUTE.md)
