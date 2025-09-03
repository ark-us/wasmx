# wasmX

The only WASM-modular blockchain engine capable of regeneration and metamorphosis.

* Multi VM, language-agnostic, general & domain-specific language interpreters, variety of host APIs.
    * EVM, TinyGo, AssemblyScript, JavaScript, Python, Rust, FSMvm (finite state machine interpreter).
    * host APIs: wasmx (our core API), wasmx crosschain, wasmx multichain, wasmx consensus, ewasm, WASI adaptor, GRPC, libp2p, HTTP, SQL, KV dbs, SMTP, IMAP
    * multichain, crosschain, atomic crosschain transactions: each node can run multiple chains and subchains

Compatible with both Ethereum and Cosmos SDK wallets.

This is the most flexible blockchain engine with 90% WASM modules and 10% Golang host.

Core contracts were written in AssemblyScript, and the consensus protocols are FSM diagrams that are INTERPRETED by an FSM interpreter (also in AssemblyScript).

Current speed of execution bottleneck is JSON encoding/decoding for AssemblyScript (this may be improved by newer `json-as` versions once it achieves feature parity with the old version). Nonetheless, speed can be greatly improved contract by contract, now that a functional implementation exists. And a Tinygo implementation of the core contracts has already been started.

This is a self-funded effort: one lead software architect [@ctzurcanu](https://github.com/ctzurcanu), and one everything software engineer [@loredanacirstea](https://github.com/loredanacirstea).

## NOT PRODUCTION READY

You will encounter bugs and incomplete features.
There are known security issues. This is a work in progress.

## WHY?

Because we want to build the most flexible software. Our end goal is building a global-level digital voting platform that is mathematically verifiable.

## Go!

```
cd mythos-wazero
make build
make install
```

* see [./wasmx/README.md](./wasmx/README.md) and [./wasmx/EXAMPLES.md](./wasmx/EXAMPLES.md) for how to initialize chains and execute transactions.

### Join the Mythos Testnet

Incoming. Join our Discord.
Past testnet instructions at: https://github.com/loredanacirstea/tempreleases/tree/main/mythos-testnet

## Core Components

* core contracts: https://github.com/loredanacirstea/wasmx-as-contracts
* JavaScript sdk: https://github.com/ark-us/wasmxjs, https://github.com/ark-us/mythosjs
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

## License

See [./LICENSE](./LICENSE), [./COPYRIGHT](./COPYRIGHT) and [./MORAL_LICENSE](./MORAL_LICENSE)

The Moral License is not yet legally binding, but one of our projects is to make such a license legally binding and computable. And we will use wasmX to manage such licensing.

If you are using wasmX, let us know so we can showcase your project.

Know that we do not condone using wasmX for the creation of blockchains without embedded identity verification, or blockchains that cannot apply justice actions. wasmX's direction is towards supporting digital identity (government-issued or other solutions).

## Resources

* docs: https://wasmx.provable.dev
* discord chat: https://discord.gg/8W5jeBke4f
* technical demo videos: https://www.youtube.com/playlist?list=PL323JufuD9JCRcY0fSEdQCC-yj0qL6Gkt

Want help or want to help? Join our Discord. We hope those interested can help one another.

## Want a production-ready release?

Contribute effort and/or pay for prioritizing features/bug fixes.

