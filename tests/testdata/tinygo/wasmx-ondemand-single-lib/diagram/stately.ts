// @ts-nocheck

import { createMachine } from "xstate";

export const machine = createMachine({
  context: {
    maxLevel: 0,
    blockTimeout: "blockTimeoutInternal",
    currentLevel: 0,
    membersCount: 1,
    blockTimeoutInternal: 3000,
  },
  id: "OnDemand-Single",
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
            start: {
              target: "started",
              actions: {
                type: "StartNode",
              },
            },
            setup: {
              target: "unstarted",
              actions: {
                type: "setup",
              },
            },
            prestart: {
              target: "prestart",
            },
            setupNode: {
              target: "unstarted",
              actions: {
                type: "setupNode",
              },
            },
          },
        },
        started: {
          initial: "active",
          on: {
            start: {
              target: "started",
              actions: {
                type: "StartNode",
              },
            },
          },
          states: {
            active: {
              on: {
                newTransaction: {
                  target: "proposer",
                  actions: [
                    {
                      type: "addToMempool",
                    },
                    {
                      type: "sendNewTransactionResponse",
                    },
                  ],
                  guard: {
                    type: "ifNewTransaction",
                  },
                },
              },
            },
            proposer: {
              always: {
                target: "active",
                actions: {
                  type: "newBlock",
                },
              },
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
      },
    },
  },
}).withConfig({
  actions: {
    StartNode: function (context, event) {
      // Add your action code here
      // ...
    },
    setup: function (context, event) {
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
    sendNewTransactionResponse: function (context, event) {
      // Add your action code here
      // ...
    },
    newBlock: function (context, event) {
      // Add your action code here
      // ...
    },
  },
  guards: {
    ifNewTransaction: function (context, event) {
      // Add your guard condition here
      return true;
    },
  },
});
