// @ts-nocheck
import { createMachine } from "xstate";

export const machine = createMachine({
  context: {
    blockTimeout: "timeoutCommit",
    max_tx_bytes: 65536,
    timeoutCommit: 4000,
    max_block_gas: "20000000",
    timeoutMissingTxs: 4000,
    kayros_base_url: "",
    kayros_user_key: "",
    threshold_commit: 51,
    threshold_finalize: 75,
    genesis_uuid: "",
    data_type_id: "",
    max_block_tx: "",
  },
  id: "Kayros-P2P-1",
  initial: "uninitialized",
  states: {
    uninitialized: {
      on: {
        initialize: {
          target: "initialized",
        },
      },
    },
    initialized: {
      initial: "unstarted",
      states: {
        unstarted: {
          on: {
            setupNode: {
              target: "unstarted",
              actions: {
                type: "setupNode",
              },
            },
            prestart: {
              target: "prestart",
            },
            setup: {
              target: "unstarted",
              actions: {
                type: "setup",
              },
            },
            start: {
              target: "started",
              actions: [
                {
                  type: "connectPeers",
                },
                {
                  type: "connectRooms",
                },
                {
                  type: "requestBlockSync",
                },
              ],
            },
          },
        },
        prestart: {
          after: {
            "500": {
              target: "started",
            },
          },
        },
        started: {
          initial: "Node",
          on: {
            newTransaction: {
              actions: [
                {
                  type: "addToMempool",
                },
                {
                  type: "registerWithKayros",
                },
                {
                  type: "sendNewTransactionResponse",
                },
                {
                  type: "forwardMsgToChat",
                  params: {
                    protocolId: "mempool",
                  },
                },
              ],
              guard: {
                type: "ifNewTransaction",
              },
            },
            receiveStateSyncRequest: {
              actions: {
                type: "receiveStateSyncRequest",
              },
            },
            updateNode: {
              actions: {
                type: "updateNodeAndReturn",
              },
            },
          },
          states: {
            Node: {
              on: {
                becomeValidator: {
                  target: "Validator",
                  actions: [
                    {
                      type: "registerValidatorWithNetwork",
                    },
                    {
                      type: "requestBlockSync",
                    },
                  ],
                },
                receiveStateSyncResponse: {
                  actions: {
                    type: "receiveStateSyncResponse",
                  },
                },
                start: {
                  actions: [
                    {
                      type: "connectPeers",
                    },
                    {
                      type: "connectRooms",
                    },
                    {
                      type: "requestBlockSync",
                    },
                  ],
                },
                receiveCommit: {
                  actions: {
                    type: "receiveCommit",
                  },
                },
                receiveUpdateNodeResponse: {
                  actions: {
                    type: "receiveUpdateNodeResponse",
                  },
                },
              },
              always: {
                target: "Validator",
                actions: {
                  type: "registerValidatorWithNetwork",
                },
                guard: {
                  type: "ifNodeIsValidator",
                },
              },
            },
            Validator: {
              initial: "active",
              on: {
                stop: {
                  target: "#Kayros-P2P-1.stopped",
                },
                receiveUpdateNodeResponse: {
                  actions: {
                    type: "receiveUpdateNodeResponse",
                  },
                },
                start: {
                  target: "Validator",
                  actions: [
                    {
                      type: "connectPeers",
                    },
                    {
                      type: "connectRooms",
                    },
                    {
                      type: "registerValidatorWithNetwork",
                    },
                    {
                      type: "requestBlockSync",
                    },
                  ],
                },
                receiveCommit: {
                  actions: {
                    type: "receiveCommit",
                  },
                  description:
                    "Receive the block hashes from other validators and finalize the block. Rollback if 2/3 with another hash.",
                },
                receiveUpdateNodeRequest: {
                  actions: {
                    type: "receiveUpdateNodeRequest",
                  },
                },
                receiveStateSyncResponse: {
                  actions: [
                    {
                      type: "receiveStateSyncResponse",
                    },
                    {
                      type: "requestValidatorNodeInfoIfSynced",
                    },
                  ],
                },
              },
              states: {
                active: {
                  on: {
                    receiveMissingTransactions: {
                      target: "active",
                      actions: {
                        type: "receiveMissingTransactions",
                      },
                    },
                  },
                  after: {
                    timeoutMissingTxs: {
                      target: "propose",
                    },
                  },
                  always: {
                    target: "propose",
                    guard: {
                      type: "ifAllTransactions",
                    },
                  },
                  entry: [
                    {
                      type: "getKayrosTxs",
                    },
                    {
                      type: "cancelActiveIntervals",
                      params: {
                        after: "timeoutMissingTxs",
                      },
                    },
                  ],
                },
                propose: {
                  after: {
                    timeoutCommit: {
                      target: "active",
                    },
                  },
                  entry: [
                    {
                      type: "commitBlock",
                    },
                    {
                      type: "sendCommit",
                    },
                    {
                      type: "cancelActiveIntervals",
                      params: {
                        after: "timeoutCommit",
                      },
                    },
                  ],
                },
              },
            },
          },
        },
      },
    },
    stopped: {
      on: {
        restart: {
          target: "#Kayros-P2P-1.initialized.unstarted",
        },
      },
    },
  },
}).withConfig({
  actions: {
    getKayrosTxs: function (context, event) {
      // Add your action code here
      // ...
    },
    cancelActiveIntervals: function (context, event) {
      // Add your action code here
      // ...
    },
    commitBlock: function (context, event) {
      // Add your action code here
      // ...
    },
    sendCommit: function (context, event) {
      // Add your action code here
      // ...
    },
    setupNode: function (context, event) {
      // Add your action code here
      // ...
    },
    addToMempool: function (context, event) {
      // Add your action code here
      // ...
    },
    registerWithKayros: function (context, event) {
      // Add your action code here
      // ...
    },
    sendNewTransactionResponse: function (context, event) {
      // Add your action code here
      // ...
    },
    forwardMsgToChat: function (context, event) {
      // Add your action code here
      // ...
    },
    setup: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveUpdateNodeResponse: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveStateSyncRequest: function (context, event) {
      // Add your action code here
      // ...
    },
    updateNodeAndReturn: function (context, event) {
      // Add your action code here
      // ...
    },
    registerValidatorWithNetwork: function (context, event) {
      // Add your action code here
      // ...
    },
    requestBlockSync: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveStateSyncResponse: function (context, event) {
      // Add your action code here
      // ...
    },
    connectPeers: function (context, event) {
      // Add your action code here
      // ...
    },
    connectRooms: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveCommit: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveUpdateNodeRequest: function (context, event) {
      // Add your action code here
      // ...
    },
    requestValidatorNodeInfoIfSynced: function (context, event) {
      // Add your action code here
      // ...
    },
    receiveMissingTransactions: function (context, event) {
      // Add your action code here
      // ...
    },
  },
  guards: {
    ifNewTransaction: function (context, event) {
      // Add your guard condition here
      return true;
    },
    ifNodeIsValidator: function (context, event) {
      // Add your guard condition here
      return true;
    },
    ifAllTransactions: function (context, event) {
      // Add your guard condition here
      return true;
    },
  },
});
