(module
(import "env" "ethereum_useGas" (func $useGas (param i64)))
(import "env" "ethereum_log" (func $log (param i32 i32 i32 i32 i32 i32 i32) ))
(import "env" "ethereum_getGasLeft" (func $getGasLeft  (result i64)))
(import "env" "ethereum_getAddress" (func $getAddress (param i32) ))
(import "env" "ethereum_getExternalBalance" (func $getExternalBalance (param i32 i32) ))
(import "env" "ethereum_getBalance" (func $getBalance (param i32) ))
(import "env" "ethereum_getChainId" (func $getChainId (param i32) ))
(import "env" "ethereum_getBaseFee" (func $getBaseFee (param i32) ))
(import "env" "ethereum_getTxOrigin" (func $getTxOrigin (param i32) ))
(import "env" "ethereum_getCaller" (func $getCaller (param i32) ))
(import "env" "ethereum_getCallValue" (func $getCallValue (param i32) ))
(import "env" "ethereum_getCallDataSize" (func $getCallDataSize  (result i32)))
(import "env" "ethereum_callDataCopy" (func $callDataCopy (param i32 i32 i32) ))
(import "env" "ethereum_getCodeSize" (func $getCodeSize  (result i32)))
(import "env" "ethereum_codeCopy" (func $codeCopy (param i32 i32 i32) ))
(import "env" "ethereum_getExternalCodeSize" (func $getExternalCodeSize (param i32) (result i32)))
(import "env" "ethereum_getExternalCodeHash" (func $getExternalCodeHash (param i32 i32) ))
(import "env" "ethereum_externalCodeCopy" (func $externalCodeCopy (param i32 i32 i32 i32) ))
(import "env" "ethereum_getTxGasPrice" (func $getTxGasPrice (param i32) ))
(import "env" "ethereum_getBlockHash" (func $getBlockHash (param i64 i32) ))
(import "env" "ethereum_getBlockCoinbase" (func $getBlockCoinbase (param i32) ))
(import "env" "ethereum_getBlockTimestamp" (func $getBlockTimestamp  (result i64)))
(import "env" "ethereum_getBlockNumber" (func $getBlockNumber  (result i64)))
(import "env" "ethereum_getBlockDifficulty" (func $getBlockDifficulty (param i32) ))
(import "env" "ethereum_getBlockGasLimit" (func $getBlockGasLimit  (result i64)))
(import "env" "ethereum_create" (func $create (param i32 i32 i32 i32) ))
(import "env" "ethereum_create2" (func $create2 (param i32 i32 i32 i32 i32) ))
(import "env" "ethereum_call" (func $call (param i64 i32 i32 i32 i32 i32 i32) (result i32)))
(import "env" "ethereum_callCode" (func $callCode (param i64 i32 i32 i32 i32 i32 i32) (result i32)))
(import "env" "ethereum_callDelegate" (func $callDelegate (param i64 i32 i32 i32 i32 i32) (result i32)))
(import "env" "ethereum_callStatic" (func $callStatic (param i64 i32 i32 i32 i32 i32) (result i32)))
(import "env" "ethereum_returnDataCopy" (func $returnDataCopy (param i32 i32 i32) ))
(import "env" "ethereum_getReturnDataSize" (func $getReturnDataSize  (result i32)))
(import "env" "ethereum_storageStore" (func $storageStore (param i32 i32) ))
(import "env" "ethereum_storageLoad" (func $storageLoad (param i32 i32) ))
(import "env" "ethereum_selfDestruct" (func $selfDestruct (param i32) ))
(import "env" "ethereum_finish" (func $finish (param i32 i32) ))
(import "env" "ethereum_revert" (func $revert (param i32 i32) ))


  (type $et12 (func))
  (func $ewasm_interface_version_1 (export "ewasm_interface_version_1") (export "requires_ewasm") (type $et12)
      (nop))

(global $sp (mut i32) (i32.const -32))

;; memory related global
(global $memstart i32  (i32.const 33832))
;; the number of 256 words stored in memory
(global $wordCount (mut i64) (i64.const 0))
;; what was charged for the last memory allocation
(global $prevMemCost (mut i64) (i64.const 0))

;; for SHL, SHR, SAR
(global $global_ (mut i64) (i64.const 0))
(global $global__1 (mut i64) (i64.const 0))
(global $global__2 (mut i64) (i64.const 0))

;; TODO: memory should only be 1, but can't resize right now
(memory 500)
(export "memory" (memory 0))

(func $global_get_sp  (export "GLOBAL_GET_SP") (result i32) (global.get $sp))
(func $global_set_sp  (export "GLOBAL_SET_SP") (param $newsp i32) (global.set $sp (local.get $newsp)))


(func $LOG  (export "LOG")
  (param $number i32)

  (local $offset i32)
  (local $offset0 i64)
  (local $offset1 i64)
  (local $offset2 i64)
  (local $offset3 i64)

  (local $length i32)
  (local $length0 i64)
  (local $length1 i64)
  (local $length2 i64)
  (local $length3 i64)

  (local.set $offset0 (i64.load          (global.get $sp)))
  (local.set $offset1 (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $offset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $offset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (local.set $length0 (i64.load (i32.sub (global.get $sp) (i32.const 32))))
  (local.set $length1 (i64.load (i32.sub (global.get $sp) (i32.const 24))))
  (local.set $length2 (i64.load (i32.sub (global.get $sp) (i32.const 16))))
  (local.set $length3 (i64.load (i32.sub (global.get $sp) (i32.const  8))))

  (local.set $offset
             (call $check_overflow (local.get $offset0)
                                   (local.get $offset1)
                                   (local.get $offset2)
                                   (local.get $offset3)))

  (local.set $length
             (call $check_overflow (local.get $length0)
                                   (local.get $length1)
                                   (local.get $length2)
                                   (local.get $length3)))

  (call $memusegas (local.get $offset) (local.get $length))

  (if (i32.eq (local.get $number) (i32.const 0))
    (then
      (call $log
             (local.get $offset)
             (local.get $length)
             (local.get $number)
             (i32.const  0)
             (i32.const  0)
             (i32.const  0)
             (i32.const  0))
    )
  )
  (if (i32.eq (local.get $number) (i32.const 1))
    (then
    (call $log
             (local.get $offset)
             (local.get $length)
             (local.get $number)
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  64)))
             (i32.const  0)
             (i32.const  0)
             (i32.const  0))
    )
  )
  (if (i32.eq (local.get $number) (i32.const 2))
    (then
    (call $log
             (local.get $offset)
             (local.get $length)
             (local.get $number)
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  64)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  96)))
             (i32.const  0)
             (i32.const  0))
    )
  )
  (if (i32.eq (local.get $number) (i32.const 3))
    (then
    (call $log
             (local.get $offset)
             (local.get $length)
             (local.get $number)
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  64)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  96)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const 128)))
             (i32.const  0))
    )
  )
  (if (i32.eq (local.get $number) (i32.const 4))
    (then
    (call $log
             (local.get $offset)
             (local.get $length)
             (local.get $number)
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  64)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const  96)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const 128)))
             (call $bswap_m256 (i32.sub (global.get $sp) (i32.const 160))))
    )
  )
)

;; stack:
;;  0: dataOffset
(func $CALLDATALOAD (export "CALLDATALOAD")
  (local $writeOffset i32)
  (local $writeOffset0 i64)
  (local $writeOffset1 i64)
  (local $writeOffset2 i64)
  (local $writeOffset3 i64)

  (local.set $writeOffset0 (i64.load (i32.add (global.get $sp) (i32.const  0))))
  (local.set $writeOffset1 (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $writeOffset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $writeOffset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (i64.store (i32.add (global.get $sp) (i32.const  0)) (i64.const 0))
  (i64.store (i32.add (global.get $sp) (i32.const  8)) (i64.const 0))
  (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.const 0))

  (local.set $writeOffset
             (call $check_overflow (local.get $writeOffset0)
                                   (local.get $writeOffset1)
                                   (local.get $writeOffset2)
                                   (local.get $writeOffset3)))

  (call $callDataCopy (global.get $sp) (local.get $writeOffset) (i32.const 32))
  ;; swap top stack item
  (drop (call $bswap_m256 (global.get $sp)))
)

;; generated by ./wasm/generateInterface.js
(func $GAS (export "GAS")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (call $getGasLeft))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $ADDRESS (export "ADDRESS")  (call $getAddress(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $BALANCE (export "BALANCE")  (call $getExternalBalance(call $bswap_m256 (global.get $sp))(global.get $sp))(drop (call $bswap_m256 (global.get $sp))))
;; generated by ./wasm/generateInterface.js
(func $SELFBALANCE (export "SELFBALANCE")  (call $getBalance(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $CHAINID (export "CHAINID")  (call $getChainId(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $BASEFEE (export "BASEFEE")  (call $getBaseFee(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $ORIGIN (export "ORIGIN")  (call $getTxOrigin(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $CALLER (export "CALLER")  (call $getCaller(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $CALLVALUE (export "CALLVALUE")  (call $getCallValue(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $CALLDATASIZE (export "CALLDATASIZE")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (i64.extend_i32_u (call $getCallDataSize)))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $CALLDATACOPY (export "CALLDATACOPY")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $callDataCopy(local.get $offset0)(call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8))))(local.get $length0)))
;; generated by ./wasm/generateInterface.js
(func $CODESIZE (export "CODESIZE")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (i64.extend_i32_u (call $getCodeSize)))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $CODECOPY (export "CODECOPY")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $codeCopy(local.get $offset0)(call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8))))(local.get $length0)))
;; generated by ./wasm/generateInterface.js
(func $EXTCODESIZE (export "EXTCODESIZE")  (i64.store (global.get $sp) (i64.extend_i32_u (call $getExternalCodeSize(call $bswap_m256 (global.get $sp)))))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 8)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $EXTCODEHASH (export "EXTCODEHASH")  (call $getExternalCodeHash(call $bswap_m256 (global.get $sp))(global.get $sp))(drop (call $bswap_m256 (global.get $sp))))
;; generated by ./wasm/generateInterface.js
(func $EXTCODECOPY (export "EXTCODECOPY")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -96)))
          (i64.load (i32.add (global.get $sp) (i32.const -88)))
          (i64.load (i32.add (global.get $sp) (i32.const -80)))
          (i64.load (i32.add (global.get $sp) (i32.const -72)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $externalCodeCopy(call $bswap_m256 (global.get $sp))(local.get $offset0)(call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40))))(local.get $length0)))
;; generated by ./wasm/generateInterface.js
(func $GASPRICE (export "GASPRICE")  (call $getTxGasPrice(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $BLOCKHASH   (call $getBlockHash(call $check_overflow_i64
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24))))(global.get $sp))(drop (call $bswap_m256 (global.get $sp))))
;; generated by ./wasm/generateInterface.js
(func $COINBASE (export "COINBASE")  (call $getBlockCoinbase(i32.add (global.get $sp) (i32.const 32)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const 32)))))
;; generated by ./wasm/generateInterface.js
(func $TIMESTAMP (export "TIMESTAMP")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (call $getBlockTimestamp))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $NUMBER (export "NUMBER")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (call $getBlockNumber))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $DIFFICULTY (export "DIFFICULTY")  (call $getBlockDifficulty(i32.add (global.get $sp) (i32.const 32))))
;; generated by ./wasm/generateInterface.js
(func $GASLIMIT (export "GASLIMIT")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (call $getBlockGasLimit))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $CREATE (export "CREATE")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $create(call $bswap_m256 (global.get $sp))(local.get $offset0)(local.get $length0)(i32.add (global.get $sp) (i32.const -64)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const -64)))))
;; generated by ./wasm/generateInterface.js
(func $CREATE2 (export "CREATE2")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $create2(call $bswap_m256 (global.get $sp))(local.get $offset0)(local.get $length0)(call $bswap_m256 (i32.add (global.get $sp) (i32.const -96)))(i32.add (global.get $sp) (i32.const -96)))(drop (call $bswap_m256 (i32.add (global.get $sp) (i32.const -96)))))
;; generated by ./wasm/generateInterface.js
(func $CALL (export "CALL")(local $offset0 i32)(local $length0 i32)(local $offset1 i32)(local $length1 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -96)))
          (i64.load (i32.add (global.get $sp) (i32.const -88)))
          (i64.load (i32.add (global.get $sp) (i32.const -80)))
          (i64.load (i32.add (global.get $sp) (i32.const -72)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -128)))
          (i64.load (i32.add (global.get $sp) (i32.const -120)))
          (i64.load (i32.add (global.get $sp) (i32.const -112)))
          (i64.load (i32.add (global.get $sp) (i32.const -104)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0)))(local.set $offset1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -160)))
          (i64.load (i32.add (global.get $sp) (i32.const -152)))
          (i64.load (i32.add (global.get $sp) (i32.const -144)))
          (i64.load (i32.add (global.get $sp) (i32.const -136)))))(local.set $length1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -192)))
          (i64.load (i32.add (global.get $sp) (i32.const -184)))
          (i64.load (i32.add (global.get $sp) (i32.const -176)))
          (i64.load (i32.add (global.get $sp) (i32.const -168)))))
    (call $memusegas (local.get $offset1) (local.get $length1))
    (local.set $offset1 (i32.add (global.get $memstart) (local.get $offset1))) (i64.store (i32.add (global.get $sp) (i32.const -192)) (i64.extend_i32_u (i32.eqz (call $call(call $check_overflow_i64
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24))))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -32)))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -64)))(local.get $offset0)(local.get $length0)(local.get $offset1)(local.get $length1)))))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const -168)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -176)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -184)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $CALLCODE (export "CALLCODE")(local $offset0 i32)(local $length0 i32)(local $offset1 i32)(local $length1 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -96)))
          (i64.load (i32.add (global.get $sp) (i32.const -88)))
          (i64.load (i32.add (global.get $sp) (i32.const -80)))
          (i64.load (i32.add (global.get $sp) (i32.const -72)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -128)))
          (i64.load (i32.add (global.get $sp) (i32.const -120)))
          (i64.load (i32.add (global.get $sp) (i32.const -112)))
          (i64.load (i32.add (global.get $sp) (i32.const -104)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0)))(local.set $offset1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -160)))
          (i64.load (i32.add (global.get $sp) (i32.const -152)))
          (i64.load (i32.add (global.get $sp) (i32.const -144)))
          (i64.load (i32.add (global.get $sp) (i32.const -136)))))(local.set $length1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -192)))
          (i64.load (i32.add (global.get $sp) (i32.const -184)))
          (i64.load (i32.add (global.get $sp) (i32.const -176)))
          (i64.load (i32.add (global.get $sp) (i32.const -168)))))
    (call $memusegas (local.get $offset1) (local.get $length1))
    (local.set $offset1 (i32.add (global.get $memstart) (local.get $offset1))) (i64.store (i32.add (global.get $sp) (i32.const -192)) (i64.extend_i32_u (i32.eqz (call $callCode(call $check_overflow_i64
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24))))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -32)))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -64)))(local.get $offset0)(local.get $length0)(local.get $offset1)(local.get $length1)))))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const -168)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -176)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -184)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $DELEGATECALL (export "DELEGATECALL")(local $offset0 i32)(local $length0 i32)(local $offset1 i32)(local $length1 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -96)))
          (i64.load (i32.add (global.get $sp) (i32.const -88)))
          (i64.load (i32.add (global.get $sp) (i32.const -80)))
          (i64.load (i32.add (global.get $sp) (i32.const -72)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0)))(local.set $offset1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -128)))
          (i64.load (i32.add (global.get $sp) (i32.const -120)))
          (i64.load (i32.add (global.get $sp) (i32.const -112)))
          (i64.load (i32.add (global.get $sp) (i32.const -104)))))(local.set $length1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -160)))
          (i64.load (i32.add (global.get $sp) (i32.const -152)))
          (i64.load (i32.add (global.get $sp) (i32.const -144)))
          (i64.load (i32.add (global.get $sp) (i32.const -136)))))
    (call $memusegas (local.get $offset1) (local.get $length1))
    (local.set $offset1 (i32.add (global.get $memstart) (local.get $offset1))) (i64.store (i32.add (global.get $sp) (i32.const -160)) (i64.extend_i32_u (i32.eqz (call $callDelegate(call $check_overflow_i64
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24))))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -32)))(local.get $offset0)(local.get $length0)(local.get $offset1)(local.get $length1)))))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const -136)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -144)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -152)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $STATICCALL (export "STATICCALL")(local $offset0 i32)(local $length0 i32)(local $offset1 i32)(local $length1 i32) (local.set $offset0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -96)))
          (i64.load (i32.add (global.get $sp) (i32.const -88)))
          (i64.load (i32.add (global.get $sp) (i32.const -80)))
          (i64.load (i32.add (global.get $sp) (i32.const -72)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0)))(local.set $offset1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -128)))
          (i64.load (i32.add (global.get $sp) (i32.const -120)))
          (i64.load (i32.add (global.get $sp) (i32.const -112)))
          (i64.load (i32.add (global.get $sp) (i32.const -104)))))(local.set $length1 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -160)))
          (i64.load (i32.add (global.get $sp) (i32.const -152)))
          (i64.load (i32.add (global.get $sp) (i32.const -144)))
          (i64.load (i32.add (global.get $sp) (i32.const -136)))))
    (call $memusegas (local.get $offset1) (local.get $length1))
    (local.set $offset1 (i32.add (global.get $memstart) (local.get $offset1))) (i64.store (i32.add (global.get $sp) (i32.const -160)) (i64.extend_i32_u (i32.eqz (call $callStatic(call $check_overflow_i64
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24))))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -32)))(local.get $offset0)(local.get $length0)(local.get $offset1)(local.get $length1)))))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const -136)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -144)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const -152)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $RETURNDATACOPY (export "RETURNDATACOPY")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -64)))
          (i64.load (i32.add (global.get $sp) (i32.const -56)))
          (i64.load (i32.add (global.get $sp) (i32.const -48)))
          (i64.load (i32.add (global.get $sp) (i32.const -40)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $returnDataCopy(local.get $offset0)(call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8))))(local.get $length0)))
;; generated by ./wasm/generateInterface.js
(func $RETURNDATASIZE (export "RETURNDATASIZE")  (i64.store (i32.add (global.get $sp) (i32.const 32)) (i64.extend_i32_u (call $getReturnDataSize)))
    ;; zero out mem
    (i64.store (i32.add (global.get $sp) (i32.const 56)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 48)) (i64.const 0))
    (i64.store (i32.add (global.get $sp) (i32.const 40)) (i64.const 0)))
;; generated by ./wasm/generateInterface.js
(func $SSTORE (export "SSTORE")  (call $storageStore(call $bswap_m256 (global.get $sp))(call $bswap_m256 (i32.add (global.get $sp) (i32.const -32)))))
;; generated by ./wasm/generateInterface.js
(func $SLOAD (export "SLOAD")  (call $storageLoad(call $bswap_m256 (global.get $sp))(global.get $sp))(drop (call $bswap_m256 (global.get $sp))))
;; generated by ./wasm/generateInterface.js
(func $SELFDESTRUCT (export "SELFDESTRUCT")  (call $selfDestruct(call $bswap_m256 (global.get $sp))))
;; generated by ./wasm/generateInterface.js
(func $RETURN (export "RETURN")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $finish(local.get $offset0)(local.get $length0)))
;; generated by ./wasm/generateInterface.js
(func $REVERT (export "REVERT")(local $offset0 i32)(local $length0 i32) (local.set $offset0 (call $check_overflow
          (i64.load (global.get $sp))
          (i64.load (i32.add (global.get $sp) (i32.const 8)))
          (i64.load (i32.add (global.get $sp) (i32.const 16)))
          (i64.load (i32.add (global.get $sp) (i32.const 24)))))(local.set $length0 (call $check_overflow
          (i64.load (i32.add (global.get $sp) (i32.const -32)))
          (i64.load (i32.add (global.get $sp) (i32.const -24)))
          (i64.load (i32.add (global.get $sp) (i32.const -16)))
          (i64.load (i32.add (global.get $sp) (i32.const -8)))))
    (call $memusegas (local.get $offset0) (local.get $length0))
    (local.set $offset0 (i32.add (global.get $memstart) (local.get $offset0))) (call $revert(local.get $offset0)(local.get $length0)))
(func $ADD (export "ADD")
  (local $sp i32)

  (local $a i64)
  (local $c i64)
  (local $d i64)
  (local $carry i64)

  (local.set $sp (global.get $sp))

  ;; d c b a
  ;; pop the stack
  (local.set $a (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $c (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d (i64.load (local.get $sp)))
  ;; decement the stack pointer
  (local.set $sp (i32.sub (local.get $sp) (i32.const 8)))

  ;; d
  (local.set $carry (i64.add (local.get $d) (i64.load (i32.sub (local.get $sp) (i32.const 24)))))
  ;; save d  to mem
  (i64.store (i32.sub (local.get $sp) (i32.const 24)) (local.get $carry))
  ;; check  for overflow
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $carry) (local.get $d))))

  ;; c use $d as reg
  (local.set $d     (i64.add (i64.load (i32.sub (local.get $sp) (i32.const 16))) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $d) (local.get $carry))))
  (local.set $d     (i64.add (local.get $c) (local.get $d)))
  ;; store the result
  (i64.store (i32.sub (local.get $sp) (i32.const 16)) (local.get $d))
  ;; check overflow
  (local.set $carry (i64.or (i64.extend_i32_u  (i64.lt_u (local.get $d) (local.get $c))) (local.get $carry)))

  ;; b
  ;; add carry
  (local.set $d     (i64.add (i64.load (i32.sub (local.get $sp) (i32.const 8))) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $d) (local.get $carry))))

  ;; use reg c
  (local.set $c (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $d (i64.add (local.get $c) (local.get $d)))
  (i64.store (i32.sub (local.get $sp) (i32.const 8)) (local.get $d))
  ;; a
  (i64.store (local.get $sp)
             (i64.add        ;; add a
               (local.get $a)
               (i64.add
                 (i64.load (local.get $sp))  ;; load the operand
                 (i64.or  ;; carry
                   (i64.extend_i32_u (i64.lt_u (local.get $d) (local.get $c)))
                   (local.get $carry)))))
)

;; stack:
;;  0: A
;; -1: B
;; -2: MOD
(func $ADDMOD (export "ADDMOD")
  (local $sp i32)

  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  (local $moda i64)
  (local $modb i64)
  (local $modc i64)
  (local $modd i64)

  (local $carry i64)

  (local.set $sp (global.get $sp))

  ;; load args from the stack
  (local.set $a (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d (i64.load (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d1 (i64.load (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $moda (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $modb (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $modc (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $modd (i64.load (local.get $sp)))

  ;; a * 64^3 + b*64^2 + c*64 + d
  ;; d
  (local.set $d     (i64.add (local.get $d1) (local.get $d)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $d) (local.get $d1))))
  ;; c
  (local.set $c     (i64.add (local.get $c) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $c) (local.get $carry))))
  (local.set $c     (i64.add (local.get $c1) (local.get $c)))
  (local.set $carry (i64.or (i64.extend_i32_u  (i64.lt_u (local.get $c) (local.get $c1))) (local.get $carry)))
  ;; b
  (local.set $b     (i64.add (local.get $b) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $b) (local.get $carry))))
  (local.set $b     (i64.add (local.get $b1) (local.get $b)))
  (local.set $carry (i64.or (i64.extend_i32_u  (i64.lt_u (local.get $b) (local.get $b1))) (local.get $carry)))
  ;; a
  (local.set $a     (i64.add (local.get $a) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $a) (local.get $carry))))
  (local.set $a     (i64.add (local.get $a1) (local.get $a)))
  (local.set $carry (i64.or (i64.extend_i32_u  (i64.lt_u (local.get $a) (local.get $a1))) (local.get $carry)))

  (call $mod_320
        (local.get $carry) (local.get $a)    (local.get $b)    (local.get $c)    (local.get $d)
        (i64.const 0)      (local.get $moda) (local.get $modb) (local.get $modc) (local.get $modd) (local.get $sp))
)

(func $AND (export "AND")
  (i64.store (i32.sub (global.get $sp) (i32.const 8))  (i64.and (i64.load (i32.sub (global.get $sp) (i32.const 8)))  (i64.load (i32.add (global.get $sp) (i32.const 24)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 16)) (i64.and (i64.load (i32.sub (global.get $sp) (i32.const 16))) (i64.load (i32.add (global.get $sp) (i32.const 16)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 24)) (i64.and (i64.load (i32.sub (global.get $sp) (i32.const 24))) (i64.load (i32.add (global.get $sp) (i32.const 8)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 32)) (i64.and (i64.load (i32.sub (global.get $sp) (i32.const 32))) (i64.load (global.get $sp))))
)

;; stack:
;;  0: offset
;; -1: value
(func $BYTE (export "BYTE")
    (local $sp i32)

    (local $x1 i64)
    (local $x2 i64)
    (local $x3 i64)
    (local $x4 i64)
    (local $y1 i64)
    (local $y2 i64)
    (local $y3 i64)
    (local $y4 i64)

    (local $r1 i64)
    (local $r2 i64)
    (local $r3 i64)
    (local $r4 i64)
    (local $component i64)
    (local $condition i64)

    ;; load args from the stack
    (local.set $x1 (i64.load (i32.add (global.get $sp) (i32.const 24))))
    (local.set $x2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
    (local.set $x3 (i64.load (i32.add (global.get $sp) (i32.const 8))))
    (local.set $x4 (i64.load (global.get $sp)))

    (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

    (local.set $y1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
    (local.set $y2 (i64.load (i32.add (local.get $sp) (i32.const 16))))
    (local.set $y3 (i64.load (i32.add (local.get $sp) (i32.const 8))))
    (local.set $y4 (i64.load (local.get $sp)))

    (block
        (if (i64.eqz (i64.or (i64.or (local.get $x1) (local.get $x2)) (local.get $x3))) (then
            (nop)
            (block
                (local.set $condition (i64.div_u (local.get $x4) (i64.const 8)))
                (if (i64.eq (local.get $condition) (i64.const 0)) (then
                    (local.set $component (local.get $y1))
                )(else
                    (if (i64.eq (local.get $condition) (i64.const 1)) (then
                        (local.set $component (local.get $y2))
                    )(else
                        (if (i64.eq (local.get $condition) (i64.const 2)) (then
                            (local.set $component (local.get $y3))
                        )(else
                            (if (i64.eq (local.get $condition) (i64.const 3)) (then
                                (local.set $component (local.get $y4))
                            ))
                        ))
                    ))
                ))

            )
            (local.set $x4 (i64.mul (i64.rem_u (local.get $x4) (i64.const 8)) (i64.const 8)))
            (local.set $r4 (i64.shr_u (local.get $component) (i64.sub (i64.const 56) (local.get $x4))))
            (local.set $r4 (i64.and (i64.const 255) (local.get $r4)))
        ))

    )
    (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $r1))
    (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $r2))
    (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $r3))
    (i64.store          (local.get $sp)                 (local.get $r4))
)

(func $DIV (export "DIV")
  (local $sp i32)
  ;; dividend
  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  ;; divisor
  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $carry i32)
  (local $temp  i64)
  (local $temp2  i64)

  (local.set $sp (global.get $sp))
  (local.set $maskd (i64.const 1))

  ;; load args from the stack
  (local.set $a (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d (i64.load (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d1 (i64.load (local.get $sp)))

  (block $main
    ;; check div by 0
    (if (call $iszero_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
      (br $main)
    )

    ;; align bits
    (block $done
      (loop $loop
        ;; align bits;
        (if
          ;; check to make sure we are not overflowing
          (i32.or (i64.eqz (i64.clz (local.get $a1)))
          ;;  divisor < dividend
          (call $gte_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $a) (local.get $b) (local.get $c) (local.get $d)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.shl (local.get $d1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.shl (local.get $maskd) (i64.const 1)))

        (br $loop)
      )
    )


    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_256 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_256 (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $d) (local.get $d1)))
            (local.set $d     (i64.sub  (local.get $d) (local.get $d1)))
            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))
            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))
            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))

            ;; result = result + mask
            (local.set $dq   (i64.add (local.get $maskd) (local.get $dq)))
            (local.set $temp (i64.extend_i32_u (i64.lt_u (local.get $dq) (local.get $maskd))))
            (local.set $cq   (i64.add (local.get $cq) (local.get $temp)))
            (local.set $temp (i64.extend_i32_u (i64.lt_u (local.get $cq) (local.get $temp))))
            (local.set $cq   (i64.add (local.get $maskc) (local.get $cq)))
            (local.set $temp (i64.or (i64.extend_i32_u  (i64.lt_u (local.get $cq) (local.get $maskc))) (local.get $temp)))
            (local.set $bq   (i64.add (local.get $bq) (local.get $temp)))
            (local.set $temp (i64.extend_i32_u (i64.lt_u (local.get $bq) (local.get $temp))))
            (local.set $bq   (i64.add (local.get $maskb) (local.get $bq)))
            (local.set $aq   (i64.add (local.get $maska) (i64.add (local.get $aq) (i64.or (i64.extend_i32_u (i64.lt_u (local.get $bq) (local.get $maskb))) (local.get $temp)))))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  );; end of main

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $aq))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $bq))
  (i64.store (i32.add (local.get $sp) (i32.const 8))  (local.get $cq))
  (i64.store (local.get $sp) (local.get $dq))
)

(func $DUP (export "DUP")
  (param $a0 i32)
  (local $sp i32)

  (local $sp_ref i32)

  (local.set $sp (i32.add (global.get $sp) (i32.const 32)))
  (local.set $sp_ref (i32.sub (i32.sub (local.get $sp) (i32.const 8)) (i32.mul (local.get $a0) (i32.const 32))))

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.load (local.get $sp_ref)))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.load (i32.sub (local.get $sp_ref) (i32.const 8))))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.load (i32.sub (local.get $sp_ref) (i32.const 16))))
  (i64.store          (local.get $sp)                 (i64.load (i32.sub (local.get $sp_ref) (i32.const 24))))
)

(func $EQ (export "EQ")
  (local $sp i32)

  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))
  (i64.store (local.get $sp)
    (i64.extend_i32_u
      (i32.and (i64.eq   (i64.load (i32.add (local.get $sp) (i32.const 56))) (i64.load (i32.add (local.get $sp) (i32.const 24))))
      (i32.and (i64.eq   (i64.load (i32.add (local.get $sp) (i32.const 48))) (i64.load (i32.add (local.get $sp) (i32.const 16))))
      (i32.and (i64.eq   (i64.load (i32.add (local.get $sp) (i32.const 40))) (i64.load (i32.add (local.get $sp) (i32.const  8))))
               (i64.eq   (i64.load (i32.add (local.get $sp) (i32.const 32))) (i64.load          (local.get $sp))))))))

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.const 0))
)

(func $EXP (export "EXP")
  (local $sp i32)

  ;; base
  (local $base0 i64)
  (local $base1 i64)
  (local $base2 i64)
  (local $base3 i64)

  ;; exp
  (local $exp0 i64)
  (local $exp1 i64)
  (local $exp2 i64)
  (local $exp3 i64)

  (local $r0 i64)
  (local $r1 i64)
  (local $r2 i64)
  (local $r3 i64)

  (local $gasCounter i32)
  (local.set $sp (global.get $sp))

  ;; load args from the stack
  (local.set $base0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $base1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $base2 (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $base3 (i64.load          (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $exp0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $exp1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $exp2 (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $exp3 (i64.load          (local.get $sp)))

  ;; let result = new BN[1]
  (local.set $r3 (i64.const 1))

  (block $done
    (loop $loop
       ;; while [exp > 0] {
      (if (call $iszero_256 (local.get $exp0) (local.get $exp1) (local.get $exp2) (local.get $exp3))
        (br $done)
      )

      ;; if[exp.modn[2] === 1]
      ;; is odd?
      (if (i64.eqz (i64.ctz (local.get $exp3)))

        ;; result = result.mul[base].mod[TWO_POW256]
        ;; r = r * a
        (then
          (call $mul_256 (local.get $r0) (local.get $r1) (local.get $r2) (local.get $r3) (local.get $base0) (local.get $base1) (local.get $base2) (local.get $base3) (i32.add (local.get $sp) (i32.const 24)))
          (local.set $r0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
          (local.set $r1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
          (local.set $r2 (i64.load (i32.add (local.get $sp) (i32.const  8))))
          (local.set $r3 (i64.load          (local.get $sp)))
        )
      )
      ;; exp = exp.shrn 1
      (local.set $exp3 (i64.add (i64.shr_u (local.get $exp3) (i64.const 1)) (i64.shl (local.get $exp2) (i64.const 63))))
      (local.set $exp2 (i64.add (i64.shr_u (local.get $exp2) (i64.const 1)) (i64.shl (local.get $exp1) (i64.const 63))))
      (local.set $exp1 (i64.add (i64.shr_u (local.get $exp1) (i64.const 1)) (i64.shl (local.get $exp0) (i64.const 63))))
      (local.set $exp0 (i64.shr_u (local.get $exp0) (i64.const 1)))

      ;; base = base.mulr[baser].modr[TWO_POW256]
      (call $mul_256 (local.get $base0) (local.get $base1) (local.get $base2) (local.get $base3) (local.get $base0) (local.get $base1) (local.get $base2) (local.get $base3) (i32.add (local.get $sp) (i32.const 24)))
      (local.set $base0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
      (local.set $base1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
      (local.set $base2 (i64.load (i32.add (local.get $sp) (i32.const  8))))
      (local.set $base3 (i64.load          (local.get $sp)))

      (local.set $gasCounter (i32.add (local.get $gasCounter) (i32.const 1)))
      (br $loop)
    )
  )

  ;; use gas
  ;; Log256[Exponent] * 10
  (call $useGas
    (i64.extend_i32_u
      (i32.mul
        (i32.const 10)
        (i32.div_u
          (i32.add (local.get $gasCounter) (i32.const 7))
          (i32.const 8)))))

  ;; decement the stack pointer
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $r0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $r1))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $r2))
  (i64.store          (local.get $sp)                 (local.get $r3))
)

(func $GT (export "GT")
  (local $sp i32)

  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)
  (local $b0 i64)
  (local $b1 i64)
  (local $b2 i64)
  (local $b3 i64)

  (local.set $sp (global.get $sp))

  ;; load args from the stack
  (local.set $a0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $a3 (i64.load (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $b0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $b2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $b3 (i64.load (local.get $sp)))

  (i64.store (local.get $sp) (i64.extend_i32_u
    (i32.or (i64.gt_u (local.get $a0) (local.get $b0)) ;; a0 > b0
    (i32.and (i64.eq   (local.get $a0) (local.get $b0)) ;; a0 == a1
    (i32.or  (i64.gt_u (local.get $a1) (local.get $b1)) ;; a1 > b1
    (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
    (i32.or  (i64.gt_u (local.get $a2) (local.get $b2)) ;; a2 > b2
    (i32.and (i64.eq   (local.get $a2) (local.get $b2)) ;; a2 == b2
             (i64.gt_u (local.get $a3) (local.get $b3)))))))))) ;; a3 > b3

  ;; zero  out the rest of the stack item
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

(func $ISZERO (export "ISZERO")
  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)

  ;; load args from the stack
  (local.set $a0 (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $a3 (i64.load (global.get $sp)))

  (i64.store (global.get $sp)
    (i64.extend_i32_u
      (call $iszero_256 (local.get $a0) (local.get $a1) (local.get $a2) (local.get $a3))
    )
  )

  ;; zero out the rest of memory
  (i64.store (i32.add (global.get $sp) (i32.const 8)) (i64.const 0))
  (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.const 0))
)

(func $LT (export "LT")
  (local $sp i32)

  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)
  (local $b0 i64)
  (local $b1 i64)
  (local $b2 i64)
  (local $b3 i64)

  (local.set $sp (global.get $sp))

  ;; load args from the stack
  (local.set $a0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $a3 (i64.load (local.get $sp)))

  (local.set $sp (i32.sub (local.get $sp) (i32.const 32)))

  (local.set $b0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $b2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $b3 (i64.load (local.get $sp)))

  (i64.store (local.get $sp) (i64.extend_i32_u
    (i32.or  (i64.lt_u (local.get $a0) (local.get $b0)) ;; a0 < b0
    (i32.and (i64.eq   (local.get $a0) (local.get $b0)) ;; a0 == b0
    (i32.or  (i64.lt_u (local.get $a1) (local.get $b1)) ;; a1 < b1
    (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
    (i32.or  (i64.lt_u (local.get $a2) (local.get $b2)) ;; a2 < b2
    (i32.and (i64.eq   (local.get $a2) (local.get $b2)) ;; a2 == b2
             (i64.lt_u (local.get $a3) (local.get $b3)))))))))) ;; a3 < b3

  ;; zero  out the rest of the stack item
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

;; stack:
;;  0: offset
(func $MLOAD (export "MLOAD")
  (local $offset i32)
  (local $offset0 i64)
  (local $offset1 i64)
  (local $offset2 i64)
  (local $offset3 i64)

  ;; load args from the stack
  (local.set $offset0 (i64.load          (global.get $sp)))
  (local.set $offset1 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $offset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $offset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (local.set $offset
             (call $check_overflow (local.get $offset0)
                                   (local.get $offset1)
                                   (local.get $offset2)
                                   (local.get $offset3)))
  ;; subttract gas useage
  (call $memusegas (local.get $offset) (i32.const  32))

  ;; FIXME: how to deal with overflow?
  (local.set $offset (i32.add (local.get $offset) (global.get $memstart)))

  (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.load (i32.add (local.get $offset) (i32.const 24))))
  (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.load (i32.add (local.get $offset) (i32.const 16))))
  (i64.store (i32.add (global.get $sp) (i32.const  8)) (i64.load (i32.add (local.get $offset) (i32.const  8))))
  (i64.store          (global.get $sp)                 (i64.load          (local.get $offset)))

  ;; swap
  (drop (call $bswap_m256 (global.get $sp)))
)

(func $MOD (export "MOD")
  (local $sp i32)

  ;; dividend
  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  ;; divisor
  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $carry i32)
  (local $temp i64)

  (local.set $maskd (i64.const 1))

  ;; load args from the stack
  (local.set $a (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $d (i64.load          (global.get $sp)))
  ;; decement the stack pointer
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $d1 (i64.load          (local.get $sp)))


  (block $main
    ;; check div by 0
    (if (call $iszero_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
      (then
        (local.set $a (i64.const 0))
        (local.set $b (i64.const 0))
        (local.set $c (i64.const 0))
        (local.set $d (i64.const 0))
        (br $main)
      )
    )

    ;; align bits
    (block $done
        (loop $loop
        ;; align bits;
        (if (i32.or (i64.eqz (i64.clz (local.get $a1))) (call $gte_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $a) (local.get $b) (local.get $c) (local.get $d)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.shl (local.get $d1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.shl (local.get $maskd) (i64.const 1)))

        (br $loop)
      )
    )

    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_256 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_256 (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $d) (local.get $d1)))
            (local.set $d     (i64.sub  (local.get $d) (local.get $d1)))
            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))
            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))
            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  );; end of main

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $a))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $b))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $c))
  (i64.store          (local.get $sp)                 (local.get $d))
)

(func $MSIZE (export "MSIZE")
  (local $sp i32)

  ;; there's no input item for us to overwrite
  (local.set $sp (i32.add (global.get $sp) (i32.const 32)))

  (i64.store (i32.add (local.get $sp) (i32.const 0))
             (i64.mul (global.get $wordCount) (i64.const 32)))
  (i64.store (i32.add (local.get $sp) (i32.const 8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

;; stack:
;;  0: word
;; -1: offset
(func $MSTORE (export "MSTORE")
  (local $sp i32)

  (local $offset   i32)

  (local $offset0 i64)
  (local $offset1 i64)
  (local $offset2 i64)
  (local $offset3 i64)

  ;; load args from the stack
  (local.set $offset0 (i64.load          (global.get $sp)))
  (local.set $offset1 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $offset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $offset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (local.set $offset
             (call $check_overflow (local.get $offset0)
                                   (local.get $offset1)
                                   (local.get $offset2)
                                   (local.get $offset3)))
  ;; subtrace gas useage
  (call $memusegas (local.get $offset) (i32.const 32))

  ;; pop item from the stack
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  ;; swap top stack item
  (drop (call $bswap_m256 (local.get $sp)))

  (local.set $offset (i32.add (local.get $offset) (global.get $memstart)))
  ;; store word to memory
  (i64.store          (local.get $offset)                 (i64.load          (local.get $sp)))
  (i64.store (i32.add (local.get $offset) (i32.const 8))  (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (i64.store (i32.add (local.get $offset) (i32.const 16)) (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (i64.store (i32.add (local.get $offset) (i32.const 24)) (i64.load (i32.add (local.get $sp) (i32.const 24))))
)

;; stack:
;;  0: offset
;; -1: word
(func $MSTORE8 (export "MSTORE8")
  (local $sp i32)

  (local $offset i32)

  (local $offset0 i64)
  (local $offset1 i64)
  (local $offset2 i64)
  (local $offset3 i64)

  ;; load args from the stack
  (local.set $offset0 (i64.load          (global.get $sp)))
  (local.set $offset1 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $offset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $offset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (local.set $offset
             (call $check_overflow (local.get $offset0)
                                   (local.get $offset1)
                                   (local.get $offset2)
                                   (local.get $offset3)))

  (call $memusegas (local.get $offset) (i32.const 1))

  ;; pop stack
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))
  (local.set $offset (i32.add (local.get $offset) (global.get $memstart)))
  (i32.store8 (i32.add (local.get $offset) (i32.const 0)) (i32.load (local.get $sp)))
)

(func $MUL (export "MUL")
  (call $mul_256
        (i64.load (i32.add (global.get $sp) (i32.const 24)))
        (i64.load (i32.add (global.get $sp) (i32.const 16)))
        (i64.load (i32.add (global.get $sp) (i32.const  8)))
        (i64.load          (global.get $sp))
        (i64.load (i32.sub (global.get $sp) (i32.const  8)))
        (i64.load (i32.sub (global.get $sp) (i32.const 16)))
        (i64.load (i32.sub (global.get $sp) (i32.const 24)))
        (i64.load (i32.sub (global.get $sp) (i32.const 32)))
        (i32.sub (global.get $sp) (i32.const 8))
  )
)

(func $MULMOD (export "MULMOD")
  (local $sp i32)

  (local $a i64)
  (local $c i64)
  (local $e i64)
  (local $g i64)
  (local $i i64)
  (local $k i64)
  (local $m i64)
  (local $o i64)
  (local $b i64)
  (local $d i64)
  (local $f i64)
  (local $h i64)
  (local $j i64)
  (local $l i64)
  (local $n i64)
  (local $p i64)
  (local $temp7 i64)
  (local $temp6 i64)
  (local $temp5 i64)
  (local $temp4 i64)
  (local $temp3 i64)
  (local $temp2 i64)
  (local $temp1 i64)
  (local $temp0 i64)
  (local $rowCarry i64)

  (local $moda i64)
  (local $modb i64)
  (local $modc i64)
  (local $modd i64)

  ;; pop two items of the stack
  (local.set $a (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $c (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $e (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $g (i64.load          (global.get $sp)))
  (local.set $i (i64.load (i32.sub (global.get $sp) (i32.const  8))))
  (local.set $k (i64.load (i32.sub (global.get $sp) (i32.const 16))))
  (local.set $m (i64.load (i32.sub (global.get $sp) (i32.const 24))))
  (local.set $o (i64.load (i32.sub (global.get $sp) (i32.const 32))))

  (local.set $sp (i32.sub (global.get $sp) (i32.const 64)))

  ;; MUL
  ;;  a b c d e f g h
  ;;* i j k l m n o p
  ;;----------------

  ;; split the ops
  (local.set $b (i64.and (local.get $a) (i64.const 4294967295)))
  (local.set $a (i64.shr_u (local.get $a) (i64.const 32)))

  (local.set $d (i64.and (local.get $c) (i64.const 4294967295)))
  (local.set $c (i64.shr_u (local.get $c) (i64.const 32)))

  (local.set $f (i64.and (local.get $e) (i64.const 4294967295)))
  (local.set $e (i64.shr_u (local.get $e) (i64.const 32)))

  (local.set $h (i64.and (local.get $g) (i64.const 4294967295)))
  (local.set $g (i64.shr_u (local.get $g) (i64.const 32)))

  (local.set $j (i64.and (local.get $i) (i64.const 4294967295)))
  (local.set $i (i64.shr_u (local.get $i) (i64.const 32)))

  (local.set $l (i64.and (local.get $k) (i64.const 4294967295)))
  (local.set $k (i64.shr_u (local.get $k) (i64.const 32)))

  (local.set $n (i64.and (local.get $m) (i64.const 4294967295)))
  (local.set $m (i64.shr_u (local.get $m) (i64.const 32)))

  (local.set $p (i64.and (local.get $o) (i64.const 4294967295)))
  (local.set $o (i64.shr_u (local.get $o) (i64.const 32)))

   ;; first row multiplication
  ;; p * h
  (local.set $temp0 (i64.mul (local.get $p) (local.get $h)))
  ;; p * g + carry
  (local.set $temp1 (i64.add (i64.mul (local.get $p) (local.get $g)) (i64.shr_u (local.get $temp0) (i64.const 32))))
  ;; p * f + carry
  (local.set $temp2 (i64.add (i64.mul (local.get $p) (local.get $f)) (i64.shr_u (local.get $temp1) (i64.const 32))))
  ;; p * e + carry
  (local.set $temp3 (i64.add (i64.mul (local.get $p) (local.get $e)) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; p * d + carry
  (local.set $temp4 (i64.add (i64.mul (local.get $p) (local.get $d)) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; p * c + carry
  (local.set $temp5 (i64.add (i64.mul (local.get $p) (local.get $c)) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; p * b + carry
  (local.set $temp6 (i64.add (i64.mul (local.get $p) (local.get $b)) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; p * a + carry
  (local.set $temp7 (i64.add (i64.mul (local.get $p) (local.get $a)) (i64.shr_u (local.get $temp6) (i64.const 32))))
  (local.set $rowCarry (i64.shr_u (local.get $temp7) (i64.const 32)))

  ;; second row
  ;; o * h + $temp1
  (local.set $temp1 (i64.add (i64.mul (local.get $o) (local.get $h)) (i64.and (local.get $temp1) (i64.const 4294967295))))
  ;; o * g + $temp2 + carry
  (local.set $temp2 (i64.add (i64.add (i64.mul (local.get $o) (local.get $g)) (i64.and (local.get $temp2) (i64.const 4294967295))) (i64.shr_u (local.get $temp1) (i64.const 32))))
  ;; o * f + $temp3 + carry
  (local.set $temp3 (i64.add (i64.add (i64.mul (local.get $o) (local.get $f)) (i64.and (local.get $temp3) (i64.const 4294967295))) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; o * e + $temp4 + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $o) (local.get $e)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; o * d + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $o) (local.get $d)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; o * c + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $o) (local.get $c)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; o * b + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $o) (local.get $b)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; o * a + carry + rowCarry
  (local.set $p (i64.add (i64.add (i64.mul (local.get $o) (local.get $a)) (i64.shr_u (local.get $temp7) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $p) (i64.const 32)))

  ;; third row - n
  ;; n * h + $temp2
  (local.set $temp2 (i64.add (i64.mul (local.get $n) (local.get $h)) (i64.and (local.get $temp2) (i64.const 4294967295))))
  ;; n * g + $temp3  carry
  (local.set $temp3 (i64.add (i64.add (i64.mul (local.get $n) (local.get $g)) (i64.and (local.get $temp3) (i64.const 4294967295))) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; n * f + $temp4) + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $n) (local.get $f)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; n * e + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $n) (local.get $e)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; n * d + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $n) (local.get $d)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; n * c + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $n) (local.get $c)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; n * b + $p + carry
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $n) (local.get $b)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; n * a + carry
  (local.set $o (i64.add (i64.add (i64.mul (local.get $n) (local.get $a)) (i64.shr_u (local.get $p) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $o) (i64.const 32)))

  ;; forth row
  ;; m * h + $temp3
  (local.set $temp3 (i64.add (i64.mul (local.get $m) (local.get $h)) (i64.and (local.get $temp3) (i64.const 4294967295))))
  ;; m * g + $temp4 + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $m) (local.get $g)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; m * f + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $m) (local.get $f)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; m * e + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $m) (local.get $e)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; m * d + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $m) (local.get $d)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; m * c + $p + carry
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $m) (local.get $c)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; m * b + $o + carry
  (local.set $o     (i64.add (i64.add (i64.mul (local.get $m) (local.get $b)) (i64.and (local.get $o)     (i64.const 4294967295))) (i64.shr_u (local.get $p)     (i64.const 32))))
  ;; m * a + carry + rowCarry
  (local.set $n     (i64.add (i64.add (i64.mul (local.get $m) (local.get $a)) (i64.shr_u (local.get $o) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $n) (i64.const 32)))

  ;; fith row
  ;; l * h + $temp4
  (local.set $temp4 (i64.add (i64.mul (local.get $l) (local.get $h)) (i64.and (local.get $temp4) (i64.const 4294967295))))
  ;; l * g + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $l) (local.get $g)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; l * f + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $l) (local.get $f)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; l * e + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $l) (local.get $e)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; l * d + $p + carry
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $l) (local.get $d)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; l * c + $o + carry
  (local.set $o     (i64.add (i64.add (i64.mul (local.get $l) (local.get $c)) (i64.and (local.get $o)     (i64.const 4294967295))) (i64.shr_u (local.get $p)     (i64.const 32))))
  ;; l * b + $n + carry
  (local.set $n     (i64.add (i64.add (i64.mul (local.get $l) (local.get $b)) (i64.and (local.get $n)     (i64.const 4294967295))) (i64.shr_u (local.get $o)     (i64.const 32))))
  ;; l * a + carry + rowCarry
  (local.set $m     (i64.add (i64.add (i64.mul (local.get $l) (local.get $a)) (i64.shr_u (local.get $n) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $m) (i64.const 32)))

  ;; sixth row
  ;; k * h + $temp5
  (local.set $temp5 (i64.add (i64.mul (local.get $k) (local.get $h)) (i64.and (local.get $temp5) (i64.const 4294967295))))
  ;; k * g + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $k) (local.get $g)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; k * f + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $k) (local.get $f)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; k * e + $p + carry
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $k) (local.get $e)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; k * d + $o + carry
  (local.set $o     (i64.add (i64.add (i64.mul (local.get $k) (local.get $d)) (i64.and (local.get $o)     (i64.const 4294967295))) (i64.shr_u (local.get $p)     (i64.const 32))))
  ;; k * c + $n + carry
  (local.set $n     (i64.add (i64.add (i64.mul (local.get $k) (local.get $c)) (i64.and (local.get $n)     (i64.const 4294967295))) (i64.shr_u (local.get $o)     (i64.const 32))))
  ;; k * b + $m + carry
  (local.set $m     (i64.add (i64.add (i64.mul (local.get $k) (local.get $b)) (i64.and (local.get $m)     (i64.const 4294967295))) (i64.shr_u (local.get $n)     (i64.const 32))))
  ;; k * a + carry
  (local.set $l     (i64.add (i64.add (i64.mul (local.get $k) (local.get $a)) (i64.shr_u (local.get $m) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $l) (i64.const 32)))

  ;; seventh row
  ;; j * h + $temp6
  (local.set $temp6 (i64.add (i64.mul (local.get $j) (local.get $h)) (i64.and (local.get $temp6) (i64.const 4294967295))))
  ;; j * g + $temp7 + carry
  (local.set $temp7 (i64.add (i64.add (i64.mul (local.get $j) (local.get $g)) (i64.and (local.get $temp7) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; j * f + $p +carry
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $j) (local.get $f)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; j * e + $o + carry
  (local.set $o     (i64.add (i64.add (i64.mul (local.get $j) (local.get $e)) (i64.and (local.get $o)     (i64.const 4294967295))) (i64.shr_u (local.get $p)     (i64.const 32))))
  ;; j * d + $n + carry
  (local.set $n     (i64.add (i64.add (i64.mul (local.get $j) (local.get $d)) (i64.and (local.get $n)     (i64.const 4294967295))) (i64.shr_u (local.get $o)     (i64.const 32))))
  ;; j * c + $m + carry
  (local.set $m     (i64.add (i64.add (i64.mul (local.get $j) (local.get $c)) (i64.and (local.get $m)     (i64.const 4294967295))) (i64.shr_u (local.get $n)     (i64.const 32))))
  ;; j * b + $l + carry
  (local.set $l     (i64.add (i64.add (i64.mul (local.get $j) (local.get $b)) (i64.and (local.get $l)     (i64.const 4294967295))) (i64.shr_u (local.get $m)     (i64.const 32))))
  ;; j * a + carry
  (local.set $k     (i64.add (i64.add (i64.mul (local.get $j) (local.get $a)) (i64.shr_u (local.get $l) (i64.const 32))) (local.get $rowCarry)))
  (local.set $rowCarry (i64.shr_u (local.get $k) (i64.const 32)))

  ;; eigth row
  ;; i * h + $temp7
  (local.set $temp7 (i64.add (i64.mul (local.get $i) (local.get $h)) (i64.and (local.get $temp7) (i64.const 4294967295))))
  ;; i * g + $p
  (local.set $p     (i64.add (i64.add (i64.mul (local.get $i) (local.get $g)) (i64.and (local.get $p)     (i64.const 4294967295))) (i64.shr_u (local.get $temp7) (i64.const 32))))
  ;; i * f + $o + carry
  (local.set $o     (i64.add (i64.add (i64.mul (local.get $i) (local.get $f)) (i64.and (local.get $o)     (i64.const 4294967295))) (i64.shr_u (local.get $p)     (i64.const 32))))
  ;; i * e + $n + carry
  (local.set $n     (i64.add (i64.add (i64.mul (local.get $i) (local.get $e)) (i64.and (local.get $n)     (i64.const 4294967295))) (i64.shr_u (local.get $o)     (i64.const 32))))
  ;; i * d + $m + carry
  (local.set $m     (i64.add (i64.add (i64.mul (local.get $i) (local.get $d)) (i64.and (local.get $m)     (i64.const 4294967295))) (i64.shr_u (local.get $n)     (i64.const 32))))
  ;; i * c + $l + carry
  (local.set $l     (i64.add (i64.add (i64.mul (local.get $i) (local.get $c)) (i64.and (local.get $l)     (i64.const 4294967295))) (i64.shr_u (local.get $m)     (i64.const 32))))
  ;; i * b + $k + carry
  (local.set $k     (i64.add (i64.add (i64.mul (local.get $i) (local.get $b)) (i64.and (local.get $k)     (i64.const 4294967295))) (i64.shr_u (local.get $l)     (i64.const 32))))
  ;; i * a + carry
  (local.set $j     (i64.add (i64.add (i64.mul (local.get $i) (local.get $a)) (i64.shr_u (local.get $k) (i64.const 32))) (local.get $rowCarry)))

  ;; combine terms
  (local.set $a (local.get $j))
  (local.set $b (i64.or (i64.shl (local.get $k)     (i64.const 32)) (i64.and (local.get $l)     (i64.const 4294967295))))
  (local.set $c (i64.or (i64.shl (local.get $m)     (i64.const 32)) (i64.and (local.get $n)     (i64.const 4294967295))))
  (local.set $d (i64.or (i64.shl (local.get $o)     (i64.const 32)) (i64.and (local.get $p)     (i64.const 4294967295))))
  (local.set $e (i64.or (i64.shl (local.get $temp7) (i64.const 32)) (i64.and (local.get $temp6) (i64.const 4294967295))))
  (local.set $f (i64.or (i64.shl (local.get $temp5) (i64.const 32)) (i64.and (local.get $temp4) (i64.const 4294967295))))
  (local.set $g (i64.or (i64.shl (local.get $temp3) (i64.const 32)) (i64.and (local.get $temp2) (i64.const 4294967295))))
  (local.set $h (i64.or (i64.shl (local.get $temp1) (i64.const 32)) (i64.and (local.get $temp0) (i64.const 4294967295))))

  ;; pop the MOD argmunet off the stack
  (local.set $moda (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $modb (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $modc (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $modd (i64.load          (local.get $sp)))

  (call $mod_512
         (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $e) (local.get $f) (local.get $g) (local.get $h)
         (i64.const 0)  (i64.const 0) (i64.const 0)  (i64.const 0)  (local.get $moda) (local.get $modb) (local.get $modc) (local.get $modd) (i32.add (local.get $sp) (i32.const 24))
  )
)

(func $NOT (export "NOT")
  ;; FIXME: consider using 0xffffffffffffffff instead of -1?
  (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.xor (i64.load (i32.add (global.get $sp) (i32.const 24))) (i64.const -1)))
  (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.xor (i64.load (i32.add (global.get $sp) (i32.const 16))) (i64.const -1)))
  (i64.store (i32.add (global.get $sp) (i32.const  8)) (i64.xor (i64.load (i32.add (global.get $sp) (i32.const  8))) (i64.const -1)))
  (i64.store (i32.add (global.get $sp) (i32.const  0)) (i64.xor (i64.load (i32.add (global.get $sp) (i32.const  0))) (i64.const -1)))
)

(func $OR (export "OR")
  (i64.store (i32.sub (global.get $sp) (i32.const  8)) (i64.or (i64.load (i32.sub (global.get $sp) (i32.const  8))) (i64.load (i32.add (global.get $sp) (i32.const 24)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 16)) (i64.or (i64.load (i32.sub (global.get $sp) (i32.const 16))) (i64.load (i32.add (global.get $sp) (i32.const 16)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 24)) (i64.or (i64.load (i32.sub (global.get $sp) (i32.const 24))) (i64.load (i32.add (global.get $sp) (i32.const  8)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 32)) (i64.or (i64.load (i32.sub (global.get $sp) (i32.const 32))) (i64.load          (global.get $sp))))
)

(func $PC (export "PC")
  (param $pc i32)
  (local $sp i32)

  ;; add one to the stack
  (local.set $sp (i32.add (global.get $sp) (i32.const 32)))
  (i64.store (local.get $sp) (i64.extend_i32_u (local.get $pc)))

  ;; zero out rest of stack
  (i64.store (i32.add (local.get $sp) (i32.const 8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

(func $PUSH (export "PUSH")
  (param $a0 i64)
  (param $a1 i64)
  (param $a2 i64)
  (param $a3 i64)
  (local $sp i32)

  ;; increament stack pointer
  (local.set $sp (i32.add (global.get $sp) (i32.const 32)))

  (i64.store (local.get $sp) (local.get $a3))
  (i64.store (i32.add (local.get $sp) (i32.const 8)) (local.get $a2))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $a1))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $a0))
)

(func $SAR (export "SAR")
    (local $sp i32)
    (local $x1 i64)
    (local $x2 i64)
    (local $x3 i64)
    (local $x4 i64)
    (local $y1 i64)
    (local $y2 i64)
    (local $y3 i64)
    (local $y4 i64)

    (local $z1 i64)
    (local $z2 i64)
    (local $z3 i64)
    (local $z4 i64)

    ;; load args from the stack
    (local.set $x1 (i64.load (i32.add (global.get $sp) (i32.const 24))))
    (local.set $x2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
    (local.set $x3 (i64.load (i32.add (global.get $sp) (i32.const 8))))
    (local.set $x4 (i64.load (global.get $sp)))

    (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

    (local.set $y1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
    (local.set $y2 (i64.load (i32.add (local.get $sp) (i32.const 16))))
    (local.set $y3 (i64.load (i32.add (local.get $sp) (i32.const 8))))
    (local.set $y4 (i64.load (local.get $sp)))

    (block $label_sar_internal
        (if (i64.gt_u (i64.clz (local.get $y1)) (i64.const 0)) (then
            (block
                (local.set $z1 (call $shr_ (local.get $x1) (local.get $x2) (local.get $x3) (local.get $x4) (local.get $y1) (local.get $y2) (local.get $y3) (local.get $y4)))
                (local.set $z2 (global.get $global_))
                (local.set $z3 (global.get $global__1))
                (local.set $z4 (global.get $global__2))

            )
            (br $label_sar_internal)
        ))
        (if (call $gte_256x256_64 (local.get $x1) (local.get $x2) (local.get $x3) (local.get $x4) (i64.const 0) (i64.const 0) (i64.const 0) (i64.const 256)) (then
            (local.set $z1 (i64.const 18446744073709551615))
            (local.set $z2 (i64.const 18446744073709551615))
            (local.set $z3 (i64.const 18446744073709551615))
            (local.set $z4 (i64.const 18446744073709551615))
        ))
        (if (call $lt_256x256_64 (local.get $x1) (local.get $x2) (local.get $x3) (local.get $x4) (i64.const 0) (i64.const 0) (i64.const 0) (i64.const 256)) (then
            (block
                (local.set $y1 (call $shr_ (i64.const 0) (i64.const 0) (i64.const 0) (local.get $x4) (local.get $y1) (local.get $y2) (local.get $y3) (local.get $y4)))
                (local.set $y2 (global.get $global_))
                (local.set $y3 (global.get $global__1))
                (local.set $y4 (global.get $global__2))

            )
            (block
                (local.set $z1 (call $shl_ (i64.const 0) (i64.const 0) (i64.const 0) (i64.sub (i64.const 256) (local.get $x4)) (i64.const 18446744073709551615) (i64.const 18446744073709551615) (i64.const 18446744073709551615) (i64.const 18446744073709551615)))
                (local.set $z2 (global.get $global_))
                (local.set $z3 (global.get $global__1))
                (local.set $z4 (global.get $global__2))

            )
            (block
                (local.set $z1 (call $or_ (local.get $y1) (local.get $y2) (local.get $y3) (local.get $y4) (local.get $z1) (local.get $z2) (local.get $z3) (local.get $z4)))
                (local.set $z2 (global.get $global_))
                (local.set $z3 (global.get $global__1))
                (local.set $z4 (global.get $global__2))

            )
        ))

    )
    (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $z1))
    (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $z2))
    (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $z3))
    (i64.store          (local.get $sp)                 (local.get $z4))
)

(func $lt_256x256_64
    (param $x1 i64)
    (param $x2 i64)
    (param $x3 i64)
    (param $x4 i64)
    (param $y1 i64)
    (param $y2 i64)
    (param $y3 i64)
    (param $y4 i64)
    (result i32)
    (local $z i32)
    (local $condition_106 i32)
    (local $condition_107 i32)
    (local $condition_108 i32)
    (block
        (block
            (local.set $condition_106 (call $cmp (local.get $x1) (local.get $y1)))
            (if (i32.eq (local.get $condition_106) (i32.const 0)) (then
                (block
                    (local.set $condition_107 (call $cmp (local.get $x2) (local.get $y2)))
                    (if (i32.eq (local.get $condition_107) (i32.const 0)) (then
                        (block
                            (local.set $condition_108 (call $cmp (local.get $x3) (local.get $y3)))
                            (if (i32.eq (local.get $condition_108) (i32.const 0)) (then
                                (local.set $z (i64.lt_u (local.get $x4) (local.get $y4)))
                            )(else
                                (if (i32.eq (local.get $condition_108) (i32.const 1)) (then
                                    (local.set $z (i32.const 0))
                                )(else
                                    (local.set $z (i32.const 1))
                                ))
                            ))

                        )
                    )(else
                        (if (i32.eq (local.get $condition_107) (i32.const 1)) (then
                            (local.set $z (i32.const 0))
                        )(else
                            (local.set $z (i32.const 1))
                        ))
                    ))

                )
            )(else
                (if (i32.eq (local.get $condition_106) (i32.const 1)) (then
                    (local.set $z (i32.const 0))
                )(else
                    (local.set $z (i32.const 1))
                ))
            ))

        )

    )
    (local.get $z)
)

(func $gte_256x256_64
    (param $x1 i64)
    (param $x2 i64)
    (param $x3 i64)
    (param $x4 i64)
    (param $y1 i64)
    (param $y2 i64)
    (param $y3 i64)
    (param $y4 i64)
    (result i32)
    (local $z i32)
    (block
        (local.set $z (i32.eqz (call $lt_256x256_64 (local.get $x1) (local.get $x2) (local.get $x3) (local.get $x4) (local.get $y1) (local.get $y2) (local.get $y3) (local.get $y4))))

    )
    (local.get $z)
)


(func $or_
    (param $x1 i64)
    (param $x2 i64)
    (param $x3 i64)
    (param $x4 i64)
    (param $y1 i64)
    (param $y2 i64)
    (param $y3 i64)
    (param $y4 i64)
    (result i64)
    (local $r1 i64)
    (local $r2 i64)
    (local $r3 i64)
    (local $r4 i64)
    (block
        (local.set $r1 (i64.or (local.get $x1) (local.get $y1)))
        (local.set $r2 (i64.or (local.get $x2) (local.get $y2)))
        (local.set $r3 (i64.or (local.get $x3) (local.get $y3)))
        (local.set $r4 (i64.or (local.get $x4) (local.get $y4)))

    )
    (global.set $global_ (local.get $r2))
    (global.set $global__1 (local.get $r3))
    (global.set $global__2 (local.get $r4))
    (local.get $r1)
)


(func $cmp
    (param $a i64)
    (param $b i64)
    (result i32)
    (local $r i32)
    (block
        (local.set $r (select (i32.const 4294967295) (i64.ne (local.get $a) (local.get $b)) (i64.lt_u (local.get $a) (local.get $b))))

    )
    (local.get $r)
)


(func $shr_single_
    (param $a i64)
    (param $amount i64)
    (result i64)
    (local $x i64)
    (local $y i64)
    (block
        (local.set $y (i64.shl (local.get $a) (i64.sub (i64.const 64) (local.get $amount))))
        (local.set $x (i64.shr_u (local.get $a) (local.get $amount)))

    )
    (global.set $global_ (local.get $y))
    (local.get $x)
)

(func $shr_
    (param $x1 i64)
    (param $x2 i64)
    (param $x3 i64)
    (param $x4 i64)
    (param $y1 i64)
    (param $y2 i64)
    (param $y3 i64)
    (param $y4 i64)
    (result i64)
    (local $z1 i64)
    (local $z2 i64)
    (local $z3 i64)
    (local $z4 i64)
    (local $t i64)
    (block
        (if (i32.and (i64.eqz (local.get $x1)) (i64.eqz (local.get $x2))) (then
            (if (i64.eqz (local.get $x3)) (then
                (if (i64.eqz (local.get $x4))
                    (then
                        (local.set $z1 (local.get $y1))
                        (local.set $z2 (local.get $y2))
                        (local.set $z3 (local.get $y3))
                        (local.set $z4 (local.get $y4))
                    )
                    (else
                        (if (i64.lt_u (local.get $x4) (i64.const 256)) (then
                            (if (i64.ge_u (local.get $x4) (i64.const 128)) (then
                                (local.set $y4 (local.get $y2))
                                (local.set $y3 (local.get $y1))
                                (local.set $y2 (i64.const 0))
                                (local.set $y1 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 128)))
                            ))
                            (if (i64.ge_u (local.get $x4) (i64.const 64)) (then
                                (local.set $y4 (local.get $y3))
                                (local.set $y3 (local.get $y2))
                                (local.set $y2 (local.get $y1))
                                (local.set $y1 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 64)))
                            ))
                            (nop)
                            (block
                                (local.set $z4 (call $shr_single (local.get $y4) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (block
                                (local.set $z3 (call $shr_single (local.get $y3) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z4 (i64.or (local.get $z4) (local.get $t)))
                            (block
                                (local.set $z2 (call $shr_single (local.get $y2) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z3 (i64.or (local.get $z3) (local.get $t)))
                            (block
                                (local.set $z1 (call $shr_single (local.get $y1) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z2 (i64.or (local.get $z2) (local.get $t)))
                        ))
                    )
                )
            ))
        ))

    )
    (global.set $global_ (local.get $z2))
    (global.set $global__1 (local.get $z3))
    (global.set $global__2 (local.get $z4))
    (local.get $z1)
)


(func $shl_single_
    (param $a i64)
    (param $amount i64)
    (result i64)
    (local $x i64)
    (local $y i64)
    (block
        (local.set $x (i64.shr_u (local.get $a) (i64.sub (i64.const 64) (local.get $amount))))
        (local.set $y (i64.shl (local.get $a) (local.get $amount)))

    )
    (global.set $global_ (local.get $y))
    (local.get $x)
)

(func $shl_
    (param $x1 i64)
    (param $x2 i64)
    (param $x3 i64)
    (param $x4 i64)
    (param $y1 i64)
    (param $y2 i64)
    (param $y3 i64)
    (param $y4 i64)
    (result i64)
    (local $z1 i64)
    (local $z2 i64)
    (local $z3 i64)
    (local $z4 i64)
    (local $t i64)
    (local $r i64)
    (block
        (if (i32.and (i64.eqz (local.get $x1)) (i64.eqz (local.get $x2))) (then
            (if (i64.eqz (local.get $x3)) (then
                (if (i64.eqz (local.get $x4))
                    (then
                        (local.set $z1 (local.get $y1))
                        (local.set $z2 (local.get $y2))
                        (local.set $z3 (local.get $y3))
                        (local.set $z4 (local.get $y4))
                    )
                    (else
                        (if (i64.lt_u (local.get $x4) (i64.const 256)) (then
                            (if (i64.ge_u (local.get $x4) (i64.const 128)) (then
                                (local.set $y1 (local.get $y3))
                                (local.set $y2 (local.get $y4))
                                (local.set $y3 (i64.const 0))
                                (local.set $y4 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 128)))
                            ))
                            (if (i64.ge_u (local.get $x4) (i64.const 64)) (then
                                (local.set $y1 (local.get $y2))
                                (local.set $y2 (local.get $y3))
                                (local.set $y3 (local.get $y4))
                                (local.set $y4 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 64)))
                            ))
                            (nop)
                            (block
                                (local.set $t (call $shl_single (local.get $y4) (local.get $x4)))
                                (local.set $z4 (global.get $global_))

                            )
                            (block
                                (local.set $r (call $shl_single (local.get $y3) (local.get $x4)))
                                (local.set $z3 (global.get $global_))

                            )
                            (local.set $z3 (i64.or (local.get $z3) (local.get $t)))
                            (block
                                (local.set $t (call $shl_single (local.get $y2) (local.get $x4)))
                                (local.set $z2 (global.get $global_))

                            )
                            (local.set $z2 (i64.or (local.get $z2) (local.get $r)))
                            (block
                                (local.set $r (call $shl_single (local.get $y1) (local.get $x4)))
                                (local.set $z1 (global.get $global_))

                            )
                            (local.set $z1 (i64.or (local.get $z1) (local.get $t)))
                        ))
                    )
                )
            ))
        ))

    )
    (global.set $global_ (local.get $z2))
    (global.set $global__1 (local.get $z3))
    (global.set $global__2 (local.get $z4))
    (local.get $z1)
)

(func $SDIV (export "SDIV")
  (local $sp i32)

  ;; dividend
  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  ;; divisor
  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $carry i32)
  (local $temp  i64)
  (local $temp2 i64)
  (local $sign i32)

  (local.set $maskd (i64.const 1))

  ;; load args from the stack
  (local.set $a (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $d (i64.load (global.get $sp)))

  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $d1 (i64.load (local.get $sp)))

  ;; get the resulting sign
  (local.set $sign (i32.wrap_i64 (i64.shr_u (i64.xor (local.get $a1) (local.get $a)) (i64.const 63))))

  ;; convert to unsigned value
  (if (i64.eqz (i64.clz (local.get $a)))
    (then
      (local.set $a (i64.xor (local.get $a) (i64.const -1)))
      (local.set $b (i64.xor (local.get $b) (i64.const -1)))
      (local.set $c (i64.xor (local.get $c) (i64.const -1)))
      (local.set $d (i64.xor (local.get $d) (i64.const -1)))

      ;; a = a + 1
      (local.set $d (i64.add (local.get $d) (i64.const 1)))
      (local.set $carry (i64.eqz (local.get $d)))
      (local.set $c (i64.add (local.get $c) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $c)) (local.get $carry)))
      (local.set $b (i64.add (local.get $b) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $b)) (local.get $carry)))
      (local.set $a (i64.add (local.get $a) (i64.extend_i32_u (local.get $carry))))
    )
  )
  (if (i64.eqz (i64.clz (local.get $a1)))
    (then
      (local.set $a1 (i64.xor (local.get $a1) (i64.const -1)))
      (local.set $b1 (i64.xor (local.get $b1) (i64.const -1)))
      (local.set $c1 (i64.xor (local.get $c1) (i64.const -1)))
      (local.set $d1 (i64.xor (local.get $d1) (i64.const -1)))

      (local.set $d1 (i64.add (local.get $d1) (i64.const 1)))
      (local.set $carry (i64.eqz (local.get $d1)))
      (local.set $c1 (i64.add (local.get $c1) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $c1)) (local.get $carry)))
      (local.set $b1 (i64.add (local.get $b1) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $b1)) (local.get $carry)))
      (local.set $a1 (i64.add (local.get $a1) (i64.extend_i32_u (local.get $carry))))
    )
  )

  (block $main
    ;; check div by 0
    (if (call $iszero_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
      (br $main)
    )

    ;; align bits
    (block $done
      (loop $loop
        ;; align bits;
        (if (i32.or (i64.eqz (i64.clz (local.get $a1))) (call $gte_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $a) (local.get $b) (local.get $c) (local.get $d)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.shl (local.get $d1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.shl (local.get $maskd) (i64.const 1)))

        (br $loop)
      )
    )

    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_256 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_256 (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $d) (local.get $d1)))
            (local.set $d     (i64.sub  (local.get $d) (local.get $d1)))
            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))
            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))
            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))

            ;; result = result + mask
            (local.set $dq    (i64.add  (local.get $maskd) (local.get $dq)))
            (local.set $carry (i64.lt_u (local.get $dq) (local.get $maskd)))
            (local.set $temp  (i64.add  (local.get $cq) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.lt_u (local.get $temp) (local.get $cq)))
            (local.set $cq    (i64.add  (local.get $maskc) (local.get $temp)))
            (local.set $carry (i32.or   (i64.lt_u (local.get $cq) (local.get $maskc)) (local.get $carry)))
            (local.set $temp  (i64.add  (local.get $bq) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.lt_u (local.get $temp) (local.get $bq)))
            (local.set $bq    (i64.add  (local.get $maskb) (local.get $temp)))
            (local.set $carry (i32.or   (i64.lt_u (local.get $bq) (local.get $maskb)) (local.get $carry)))
            (local.set $aq    (i64.add  (local.get $maska) (i64.add (local.get $aq) (i64.extend_i32_u (local.get $carry)))))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  );; end of main

  ;; convert to signed
  (if (local.get $sign)
    (then
      (local.set $aq (i64.xor (local.get $aq) (i64.const -1)))
      (local.set $bq (i64.xor (local.get $bq) (i64.const -1)))
      (local.set $cq (i64.xor (local.get $cq) (i64.const -1)))
      (local.set $dq (i64.xor (local.get $dq) (i64.const -1)))

      (local.set $dq (i64.add (local.get $dq) (i64.const 1)))
      (local.set $cq (i64.add (local.get $cq) (i64.extend_i32_u (i64.eqz (local.get $dq)))))
      (local.set $bq (i64.add (local.get $bq) (i64.extend_i32_u (i64.eqz (local.get $cq)))))
      (local.set $aq (i64.add (local.get $aq) (i64.extend_i32_u (i64.eqz (local.get $bq)))))
    )
  )

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $aq))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $bq))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $cq))
  (i64.store          (local.get $sp)                 (local.get $dq))
)

(func $SGT (export "SGT")
  (local $sp i32)

  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)
  (local $b0 i64)
  (local $b1 i64)
  (local $b2 i64)
  (local $b3 i64)

  ;; load args from the stack
  (local.set $a0 (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $a3 (i64.load (global.get $sp)))

  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $b0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $b2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $b3 (i64.load (local.get $sp)))

  (i64.store (local.get $sp) (i64.extend_i32_u
    (i32.or  (i64.gt_s (local.get $a0) (local.get $b0)) ;; a0 > b0
    (i32.and (i64.eq   (local.get $a0) (local.get $b0)) ;; a0 == a1
    (i32.or  (i64.gt_u (local.get $a1) (local.get $b1)) ;; a1 > b1
    (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
    (i32.or  (i64.gt_u (local.get $a2) (local.get $b2)) ;; a2 > b2
    (i32.and (i64.eq   (local.get $a2) (local.get $b2)) ;; a2 == b2
             (i64.gt_u (local.get $a3) (local.get $b3)))))))))) ;; a3 > b3

  ;; zero  out the rest of the stack item
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

(func $SHA3 (export "SHA3")
  (local $dataOffset i32)
  (local $dataOffset0 i64)
  (local $dataOffset1 i64)
  (local $dataOffset2 i64)
  (local $dataOffset3 i64)

  (local $length i32)
  (local $length0 i64)
  (local $length1 i64)
  (local $length2 i64)
  (local $length3 i64)

  (local $contextOffset i32)
  (local $outputOffset i32)

  (local.set $length0 (i64.load (i32.sub (global.get $sp) (i32.const 32))))
  (local.set $length1 (i64.load (i32.sub (global.get $sp) (i32.const 24))))
  (local.set $length2 (i64.load (i32.sub (global.get $sp) (i32.const 16))))
  (local.set $length3 (i64.load (i32.sub (global.get $sp) (i32.const 8))))

  (local.set $dataOffset0 (i64.load (i32.add (global.get $sp) (i32.const 0))))
  (local.set $dataOffset1 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $dataOffset2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $dataOffset3 (i64.load (i32.add (global.get $sp) (i32.const 24))))

  (local.set $length
             (call $check_overflow (local.get $length0)
                                   (local.get $length1)
                                   (local.get $length2)
                                   (local.get $length3)))
  (local.set $dataOffset
             (call $check_overflow (local.get $dataOffset0)
                                   (local.get $dataOffset1)
                                   (local.get $dataOffset2)
                                   (local.get $dataOffset3)))

  ;; charge copy fee ceil(words/32) * 6
  (call $useGas (i64.extend_i32_u (i32.mul (i32.div_u (i32.add (local.get $length) (i32.const 31)) (i32.const 32)) (i32.const 6))))
  (call $memusegas (local.get $dataOffset) (local.get $length))

  (local.set $dataOffset (i32.add (global.get $memstart) (local.get $dataOffset)))

  (local.set $contextOffset (i32.const 32808))
  (local.set $outputOffset (i32.sub (global.get $sp) (i32.const 32)))

  (call $keccak (local.get $contextOffset) (local.get $dataOffset) (local.get $length) (local.get $outputOffset))

  (drop (call $bswap_m256 (local.get $outputOffset)))
)

(func $SHL (export "SHL")
    (local $sp i32)
    (local $x1 i64)
    (local $x2 i64)
    (local $x3 i64)
    (local $x4 i64)
    (local $y1 i64)
    (local $y2 i64)
    (local $y3 i64)
    (local $y4 i64)

    (local $z1 i64)
    (local $z2 i64)
    (local $z3 i64)
    (local $z4 i64)
    (local $t i64)
    (local $r i64)

    ;; load args from the stack
    (local.set $x1 (i64.load (i32.add (global.get $sp) (i32.const 24))))
    (local.set $x2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
    (local.set $x3 (i64.load (i32.add (global.get $sp) (i32.const 8))))
    (local.set $x4 (i64.load (global.get $sp)))

    (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

    (local.set $y1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
    (local.set $y2 (i64.load (i32.add (local.get $sp) (i32.const 16))))
    (local.set $y3 (i64.load (i32.add (local.get $sp) (i32.const 8))))
    (local.set $y4 (i64.load (local.get $sp)))

    (block
        (if (i32.and (i64.eqz (local.get $x1)) (i64.eqz (local.get $x2))) (then
            (if (i64.eqz (local.get $x3)) (then
                (if (i64.eqz (local.get $x4))
                    (then
                        (local.set $z1 (local.get $y1))
                        (local.set $z2 (local.get $y2))
                        (local.set $z3 (local.get $y3))
                        (local.set $z4 (local.get $y4))
                    )
                    (else
                        (if (i64.lt_u (local.get $x4) (i64.const 256)) (then
                            (if (i64.ge_u (local.get $x4) (i64.const 128)) (then
                                (local.set $y1 (local.get $y3))
                                (local.set $y2 (local.get $y4))
                                (local.set $y3 (i64.const 0))
                                (local.set $y4 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 128)))
                            ))
                            (if (i64.ge_u (local.get $x4) (i64.const 64)) (then
                                (local.set $y1 (local.get $y2))
                                (local.set $y2 (local.get $y3))
                                (local.set $y3 (local.get $y4))
                                (local.set $y4 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 64)))
                            ))
                            (nop)
                            (block
                                (local.set $t (call $shl_single (local.get $y4) (local.get $x4)))
                                (local.set $z4 (global.get $global_))

                            )
                            (block
                                (local.set $r (call $shl_single (local.get $y3) (local.get $x4)))
                                (local.set $z3 (global.get $global_))

                            )
                            (local.set $z3 (i64.or (local.get $z3) (local.get $t)))
                            (block
                                (local.set $t (call $shl_single (local.get $y2) (local.get $x4)))
                                (local.set $z2 (global.get $global_))

                            )
                            (local.set $z2 (i64.or (local.get $z2) (local.get $r)))
                            (block
                                (local.set $r (call $shl_single (local.get $y1) (local.get $x4)))
                                (local.set $z1 (global.get $global_))

                            )
                            (local.set $z1 (i64.or (local.get $z1) (local.get $t)))
                        ))
                    )
                )
            ))
        ))

    )
    (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $z1))
    (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $z2))
    (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $z3))
    (i64.store          (local.get $sp)                 (local.get $z4))
)

(func $shl_single
    (param $a i64)
    (param $amount i64)
    (result i64)
    (local $x i64)
    (local $y i64)
    (block
        (local.set $x (i64.shr_u (local.get $a) (i64.sub (i64.const 64) (local.get $amount))))
        (local.set $y (i64.shl (local.get $a) (local.get $amount)))

    )
    (global.set $global_ (local.get $y))
    (local.get $x)
)

(func $SHR (export "SHR")
    (local $sp i32)
    (local $x1 i64)
    (local $x2 i64)
    (local $x3 i64)
    (local $x4 i64)
    (local $y1 i64)
    (local $y2 i64)
    (local $y3 i64)
    (local $y4 i64)

    (local $z1 i64)
    (local $z2 i64)
    (local $z3 i64)
    (local $z4 i64)
    (local $t i64)

    ;; load args from the stack
    (local.set $x1 (i64.load (i32.add (global.get $sp) (i32.const 24))))
    (local.set $x2 (i64.load (i32.add (global.get $sp) (i32.const 16))))
    (local.set $x3 (i64.load (i32.add (global.get $sp) (i32.const 8))))
    (local.set $x4 (i64.load (global.get $sp)))

    (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

    (local.set $y1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
    (local.set $y2 (i64.load (i32.add (local.get $sp) (i32.const 16))))
    (local.set $y3 (i64.load (i32.add (local.get $sp) (i32.const 8))))
    (local.set $y4 (i64.load (local.get $sp)))

    (block
        (if (i32.and (i64.eqz (local.get $x1)) (i64.eqz (local.get $x2))) (then
            (if (i64.eqz (local.get $x3)) (then
                (if (i64.eqz (local.get $x4))
                    (then
                        (local.set $z1 (local.get $y1))
                        (local.set $z2 (local.get $y2))
                        (local.set $z3 (local.get $y3))
                        (local.set $z4 (local.get $y4))
                    )
                    (else
                        (if (i64.lt_u (local.get $x4) (i64.const 256)) (then
                            (if (i64.ge_u (local.get $x4) (i64.const 128)) (then
                                (local.set $y4 (local.get $y2))
                                (local.set $y3 (local.get $y1))
                                (local.set $y2 (i64.const 0))
                                (local.set $y1 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 128)))
                            ))
                            (if (i64.ge_u (local.get $x4) (i64.const 64)) (then
                                (local.set $y4 (local.get $y3))
                                (local.set $y3 (local.get $y2))
                                (local.set $y2 (local.get $y1))
                                (local.set $y1 (i64.const 0))
                                (local.set $x4 (i64.sub (local.get $x4) (i64.const 64)))
                            ))
                            (nop)
                            (block
                                (local.set $z4 (call $shr_single (local.get $y4) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (block
                                (local.set $z3 (call $shr_single (local.get $y3) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z4 (i64.or (local.get $z4) (local.get $t)))
                            (block
                                (local.set $z2 (call $shr_single (local.get $y2) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z3 (i64.or (local.get $z3) (local.get $t)))
                            (block
                                (local.set $z1 (call $shr_single (local.get $y1) (local.get $x4)))
                                (local.set $t (global.get $global_))

                            )
                            (local.set $z2 (i64.or (local.get $z2) (local.get $t)))
                        ))
                    )
                )
            ))
        ))

    )
    (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $z1))
    (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $z2))
    (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $z3))
    (i64.store          (local.get $sp)                 (local.get $z4))
)

(func $shr_single
    (param $a i64)
    (param $amount i64)
    (result i64)
    (local $x i64)
    (local $y i64)
    (block
        (local.set $y (i64.shl (local.get $a) (i64.sub (i64.const 64) (local.get $amount))))
        (local.set $x (i64.shr_u (local.get $a) (local.get $amount)))

    )
    (global.set $global_ (local.get $y))
    (local.get $x)
)

(func $SIGNEXTEND (export "SIGNEXTEND")
  (local $sp i32)

  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)

  (local $b0 i64)
  (local $b1 i64)
  (local $b2 i64)
  (local $b3 i64)
  (local $sign i64)
  (local $t i32)
  (local $end i32)

  (local.set $a0 (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $a3 (i64.load          (global.get $sp)))

  (local.set $end (global.get $sp))
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (if (i32.and
        (i32.and
          (i32.and
            (i64.lt_u (local.get $a3) (i64.const 32))
            (i64.eqz (local.get $a2)))
          (i64.eqz (local.get $a1)))
        (i64.eqz (local.get $a0)))
    (then
      (local.set $t (i32.add (i32.wrap_i64 (local.get $a3)) (local.get $sp)))
      (local.set $sign (i64.shr_s (i64.load8_s (local.get $t)) (i64.const 8)))
      (local.set $t (i32.add (local.get $t) (i32.const 1)))
      (block $done
        (loop $loop
          (if (i32.lt_u (local.get $end) (local.get $t))
            (br $done)
          )
          (i64.store (local.get $t) (local.get $sign))
          (local.set $t (i32.add (local.get $t) (i32.const 8)))
          (br $loop)
        )
      )
    )
  )
)


(func $SLT (export "SLT")
  (local $sp i32)

  (local $a0 i64)
  (local $a1 i64)
  (local $a2 i64)
  (local $a3 i64)
  (local $b0 i64)
  (local $b1 i64)
  (local $b2 i64)
  (local $b3 i64)

  ;; load args from the stack
  (local.set $a0 (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $a1 (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $a2 (i64.load (i32.add (global.get $sp) (i32.const 8))))
  (local.set $a3 (i64.load (global.get $sp)))

  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $b0 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $b2 (i64.load (i32.add (local.get $sp) (i32.const 8))))
  (local.set $b3 (i64.load (local.get $sp)))

  (i64.store (local.get $sp) (i64.extend_i32_u
    (i32.or  (i64.lt_s (local.get $a0) (local.get $b0)) ;; a0 < b0
    (i32.and (i64.eq   (local.get $a0) (local.get $b0)) ;; a0 == b0
    (i32.or  (i64.lt_u (local.get $a1) (local.get $b1)) ;; a1 < b1
    (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
    (i32.or  (i64.lt_u (local.get $a2) (local.get $b2)) ;; a2 < b2
    (i32.and (i64.eq   (local.get $a2) (local.get $b2)) ;; a2 == b2
             (i64.lt_u (local.get $a3) (local.get $b3)))))))))) ;; a3 < b3

  ;; zero  out the rest of the stack item
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (i64.const 0))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (i64.const 0))
)

(func $SMOD (export "SMOD")
  (local $sp i32)
  ;; dividend
  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  ;; divisor
  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $carry i32)
  (local $sign i32)
  (local $temp  i64)
  (local $temp2  i64)

  ;; load args from the stack
  (local.set $a (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $d (i64.load          (global.get $sp)))
  ;; decement the stack pointer
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $d1 (i64.load          (local.get $sp)))

  (local.set $maskd (i64.const 1))
  (local.set $sign (i32.wrap_i64 (i64.shr_u (local.get $d) (i64.const 63))))

  ;; convert to unsigned value
  (if (i64.eqz (i64.clz (local.get $a)))
    (then
      (local.set $a (i64.xor (local.get $a) (i64.const -1)))
      (local.set $b (i64.xor (local.get $b) (i64.const -1)))
      (local.set $c (i64.xor (local.get $c) (i64.const -1)))
      (local.set $d (i64.xor (local.get $d) (i64.const -1)))

      ;; a = a + 1
      (local.set $d (i64.add (local.get $d) (i64.const 1)))
      (local.set $carry (i64.eqz (local.get $d)))
      (local.set $c (i64.add (local.get $c) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $c)) (local.get $carry)))
      (local.set $b (i64.add (local.get $b) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $b)) (local.get $carry)))
      (local.set $a (i64.add (local.get $a) (i64.extend_i32_u (local.get $carry))))
    )
  )

  (if (i64.eqz (i64.clz (local.get $a1)))
    (then
      (local.set $a1 (i64.xor (local.get $a1) (i64.const -1)))
      (local.set $b1 (i64.xor (local.get $b1) (i64.const -1)))
      (local.set $c1 (i64.xor (local.get $c1) (i64.const -1)))
      (local.set $d1 (i64.xor (local.get $d1) (i64.const -1)))

      (local.set $d1 (i64.add (local.get $d1) (i64.const 1)))
      (local.set $carry (i64.eqz (local.get $d1)))
      (local.set $c1 (i64.add (local.get $c1) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $c1)) (local.get $carry)))
      (local.set $b1 (i64.add (local.get $b1) (i64.extend_i32_u (local.get $carry))))
      (local.set $carry (i32.and (i64.eqz (local.get $b1)) (local.get $carry)))
      (local.set $a1 (i64.add (local.get $a1) (i64.extend_i32_u (local.get $carry))))
    )
  )

  (block $main
    ;; check div by 0
    (if (call $iszero_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
      (then
        (local.set $a (i64.const 0))
        (local.set $b (i64.const 0))
        (local.set $c (i64.const 0))
        (local.set $d (i64.const 0))
        (br $main)
      )
    )

    ;; align bits
    (block $done
      (loop $loop
        ;; align bits;
        (if (i32.or (i64.eqz (i64.clz (local.get $a1))) (call $gte_256 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $a) (local.get $b) (local.get $c) (local.get $d)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.shl (local.get $d1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.shl (local.get $maskd) (i64.const 1)))

        (br $loop)
      )
    )

    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_256 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_256 (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $d) (local.get $d1)))
            (local.set $d     (i64.sub  (local.get $d) (local.get $d1)))
            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))
            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))
            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  )

  ;; convert to signed
  (if (local.get $sign)
    (then
      (local.set $a (i64.xor (local.get $a) (i64.const -1)))
      (local.set $b (i64.xor (local.get $b) (i64.const -1)))
      (local.set $c (i64.xor (local.get $c) (i64.const -1)))
      (local.set $d (i64.xor (local.get $d) (i64.const -1)))

      (local.set $d (i64.add (local.get $d) (i64.const 1)))
      (local.set $c (i64.add (local.get $c) (i64.extend_i32_u (i64.eqz (local.get $d)))))
      (local.set $b (i64.add (local.get $b) (i64.extend_i32_u (i64.eqz (local.get $c)))))
      (local.set $a (i64.add (local.get $a) (i64.extend_i32_u (i64.eqz (local.get $b)))))
    )
  )

  ;; save the stack
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $a))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $b))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $c))
  (i64.store          (local.get $sp)                 (local.get $d))
) ;; end for SMOD

(func $SUB (export "SUB")
  (local $sp i32)

  (local $a i64)
  (local $b i64)
  (local $c i64)
  (local $d i64)

  (local $a1 i64)
  (local $b1 i64)
  (local $c1 i64)
  (local $d1 i64)

  (local $carry i64)
  (local $temp i64)

  (local.set $a (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $b (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $c (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $d (i64.load          (global.get $sp)))
  ;; decement the stack pointer
  (local.set $sp (i32.sub (global.get $sp) (i32.const 32)))

  (local.set $a1 (i64.load (i32.add (local.get $sp) (i32.const 24))))
  (local.set $b1 (i64.load (i32.add (local.get $sp) (i32.const 16))))
  (local.set $c1 (i64.load (i32.add (local.get $sp) (i32.const  8))))
  (local.set $d1 (i64.load          (local.get $sp)))

  ;; a * 64^3 + b*64^2 + c*64 + d
  ;; d
  (local.set $carry (i64.extend_i32_u (i64.lt_u (local.get $d) (local.get $d1))))
  (local.set $d (i64.sub (local.get $d) (local.get $d1)))

  ;; c
  (local.set $temp (i64.sub (local.get $c) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.gt_u (local.get $temp) (local.get $c))))
  (local.set $c (i64.sub (local.get $temp) (local.get $c1)))
  (local.set $carry (i64.or (i64.extend_i32_u (i64.gt_u (local.get $c) (local.get $temp))) (local.get $carry)))

  ;; b
  (local.set $temp (i64.sub (local.get $b) (local.get $carry)))
  (local.set $carry (i64.extend_i32_u (i64.gt_u (local.get $temp) (local.get $b))))
  (local.set $b (i64.sub (local.get $temp) (local.get $b1)))

  ;; a
  (local.set $a (i64.sub (i64.sub (local.get $a) (i64.or (i64.extend_i32_u (i64.gt_u (local.get $b) (local.get $temp))) (local.get $carry))) (local.get $a1)))

  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $a))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $b))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (local.get $c))
  (i64.store          (local.get $sp)                 (local.get $d))
)

(func $SWAP (export "SWAP")
  (param $a0 i32)
  (local $sp_ref i32)

  (local $topa i64)
  (local $topb i64)
  (local $topc i64)
  (local $topd i64)

  (local.set $sp_ref (i32.sub (i32.add  (global.get $sp) (i32.const 24)) (i32.mul (i32.add (local.get $a0) (i32.const 1)) (i32.const 32))))

  (local.set $topa (i64.load (i32.add (global.get $sp) (i32.const 24))))
  (local.set $topb (i64.load (i32.add (global.get $sp) (i32.const 16))))
  (local.set $topc (i64.load (i32.add (global.get $sp) (i32.const  8))))
  (local.set $topd (i64.load          (global.get $sp)))

  ;; replace the top element
  (i64.store (i32.add (global.get $sp) (i32.const 24)) (i64.load (local.get $sp_ref)))
  (i64.store (i32.add (global.get $sp) (i32.const 16)) (i64.load (i32.sub (local.get $sp_ref) (i32.const 8))))
  (i64.store (i32.add (global.get $sp) (i32.const  8)) (i64.load (i32.sub (local.get $sp_ref) (i32.const 16))))
  (i64.store          (global.get $sp)                 (i64.load (i32.sub (local.get $sp_ref) (i32.const 24))))

  ;; store the old top element
  (i64.store (local.get $sp_ref)                          (local.get $topa))
  (i64.store (i32.sub (local.get $sp_ref) (i32.const 8))  (local.get $topb))
  (i64.store (i32.sub (local.get $sp_ref) (i32.const 16)) (local.get $topc))
  (i64.store (i32.sub (local.get $sp_ref) (i32.const 24)) (local.get $topd))
)

(func $XOR (export "XOR")
  (i64.store (i32.sub (global.get $sp) (i32.const  8)) (i64.xor (i64.load (i32.sub (global.get $sp) (i32.const  8))) (i64.load (i32.add (global.get $sp) (i32.const 24)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 16)) (i64.xor (i64.load (i32.sub (global.get $sp) (i32.const 16))) (i64.load (i32.add (global.get $sp) (i32.const 16)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 24)) (i64.xor (i64.load (i32.sub (global.get $sp) (i32.const 24))) (i64.load (i32.add (global.get $sp) (i32.const  8)))))
  (i64.store (i32.sub (global.get $sp) (i32.const 32)) (i64.xor (i64.load (i32.sub (global.get $sp) (i32.const 32))) (i64.load (i32.add (global.get $sp) (i32.const  0)))))
)

(func $bswap_i32
  (param $int i32)
  (result i32)

  (i32.or
    (i32.or
      (i32.and (i32.shr_u (local.get $int) (i32.const 24)) (i32.const 0xff)) ;; 7 -> 0
      (i32.and (i32.shr_u (local.get $int) (i32.const 8)) (i32.const 0xff00))) ;; 6 -> 1
    (i32.or
      (i32.and (i32.shl (local.get $int) (i32.const 8)) (i32.const 0xff0000)) ;; 5 -> 2
      (i32.and (i32.shl (local.get $int) (i32.const 24)) (i32.const 0xff000000)))) ;; 4 -> 3
)

(func $bswap_i64
  (param $int i64)
  (result i64)

  (i64.or
    (i64.or
      (i64.or
        (i64.and (i64.shr_u (local.get $int) (i64.const 56)) (i64.const 0xff)) ;; 7 -> 0
        (i64.and (i64.shr_u (local.get $int) (i64.const 40)) (i64.const 0xff00))) ;; 6 -> 1
      (i64.or
        (i64.and (i64.shr_u (local.get $int) (i64.const 24)) (i64.const 0xff0000)) ;; 5 -> 2
        (i64.and (i64.shr_u (local.get $int) (i64.const  8)) (i64.const 0xff000000)))) ;; 4 -> 3
    (i64.or
      (i64.or
        (i64.and (i64.shl (local.get $int) (i64.const 8))   (i64.const 0xff00000000)) ;; 3 -> 4
        (i64.and (i64.shl (local.get $int) (i64.const 24))   (i64.const 0xff0000000000))) ;; 2 -> 5
      (i64.or
        (i64.and (i64.shl (local.get $int) (i64.const 40))   (i64.const 0xff000000000000)) ;; 1 -> 6
        (i64.and (i64.shl (local.get $int) (i64.const 56))   (i64.const 0xff00000000000000))))) ;; 0 -> 7
)

(func $bswap_m128
  (param $sp i32)
  (result i32)
  (local $temp i64)

  (local.set $temp (call $bswap_i64 (i64.load (local.get $sp))))
  (i64.store (local.get $sp) (call $bswap_i64 (i64.load (i32.add (local.get $sp) (i32.const 8)))))
  (i64.store (i32.add (local.get $sp) (i32.const 8)) (local.get $temp))
  (local.get $sp)
)

(func $bswap_m160
  (param $sp i32)
  (result i32)
  (local $temp i64)

  (local.set $temp (call $bswap_i64 (i64.load (local.get $sp))))
  (i64.store (local.get $sp) (call $bswap_i64 (i64.load (i32.add (local.get $sp) (i32.const 12)))))
  (i64.store (i32.add (local.get $sp) (i32.const 12)) (local.get $temp))

  (i32.store (i32.add (local.get $sp) (i32.const 8)) (call $bswap_i32 (i32.load (i32.add (local.get $sp) (i32.const 8)))))
  (local.get $sp)
)

(func $bswap_m256
  (param $sp i32)
  (result i32)
  (local $temp i64)

  (local.set $temp (call $bswap_i64 (i64.load (local.get $sp))))
  (i64.store (local.get $sp) (call $bswap_i64 (i64.load (i32.add (local.get $sp) (i32.const 24)))))
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $temp))

  (local.set $temp (call $bswap_i64 (i64.load (i32.add (local.get $sp) (i32.const 8)))))
  (i64.store (i32.add (local.get $sp) (i32.const  8)) (call $bswap_i64 (i64.load (i32.add (local.get $sp) (i32.const 16)))))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $temp))
  (local.get $sp)
)

(func $check_overflow
  (param $a i64)
  (param $b i64)
  (param $c i64)
  (param $d i64)
  (result i32)

  (local $MAX_INT i32)
  (local.set $MAX_INT (i32.const -1))

  (if
    (i32.and
      (i32.and
        (i64.eqz  (local.get $d))
        (i64.eqz  (local.get $c)))
      (i32.and
        (i64.eqz  (local.get $b))
        (i64.lt_u (local.get $a) (i64.extend_i32_u (local.get $MAX_INT)))))
     (return (i32.wrap_i64 (local.get $a))))

     (return (local.get $MAX_INT))
)

(func $check_overflow_i64
  (param $a i64)
  (param $b i64)
  (param $c i64)
  (param $d i64)
  (result i64)

  (if
    (i32.and
      (i32.and
        (i64.eqz  (local.get $d))
        (i64.eqz  (local.get $c)))
      (i64.eqz  (local.get $b)))
    (return (local.get $a)))

    (return (i64.const 0xffffffffffffffff))
)

;; is a less than or equal to b // a >= b
(func $gte_256
  (param $a0 i64)
  (param $a1 i64)
  (param $a2 i64)
  (param $a3 i64)

  (param $b0 i64)
  (param $b1 i64)
  (param $b2 i64)
  (param $b3 i64)

  (result i32)

  ;; a0 > b0 || [a0 == b0 && [a1 > b1 || [a1 == b1 && [a2 > b2 || [a2 == b2 && a3 >= b3 ]]]]
  (i32.or  (i64.gt_u (local.get $a0) (local.get $b0)) ;; a0 > b0
  (i32.and (i64.eq   (local.get $a0) (local.get $b0))
  (i32.or  (i64.gt_u (local.get $a1) (local.get $b1)) ;; a1 > b1
  (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
  (i32.or  (i64.gt_u (local.get $a2) (local.get $b2)) ;; a2 > b2
  (i32.and (i64.eq   (local.get $a2) (local.get $b2))
           (i64.ge_u (local.get $a3) (local.get $b3))))))))
)

(func $gte_320
  (param $a0 i64)
  (param $a1 i64)
  (param $a2 i64)
  (param $a3 i64)
  (param $a4 i64)

  (param $b0 i64)
  (param $b1 i64)
  (param $b2 i64)
  (param $b3 i64)
  (param $b4 i64)

  (result i32)

  ;; a0 > b0 || [a0 == b0 && [a1 > b1 || [a1 == b1 && [a2 > b2 || [a2 == b2 && a3 >= b3 ]]]]
  (i32.or  (i64.gt_u (local.get $a0) (local.get $b0)) ;; a0 > b0
  (i32.and (i64.eq   (local.get $a0) (local.get $b0))
  (i32.or  (i64.gt_u (local.get $a1) (local.get $b1)) ;; a1 > b1
  (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
  (i32.or  (i64.gt_u (local.get $a2) (local.get $b2)) ;; a2 > b2
  (i32.and (i64.eq   (local.get $a2) (local.get $b2))
  (i32.or  (i64.gt_u (local.get $a3) (local.get $b3)) ;; a2 > b2
  (i32.and (i64.eq   (local.get $a3) (local.get $b3))
           (i64.ge_u (local.get $a4) (local.get $b4))))))))))
)

(func $gte_512
  (param $a0 i64)
  (param $a1 i64)
  (param $a2 i64)
  (param $a3 i64)
  (param $a4 i64)
  (param $a5 i64)
  (param $a6 i64)
  (param $a7 i64)

  (param $b0 i64)
  (param $b1 i64)
  (param $b2 i64)
  (param $b3 i64)
  (param $b4 i64)
  (param $b5 i64)
  (param $b6 i64)
  (param $b7 i64)

  (result i32)

  ;; a0 > b0 || [a0 == b0 && [a1 > b1 || [a1 == b1 && [a2 > b2 || [a2 == b2 && a3 >= b3 ]]]]
  (i32.or  (i64.gt_u (local.get $a0) (local.get $b0)) ;; a0 > b0
  (i32.and (i64.eq   (local.get $a0) (local.get $b0))
  (i32.or  (i64.gt_u (local.get $a1) (local.get $b1)) ;; a1 > b1
  (i32.and (i64.eq   (local.get $a1) (local.get $b1)) ;; a1 == b1
  (i32.or  (i64.gt_u (local.get $a2) (local.get $b2)) ;; a2 > b2
  (i32.and (i64.eq   (local.get $a2) (local.get $b2))
  (i32.or  (i64.gt_u (local.get $a3) (local.get $b3)) ;; a3 > b3
  (i32.and (i64.eq   (local.get $a3) (local.get $b3))
  (i32.or  (i64.gt_u (local.get $a4) (local.get $b4)) ;; a4 > b4
  (i32.and (i64.eq   (local.get $a4) (local.get $b4))
  (i32.or  (i64.gt_u (local.get $a5) (local.get $b5)) ;; a5 > b5
  (i32.and (i64.eq   (local.get $a5) (local.get $b5))
  (i32.or  (i64.gt_u (local.get $a6) (local.get $b6)) ;; a6 > b6
  (i32.and (i64.eq   (local.get $a6) (local.get $b6))
           (i64.ge_u (local.get $a7) (local.get $b7))))))))))))))))
)

(func $iszero_256
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (result i32)

  (i64.eqz (i64.or (i64.or (i64.or (local.get 0) (local.get 1)) (local.get 2)) (local.get 3)))
)

(func $iszero_320
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (result i32)

  (i64.eqz (i64.or (i64.or (i64.or (i64.or (local.get 0) (local.get 1)) (local.get 2)) (local.get 3)) (local.get 4)))
)

(func $iszero_512
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (param i64)
  (result i32)
  (i64.eqz (i64.or (i64.or (i64.or (i64.or (i64.or (i64.or (i64.or (local.get 0) (local.get 1)) (local.get 2)) (local.get 3)) (local.get 4)) (local.get 5)) (local.get 6)) (local.get 7)))
)

;;
;; Copied from https://github.com/axic/keccak-wasm (has more comments)
;;

(func $keccak_theta
  (param $context_offset i32)

  (local $C0 i64)
  (local $C1 i64)
  (local $C2 i64)
  (local $C3 i64)
  (local $C4 i64)
  (local $D0 i64)
  (local $D1 i64)
  (local $D2 i64)
  (local $D3 i64)
  (local $D4 i64)

  ;; C[x] = A[x] ^ A[x + 5] ^ A[x + 10] ^ A[x + 15] ^ A[x + 20];
  (local.set $C0
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 0)))
      (i64.xor
        (i64.load (i32.add (local.get $context_offset) (i32.const 40)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.const 80)))
          (i64.xor
            (i64.load (i32.add (local.get $context_offset) (i32.const 120)))
            (i64.load (i32.add (local.get $context_offset) (i32.const 160)))
          )
        )
      )
    )
  )

  (local.set $C1
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 8)))
      (i64.xor
        (i64.load (i32.add (local.get $context_offset) (i32.const 48)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.const 88)))
          (i64.xor
            (i64.load (i32.add (local.get $context_offset) (i32.const 128)))
            (i64.load (i32.add (local.get $context_offset) (i32.const 168)))
          )
        )
      )
    )
  )

  (local.set $C2
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 16)))
      (i64.xor
        (i64.load (i32.add (local.get $context_offset) (i32.const 56)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.const 96)))
          (i64.xor
            (i64.load (i32.add (local.get $context_offset) (i32.const 136)))
            (i64.load (i32.add (local.get $context_offset) (i32.const 176)))
          )
        )
      )
    )
  )

  (local.set $C3
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 24)))
      (i64.xor
        (i64.load (i32.add (local.get $context_offset) (i32.const 64)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.const 104)))
          (i64.xor
            (i64.load (i32.add (local.get $context_offset) (i32.const 144)))
            (i64.load (i32.add (local.get $context_offset) (i32.const 184)))
          )
        )
      )
    )
  )

  (local.set $C4
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 32)))
      (i64.xor
        (i64.load (i32.add (local.get $context_offset) (i32.const 72)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.const 112)))
          (i64.xor
            (i64.load (i32.add (local.get $context_offset) (i32.const 152)))
            (i64.load (i32.add (local.get $context_offset) (i32.const 192)))
          )
        )
      )
    )
  )

  ;; D[0] = ROTL64(C[1], 1) ^ C[4];
  (local.set $D0
    (i64.xor
      (local.get $C4)
      (i64.rotl
        (local.get $C1)
        (i64.const 1)
      )
    )
  )

  ;; D[1] = ROTL64(C[2], 1) ^ C[0];
  (local.set $D1
    (i64.xor
      (local.get $C0)
      (i64.rotl
        (local.get $C2)
        (i64.const 1)
      )
    )
  )

  ;; D[2] = ROTL64(C[3], 1) ^ C[1];
  (local.set $D2
    (i64.xor
      (local.get $C1)
      (i64.rotl
        (local.get $C3)
        (i64.const 1)
      )
    )
  )

  ;; D[3] = ROTL64(C[4], 1) ^ C[2];
  (local.set $D3
    (i64.xor
      (local.get $C2)
      (i64.rotl
        (local.get $C4)
        (i64.const 1)
      )
    )
  )

  ;; D[4] = ROTL64(C[0], 1) ^ C[3];
  (local.set $D4
    (i64.xor
      (local.get $C3)
      (i64.rotl
        (local.get $C0)
        (i64.const 1)
      )
    )
  )

  ;; A[x]      ^= D[x];
  ;; A[x + 5]  ^= D[x];
  ;; A[x + 10] ^= D[x];
  ;; A[x + 15] ^= D[x];
  ;; A[x + 20] ^= D[x];

  ;; x = 0
  (i64.store (i32.add (local.get $context_offset) (i32.const 0))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 0)))
      (local.get $D0)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 40))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 40)))
      (local.get $D0)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 80))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 80)))
      (local.get $D0)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 120))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 120)))
      (local.get $D0)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 160))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 160)))
      (local.get $D0)
    )
  )

  ;; x = 1
  (i64.store (i32.add (local.get $context_offset) (i32.const 8))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 8)))
      (local.get $D1)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 48))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 48)))
      (local.get $D1)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 88))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 88)))
      (local.get $D1)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 128))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 128)))
      (local.get $D1)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 168))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 168)))
      (local.get $D1)
    )
  )

  ;; x = 2
  (i64.store (i32.add (local.get $context_offset) (i32.const 16))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 16)))
      (local.get $D2)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 56))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 56)))
      (local.get $D2)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 96))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 96)))
      (local.get $D2)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 136))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 136)))
      (local.get $D2)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 176))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 176)))
      (local.get $D2)
    )
  )

  ;; x = 3
  (i64.store (i32.add (local.get $context_offset) (i32.const 24))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 24)))
      (local.get $D3)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 64))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 64)))
      (local.get $D3)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 104))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 104)))
      (local.get $D3)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 144))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 144)))
      (local.get $D3)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 184))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 184)))
      (local.get $D3)
    )
  )

  ;; x = 4
  (i64.store (i32.add (local.get $context_offset) (i32.const 32))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 32)))
      (local.get $D4)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 72))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 72)))
      (local.get $D4)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 112))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 112)))
      (local.get $D4)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 152))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 152)))
      (local.get $D4)
    )
  )

  (i64.store (i32.add (local.get $context_offset) (i32.const 192))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 192)))
      (local.get $D4)
    )
  )
)

(func $keccak_rho
  (param $context_offset i32)
  (param $rotation_consts i32)

  ;;(local $tmp i32)

  ;; state[ 1] = ROTL64(state[ 1],  1);
  ;;(local.set $tmp (i32.add (local.get $context_offset) (i32.const 1)))
  ;;(i64.store (local.get $tmp) (i64.rotl (i64.load (local.get $context_offset)) (i64.const 1)))

  ;;(local.set $tmp (i32.add (local.get $context_offset) (i32.const 2)))
  ;;(i64.store (local.get $tmp) (i64.rotl (i64.load (local.get $context_offset)) (i64.const 62)))

  (local $tmp i32)
  (local $i i32)

  ;; for (i = 0; i <= 24; i++)
  (local.set $i (i32.const 0))
  (block $done
    (loop $loop
      (if (i32.ge_u (local.get $i) (i32.const 24))
        (br $done)
      )

      (local.set $tmp (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (i32.const 1) (local.get $i)))))

      (i64.store (local.get $tmp) (i64.rotl (i64.load (local.get $tmp)) (i64.load8_u (i32.add (local.get $rotation_consts) (local.get $i)))))

      (local.set $i (i32.add (local.get $i) (i32.const 1)))
      (br $loop)
    )
  )
)

(func $keccak_pi
  (param $context_offset i32)

  (local $A1 i64)
  (local.set $A1 (i64.load (i32.add (local.get $context_offset) (i32.const 8))))

  ;; Swap non-overlapping fields, i.e. $A1 = $A6, etc.
  ;; NOTE: $A0 is untouched
  (i64.store (i32.add (local.get $context_offset) (i32.const 8)) (i64.load (i32.add (local.get $context_offset) (i32.const 48))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 48)) (i64.load (i32.add (local.get $context_offset) (i32.const 72))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 72)) (i64.load (i32.add (local.get $context_offset) (i32.const 176))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 176)) (i64.load (i32.add (local.get $context_offset) (i32.const 112))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 112)) (i64.load (i32.add (local.get $context_offset) (i32.const 160))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 160)) (i64.load (i32.add (local.get $context_offset) (i32.const 16))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 16)) (i64.load (i32.add (local.get $context_offset) (i32.const 96))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 96)) (i64.load (i32.add (local.get $context_offset) (i32.const 104))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 104)) (i64.load (i32.add (local.get $context_offset) (i32.const 152))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 152)) (i64.load (i32.add (local.get $context_offset) (i32.const 184))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 184)) (i64.load (i32.add (local.get $context_offset) (i32.const 120))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 120)) (i64.load (i32.add (local.get $context_offset) (i32.const 32))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 32)) (i64.load (i32.add (local.get $context_offset) (i32.const 192))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 192)) (i64.load (i32.add (local.get $context_offset) (i32.const 168))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 168)) (i64.load (i32.add (local.get $context_offset) (i32.const 64))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 64)) (i64.load (i32.add (local.get $context_offset) (i32.const 128))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 128)) (i64.load (i32.add (local.get $context_offset) (i32.const 40))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 40)) (i64.load (i32.add (local.get $context_offset) (i32.const 24))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 24)) (i64.load (i32.add (local.get $context_offset) (i32.const 144))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 144)) (i64.load (i32.add (local.get $context_offset) (i32.const 136))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 136)) (i64.load (i32.add (local.get $context_offset) (i32.const 88))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 88)) (i64.load (i32.add (local.get $context_offset) (i32.const 56))))
  (i64.store (i32.add (local.get $context_offset) (i32.const 56)) (i64.load (i32.add (local.get $context_offset) (i32.const 80))))

  ;; Place the previously saved overlapping field
  (i64.store (i32.add (local.get $context_offset) (i32.const 80)) (local.get $A1))
)

(func $keccak_chi
  (param $context_offset i32)

  (local $A0 i64)
  (local $A1 i64)
  (local $i i32)

  ;; for (round = 0; round < 25; i += 5)
  (local.set $i (i32.const 0))
  (block $done
    (loop $loop
      (if (i32.ge_u (local.get $i) (i32.const 25))
        (br $done)
      )

      (local.set $A0 (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (local.get $i)))))
      (local.set $A1 (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 1))))))

      ;; A[0 + i] ^= ~A1 & A[2 + i];
      (i64.store (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (local.get $i)))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (local.get $i))))
          (i64.and
            (i64.xor (local.get $A1) (i64.const 0xFFFFFFFFFFFFFFFF)) ;; bitwise not
            (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 2)))))
          )
        )
      )

      ;; A[1 + i] ^= ~A[2 + i] & A[3 + i];
      (i64.store (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 1))))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 1)))))
          (i64.and
            (i64.xor (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 2))))) (i64.const 0xFFFFFFFFFFFFFFFF)) ;; bitwise not
            (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 3)))))
          )
        )
      )

      ;; A[2 + i] ^= ~A[3 + i] & A[4 + i];
      (i64.store (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 2))))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 2)))))
          (i64.and
            (i64.xor (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 3))))) (i64.const 0xFFFFFFFFFFFFFFFF)) ;; bitwise not
            (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 4)))))
          )
        )
      )

      ;; A[3 + i] ^= ~A[4 + i] & A0;
      (i64.store (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 3))))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 3)))))
          (i64.and
            (i64.xor (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 4))))) (i64.const 0xFFFFFFFFFFFFFFFF)) ;; bitwise not
            (local.get $A0)
          )
        )
      )

      ;; A[4 + i] ^= ~A0 & A1;
      (i64.store (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 4))))
        (i64.xor
          (i64.load (i32.add (local.get $context_offset) (i32.mul (i32.const 8) (i32.add (local.get $i) (i32.const 4)))))
          (i64.and
            (i64.xor (local.get $A0) (i64.const 0xFFFFFFFFFFFFFFFF)) ;; bitwise not
            (local.get $A1)
          )
        )
      )

      (local.set $i (i32.add (local.get $i) (i32.const 5)))
      (br $loop)
    )
  )
)

(func $keccak_permute
  (param $context_offset i32)

  (local $rotation_consts i32)
  (local $round_consts i32)
  (local $round i32)

  (local.set $round_consts (i32.add (local.get $context_offset) (i32.const 400)))
  (local.set $rotation_consts (i32.add (local.get $context_offset) (i32.const 592)))

  ;; for (round = 0; round < 24; round++)
  (local.set $round (i32.const 0))
  (block $done
    (loop $loop
      (if (i32.ge_u (local.get $round) (i32.const 24))
        (br $done)
      )

      ;; theta transform
      (call $keccak_theta (local.get $context_offset))

      ;; rho transform
      (call $keccak_rho (local.get $context_offset) (local.get $rotation_consts))

      ;; pi transform
      (call $keccak_pi (local.get $context_offset))

      ;; chi transform
      (call $keccak_chi (local.get $context_offset))

      ;; iota transform
      ;; context_offset[0] ^= KECCAK_ROUND_CONSTANTS[round];
      (i64.store (local.get $context_offset)
        (i64.xor
          (i64.load (local.get $context_offset))
          (i64.load (i32.add (local.get $round_consts) (i32.mul (i32.const 8) (local.get $round))))
        )
      )

      (local.set $round (i32.add (local.get $round) (i32.const 1)))
      (br $loop)
    )
  )
)

(func $keccak_block
  (param $input_offset i32)
  (param $input_length i32) ;; ignored, we expect keccak256
  (param $context_offset i32)

  ;; read blocks in little-endian order and XOR against context_offset

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 0))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 0)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 0)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 8))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 8)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 8)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 16))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 16)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 16)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 24))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 24)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 24)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 32))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 32)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 32)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 40))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 40)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 40)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 48))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 48)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 48)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 56))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 56)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 56)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 64))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 64)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 64)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 72))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 72)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 72)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 80))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 80)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 80)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 88))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 88)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 88)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 96))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 96)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 96)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 104))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 104)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 104)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 112))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 112)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 112)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 120))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 120)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 120)))
    )
  )

  (i64.store
    (i32.add (local.get $context_offset) (i32.const 128))
    (i64.xor
      (i64.load (i32.add (local.get $context_offset) (i32.const 128)))
      (i64.load (i32.add (local.get $input_offset) (i32.const 128)))
    )
  )

  (call $keccak_permute (local.get $context_offset))
)

;;
;; Initialise the context
;;
(func $keccak_init
  (param $context_offset i32)
  (local $round_consts i32)
  (local $rotation_consts i32)

  (call $keccak_reset (local.get $context_offset))

  ;; insert the round constants (used by $KECCAK_IOTA)
  (local.set $round_consts (i32.add (local.get $context_offset) (i32.const 400)))
  (i64.store (i32.add (local.get $round_consts) (i32.const 0)) (i64.const 0x0000000000000001))
  (i64.store (i32.add (local.get $round_consts) (i32.const 8)) (i64.const 0x0000000000008082))
  (i64.store (i32.add (local.get $round_consts) (i32.const 16)) (i64.const 0x800000000000808A))
  (i64.store (i32.add (local.get $round_consts) (i32.const 24)) (i64.const 0x8000000080008000))
  (i64.store (i32.add (local.get $round_consts) (i32.const 32)) (i64.const 0x000000000000808B))
  (i64.store (i32.add (local.get $round_consts) (i32.const 40)) (i64.const 0x0000000080000001))
  (i64.store (i32.add (local.get $round_consts) (i32.const 48)) (i64.const 0x8000000080008081))
  (i64.store (i32.add (local.get $round_consts) (i32.const 56)) (i64.const 0x8000000000008009))
  (i64.store (i32.add (local.get $round_consts) (i32.const 64)) (i64.const 0x000000000000008A))
  (i64.store (i32.add (local.get $round_consts) (i32.const 72)) (i64.const 0x0000000000000088))
  (i64.store (i32.add (local.get $round_consts) (i32.const 80)) (i64.const 0x0000000080008009))
  (i64.store (i32.add (local.get $round_consts) (i32.const 88)) (i64.const 0x000000008000000A))
  (i64.store (i32.add (local.get $round_consts) (i32.const 96)) (i64.const 0x000000008000808B))
  (i64.store (i32.add (local.get $round_consts) (i32.const 104)) (i64.const 0x800000000000008B))
  (i64.store (i32.add (local.get $round_consts) (i32.const 112)) (i64.const 0x8000000000008089))
  (i64.store (i32.add (local.get $round_consts) (i32.const 120)) (i64.const 0x8000000000008003))
  (i64.store (i32.add (local.get $round_consts) (i32.const 128)) (i64.const 0x8000000000008002))
  (i64.store (i32.add (local.get $round_consts) (i32.const 136)) (i64.const 0x8000000000000080))
  (i64.store (i32.add (local.get $round_consts) (i32.const 144)) (i64.const 0x000000000000800A))
  (i64.store (i32.add (local.get $round_consts) (i32.const 152)) (i64.const 0x800000008000000A))
  (i64.store (i32.add (local.get $round_consts) (i32.const 160)) (i64.const 0x8000000080008081))
  (i64.store (i32.add (local.get $round_consts) (i32.const 168)) (i64.const 0x8000000000008080))
  (i64.store (i32.add (local.get $round_consts) (i32.const 176)) (i64.const 0x0000000080000001))
  (i64.store (i32.add (local.get $round_consts) (i32.const 184)) (i64.const 0x8000000080008008))

  ;; insert the rotation constants (used by $keccak_rho)
  (local.set $rotation_consts (i32.add (local.get $context_offset) (i32.const 592)))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 0)) (i32.const 1))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 1)) (i32.const 62))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 2)) (i32.const 28))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 3)) (i32.const 27))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 4)) (i32.const 36))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 5)) (i32.const 44))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 6)) (i32.const 6))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 7)) (i32.const 55))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 8)) (i32.const 20))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 9)) (i32.const 3))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 10)) (i32.const 10))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 11)) (i32.const 43))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 12)) (i32.const 25))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 13)) (i32.const 39))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 14)) (i32.const 41))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 15)) (i32.const 45))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 16)) (i32.const 15))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 17)) (i32.const 21))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 18)) (i32.const 8))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 19)) (i32.const 18))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 20)) (i32.const 2))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 21)) (i32.const 61))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 22)) (i32.const 56))
  (i32.store8 (i32.add (local.get $rotation_consts) (i32.const 23)) (i32.const 14))
)

;;
;; Reset the context
;;
(func $keccak_reset
  (param $context_offset i32)

  ;; clear out the context memory
  (drop (call $memset (local.get $context_offset) (i32.const 0) (i32.const 400)))
)

;;
;; Push input to the context
;;
(func $keccak_update
  (param $context_offset i32)
  (param $input_offset i32)
  (param $input_length i32)

  (local $residue_offset i32)
  (local $residue_buffer i32)
  (local $residue_index i32)
  (local $tmp i32)

  ;; this is where we store the pointer
  (local.set $residue_offset (i32.add (local.get $context_offset) (i32.const 200)))
  ;; this is where the buffer is
  (local.set $residue_buffer (i32.add (local.get $context_offset) (i32.const 208)))

  (local.set $residue_index (i32.load (local.get $residue_offset)))

  ;; process residue from last block
  (if (i32.ne (local.get $residue_index) (i32.const 0))
    (then
      ;; the space left in the residue buffer
      (local.set $tmp (i32.sub (i32.const 136) (local.get $residue_index)))

      ;; limit to what we have as an input
      (if (i32.lt_u (local.get $input_length) (local.get $tmp))
        (local.set $tmp (local.get $input_length))
      )

      ;; fill up the residue buffer
      (drop (call $memcpy
        (i32.add (local.get $residue_buffer) (local.get $residue_index))
        (local.get $input_offset)
        (local.get $tmp)
      ))

      (local.set $residue_index (i32.add (local.get $residue_index) (local.get $tmp)))

      ;; block complete
      (if (i32.eq (local.get $residue_index) (i32.const 136))
        (call $keccak_block (local.get $input_offset) (i32.const 136) (local.get $context_offset))

        (local.set $residue_index (i32.const 0))
      )

      (i32.store (local.get $residue_offset) (local.get $residue_index))

      (local.set $input_length (i32.sub (local.get $input_length) (local.get $tmp)))
    )
  )

  ;; while (input_length > block_size)
  (block $done
    (loop $loop
      (if (i32.lt_u (local.get $input_length) (i32.const 136))
        (br $done)
      )

      (call $keccak_block (local.get $input_offset) (i32.const 136) (local.get $context_offset))

      (local.set $input_offset (i32.add (local.get $input_offset) (i32.const 136)))
      (local.set $input_length (i32.sub (local.get $input_length) (i32.const 136)))
      (br $loop)
    )
  )

  ;; copy to the residue buffer
  (if (i32.gt_u (local.get $input_length) (i32.const 0))
    (then
      (drop (call $memcpy
        (i32.add (local.get $residue_buffer) (local.get $residue_index))
        (local.get $input_offset)
        (local.get $input_length)
      ))

      (local.set $residue_index (i32.add (local.get $residue_index) (local.get $input_length)))
      (i32.store (local.get $residue_offset) (local.get $residue_index))
    )
  )
)

;;
;; Finalise and return the hash
;;
;; The 256 bit hash is returned at the output offset.
;;
(func $keccak_finish
  (param $context_offset i32)
  (param $output_offset i32)

  (local $residue_offset i32)
  (local $residue_buffer i32)
  (local $residue_index i32)
  (local $tmp i32)

  ;; this is where we store the pointer
  (local.set $residue_offset (i32.add (local.get $context_offset) (i32.const 200)))
  ;; this is where the buffer is
  (local.set $residue_buffer (i32.add (local.get $context_offset) (i32.const 208)))

  (local.set $residue_index (i32.load (local.get $residue_offset)))
  (local.set $tmp (local.get $residue_index))

  ;; clear the rest of the residue buffer
  (drop (call $memset (i32.add (local.get $residue_buffer) (local.get $tmp)) (i32.const 0) (i32.sub (i32.const 136) (local.get $tmp))))

  ;; ((char*)ctx->message)[ctx->rest] |= 0x01;
  (local.set $tmp (i32.add (local.get $residue_buffer) (local.get $residue_index)))
  (i32.store8 (local.get $tmp) (i32.or (i32.load8_u (local.get $tmp)) (i32.const 0x01)))

  ;; ((char*)ctx->message)[block_size - 1] |= 0x80;
  (local.set $tmp (i32.add (local.get $residue_buffer) (i32.const 135)))
  (i32.store8 (local.get $tmp) (i32.or (i32.load8_u (local.get $tmp)) (i32.const 0x80)))

  (call $keccak_block (local.get $residue_buffer) (i32.const 136) (local.get $context_offset))

  ;; the first 32 bytes pointed at by $output_offset is the final hash
  (i64.store (local.get $output_offset) (i64.load (local.get $context_offset)))
  (i64.store (i32.add (local.get $output_offset) (i32.const 8)) (i64.load (i32.add (local.get $context_offset) (i32.const 8))))
  (i64.store (i32.add (local.get $output_offset) (i32.const 16)) (i64.load (i32.add (local.get $context_offset) (i32.const 16))))
  (i64.store (i32.add (local.get $output_offset) (i32.const 24)) (i64.load (i32.add (local.get $context_offset) (i32.const 24))))
)

;;
;; Calculate the hash. Helper method incorporating the above three.
;;
(func $keccak (export "KECCAK256")
  (param $context_offset i32)
  (param $input_offset i32)
  (param $input_length i32)
  (param $output_offset i32)

  (call $keccak_init (local.get $context_offset))
  (call $keccak_update (local.get $context_offset) (local.get $input_offset) (local.get $input_length))
  (call $keccak_finish (local.get $context_offset) (local.get $output_offset))
)

;;
;; memcpy from ewasm-libc/ewasm-cleanup
;;
(func $memcpy
  (param $dst i32)
  (param $src i32)
  (param $length i32)
  (result i32)

  (local $i i32)

  (local.set $i (i32.const 0))

  (block $done
    (loop $loop
      (if (i32.ge_u (local.get $i) (local.get $length))
        (br $done)
      )

      (i32.store8 (i32.add (local.get $dst) (local.get $i)) (i32.load8_u (i32.add (local.get $src) (local.get $i))))

      (local.set $i (i32.add (local.get $i) (i32.const 1)))
      (br $loop)
    )
  )

  (return (local.get $dst))
)

;;
;; memcpy from ewasm-libc/ewasm-cleanup
;;
(func $memset
  (param $ptr i32)
  (param $value i32)
  (param $length i32)
  (result i32)
  (local $i i32)

  (local.set $i (i32.const 0))

  (block $done
    (loop $loop
      (if (i32.ge_u (local.get $i) (local.get $length))
        (br $done)
      )

      (i32.store8 (i32.add (local.get $ptr) (local.get $i)) (local.get $value))

      (local.set $i (i32.add (local.get $i) (i32.const 1)))
      (br $loop)
    )
  )
  (local.get $ptr)
)

(func $memusegas
  (param $offset i32)
  (param $length i32)

  (local $cost i64)
  ;; the number of new words being allocated
  (local $newWordCount i64)

  (if (i32.eqz (local.get $length))
    (then (return))
  )

  ;; const newMemoryWordCount = Math.ceil[[offset + length] / 32]
  (local.set $newWordCount
    (i64.div_u (i64.add (i64.const 31) (i64.add (i64.extend_i32_u (local.get $offset)) (i64.extend_i32_u (local.get $length))))
               (i64.const 32)))

  ;;if [runState.highestMem >= highestMem]  return
  (if (i64.le_u (local.get $newWordCount) (global.get $wordCount))
    (then (return))
  )

  ;; words * 3 + words ^2 / 512
  (local.set $cost
     (i64.add
       (i64.mul (local.get $newWordCount) (i64.const 3))
       (i64.div_u
         (i64.mul (local.get $newWordCount)
                  (local.get $newWordCount))
         (i64.const 512))))

  (call $useGas  (i64.sub (local.get $cost) (global.get $prevMemCost)))
  (global.set $prevMemCost (local.get $cost))
  (global.set $wordCount (local.get $newWordCount))

  ;; grow actual memory
  ;; the first 31704 bytes are guaranteed to be available
  ;; adjust for 32 bytes  - the maximal size of MSTORE write
  ;; TODO it should be memory.size * page_size
  (local.set $offset (i32.add (local.get $length) (i32.add (local.get $offset) (global.get $memstart))))
  (if (i32.gt_u (local.get $offset) (i32.mul (i32.const 65536) (memory.size)))
    (then
      (drop (memory.grow
        (i32.div_u (i32.add (i32.const 65535) (i32.sub (local.get $offset) (memory.size))) (i32.const 65536))))
    )
  )
)

(func $mod_320
  ;; dividend
  (param $a i64)
  (param $b i64)
  (param $c i64)
  (param $d i64)
  (param $e i64)

  ;; divisor
  (param $a1 i64)
  (param $b1 i64)
  (param $c1 i64)
  (param $d1 i64)
  (param $e1 i64)

  ;; stack pointer
  (param $sp i32)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)
  (local $eq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $maske i64)

  (local $carry i32)
  (local $temp i64)

  (local.set $maske (i64.const 1))
  (block $main
    ;; check div by 0
    (if (call $iszero_320 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1))
      (then
        (local.set $a (i64.const 0))
        (local.set $b (i64.const 0))
        (local.set $c (i64.const 0))
        (local.set $d (i64.const 0))
        (local.set $e (i64.const 0))
        (br $main)
      )
    )

    (block $done
      ;; align bits
      (loop $loop
        ;; align bits;
        (if (i32.or (i64.eqz (i64.clz (local.get $a1))) (call $gte_320
                                                            (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1)
                                                            (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $e)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.add (i64.shl (local.get $d1) (i64.const 1)) (i64.shr_u (local.get $e1) (i64.const 63))))
        (local.set $e1 (i64.shl (local.get $e1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.add (i64.shl (local.get $maskd) (i64.const 1)) (i64.shr_u (local.get $maske) (i64.const 63))))
        (local.set $maske (i64.shl (local.get $maske) (i64.const 1)))
        (br $loop)
      )
    )

    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_320 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd) (local.get $maske))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_320 (local.get $a) (local.get $b) (local.get $c) (local.get $d) (local.get $e) (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $e) (local.get $e1)))
            (local.set $e     (i64.sub  (local.get $e) (local.get $e1)))

            (local.set $temp  (i64.sub  (local.get $d) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $d)))
            (local.set $d     (i64.sub  (local.get $temp) (local.get $d1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $d) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))

            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $e1 (i64.add (i64.shr_u (local.get $e1) (i64.const 1)) (i64.shl (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maske (i64.add (i64.shr_u (local.get $maske) (i64.const 1)) (i64.shl (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  );; end of main
  (i64.store (i32.add (local.get $sp) (i32.const 24)) (local.get $b))
  (i64.store (i32.add (local.get $sp) (i32.const 16)) (local.get $c))
  (i64.store (i32.add (local.get $sp) (i32.const 8))  (local.get $d))
  (i64.store (local.get $sp)                          (local.get $e))
)

;; Modulo 0x06
(func $mod_512
  ;; dividend
  (param $a i64)
  (param $b i64)
  (param $c i64)
  (param $d i64)
  (param $e i64)
  (param $f i64)
  (param $g i64)
  (param $h i64)

  ;; divisor
  (param $a1 i64)
  (param $b1 i64)
  (param $c1 i64)
  (param $d1 i64)
  (param $e1 i64)
  (param $f1 i64)
  (param $g1 i64)
  (param $h1 i64)

  (param $sp i32)

  ;; quotient
  (local $aq i64)
  (local $bq i64)
  (local $cq i64)
  (local $dq i64)

  ;; mask
  (local $maska i64)
  (local $maskb i64)
  (local $maskc i64)
  (local $maskd i64)
  (local $maske i64)
  (local $maskf i64)
  (local $maskg i64)
  (local $maskh i64)

  (local $carry i32)
  (local $temp i64)

  (local.set $maskh (i64.const 1))

  (block $main
    ;; check div by 0
    (if (call $iszero_512 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1) (local.get $f1) (local.get $g1) (local.get $h1))
      (then
        (local.set $e (i64.const 0))
        (local.set $f (i64.const 0))
        (local.set $g (i64.const 0))
        (local.set $h (i64.const 0))
        (br $main)
      )
    )

    ;; align bits
    (block $done
      (loop $loop
        ;; align bits;
        (if (i32.or (i64.eqz (i64.clz (local.get $a1)))
          (call $gte_512 (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1) (local.get $f1) (local.get $g1) (local.get $h1)
                         (local.get $a)  (local.get $b)  (local.get $c)  (local.get $d)  (local.get $e)  (local.get $f)  (local.get $g)  (local.get $h)))
          (br $done)
        )

        ;; divisor = divisor << 1
        (local.set $a1 (i64.add (i64.shl (local.get $a1) (i64.const 1)) (i64.shr_u (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shl (local.get $b1) (i64.const 1)) (i64.shr_u (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shl (local.get $c1) (i64.const 1)) (i64.shr_u (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.add (i64.shl (local.get $d1) (i64.const 1)) (i64.shr_u (local.get $e1) (i64.const 63))))
        (local.set $e1 (i64.add (i64.shl (local.get $e1) (i64.const 1)) (i64.shr_u (local.get $f1) (i64.const 63))))
        (local.set $f1 (i64.add (i64.shl (local.get $f1) (i64.const 1)) (i64.shr_u (local.get $g1) (i64.const 63))))
        (local.set $g1 (i64.add (i64.shl (local.get $g1) (i64.const 1)) (i64.shr_u (local.get $h1) (i64.const 63))))
        (local.set $h1 (i64.shl (local.get $h1) (i64.const 1)))

        ;; mask = mask << 1
        (local.set $maska (i64.add (i64.shl (local.get $maska) (i64.const 1)) (i64.shr_u (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shl (local.get $maskb) (i64.const 1)) (i64.shr_u (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shl (local.get $maskc) (i64.const 1)) (i64.shr_u (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.add (i64.shl (local.get $maskd) (i64.const 1)) (i64.shr_u (local.get $maske) (i64.const 63))))
        (local.set $maske (i64.add (i64.shl (local.get $maske) (i64.const 1)) (i64.shr_u (local.get $maskf) (i64.const 63))))
        (local.set $maskf (i64.add (i64.shl (local.get $maskf) (i64.const 1)) (i64.shr_u (local.get $maskg) (i64.const 63))))
        (local.set $maskg (i64.add (i64.shl (local.get $maskg) (i64.const 1)) (i64.shr_u (local.get $maskh) (i64.const 63))))
        (local.set $maskh (i64.shl (local.get $maskh) (i64.const 1)))
        (br $loop)
      )
    )

    (block $done
      (loop $loop
        ;; loop while mask != 0
        (if (call $iszero_512 (local.get $maska) (local.get $maskb) (local.get $maskc) (local.get $maskd) (local.get $maske) (local.get $maskf) (local.get $maskg) (local.get $maskh))
          (br $done)
        )
        ;; if dividend >= divisor
        (if (call $gte_512
          (local.get $a)  (local.get $b)  (local.get $c)  (local.get $d)  (local.get $e)  (local.get $f)  (local.get $g)  (local.get $h)
          (local.get $a1) (local.get $b1) (local.get $c1) (local.get $d1) (local.get $e1) (local.get $f1) (local.get $g1) (local.get $h1))
          (then
            ;; dividend = dividend - divisor
            (local.set $carry (i64.lt_u (local.get $h) (local.get $h1)))
            (local.set $h     (i64.sub  (local.get $h) (local.get $h1)))

            (local.set $temp  (i64.sub  (local.get $g) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $g)))
            (local.set $g     (i64.sub  (local.get $temp) (local.get $g1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $g) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $f) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $f)))
            (local.set $f     (i64.sub  (local.get $temp) (local.get $f1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $f) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $e) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $e)))
            (local.set $e     (i64.sub  (local.get $temp) (local.get $e1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $e) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $d) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $d)))
            (local.set $d     (i64.sub  (local.get $temp) (local.get $d1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $d) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $c) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $c)))
            (local.set $c     (i64.sub  (local.get $temp) (local.get $c1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $c) (local.get $temp)) (local.get $carry)))

            (local.set $temp  (i64.sub  (local.get $b) (i64.extend_i32_u (local.get $carry))))
            (local.set $carry (i64.gt_u (local.get $temp) (local.get $b)))
            (local.set $b     (i64.sub  (local.get $temp) (local.get $b1)))
            (local.set $carry (i32.or   (i64.gt_u (local.get $b) (local.get $temp)) (local.get $carry)))
            (local.set $a     (i64.sub  (i64.sub (local.get $a) (i64.extend_i32_u (local.get $carry))) (local.get $a1)))
          )
        )
        ;; divisor = divisor >> 1
        (local.set $h1 (i64.add (i64.shr_u (local.get $h1) (i64.const 1)) (i64.shl (local.get $g1) (i64.const 63))))
        (local.set $g1 (i64.add (i64.shr_u (local.get $g1) (i64.const 1)) (i64.shl (local.get $f1) (i64.const 63))))
        (local.set $f1 (i64.add (i64.shr_u (local.get $f1) (i64.const 1)) (i64.shl (local.get $e1) (i64.const 63))))
        (local.set $e1 (i64.add (i64.shr_u (local.get $e1) (i64.const 1)) (i64.shl (local.get $d1) (i64.const 63))))
        (local.set $d1 (i64.add (i64.shr_u (local.get $d1) (i64.const 1)) (i64.shl (local.get $c1) (i64.const 63))))
        (local.set $c1 (i64.add (i64.shr_u (local.get $c1) (i64.const 1)) (i64.shl (local.get $b1) (i64.const 63))))
        (local.set $b1 (i64.add (i64.shr_u (local.get $b1) (i64.const 1)) (i64.shl (local.get $a1) (i64.const 63))))
        (local.set $a1 (i64.shr_u (local.get $a1) (i64.const 1)))

        ;; mask = mask >> 1
        (local.set $maskh (i64.add (i64.shr_u (local.get $maskh) (i64.const 1)) (i64.shl (local.get $maskg) (i64.const 63))))
        (local.set $maskg (i64.add (i64.shr_u (local.get $maskg) (i64.const 1)) (i64.shl (local.get $maskf) (i64.const 63))))
        (local.set $maskf (i64.add (i64.shr_u (local.get $maskf) (i64.const 1)) (i64.shl (local.get $maske) (i64.const 63))))
        (local.set $maske (i64.add (i64.shr_u (local.get $maske) (i64.const 1)) (i64.shl (local.get $maskd) (i64.const 63))))
        (local.set $maskd (i64.add (i64.shr_u (local.get $maskd) (i64.const 1)) (i64.shl (local.get $maskc) (i64.const 63))))
        (local.set $maskc (i64.add (i64.shr_u (local.get $maskc) (i64.const 1)) (i64.shl (local.get $maskb) (i64.const 63))))
        (local.set $maskb (i64.add (i64.shr_u (local.get $maskb) (i64.const 1)) (i64.shl (local.get $maska) (i64.const 63))))
        (local.set $maska (i64.shr_u (local.get $maska) (i64.const 1)))
        (br $loop)
      )
    )
  );; end of main

  (i64.store (local.get $sp) (local.get $e))
  (i64.store (i32.sub (local.get $sp) (i32.const 8)) (local.get $f))
  (i64.store (i32.sub (local.get $sp) (i32.const 16)) (local.get $g))
  (i64.store (i32.sub (local.get $sp) (i32.const 24)) (local.get $h))
)

(func $mul_256
  ;;  a b c d e f g h
  ;;* i j k l m n o p
  ;;----------------
  (param $a i64)
  (param $c i64)
  (param $e i64)
  (param $g i64)

  (param $i i64)
  (param $k i64)
  (param $m i64)
  (param $o i64)

  (param $sp i32)

  (local $b i64)
  (local $d i64)
  (local $f i64)
  (local $h i64)
  (local $j i64)
  (local $l i64)
  (local $n i64)
  (local $p i64)
  (local $temp6 i64)
  (local $temp5 i64)
  (local $temp4 i64)
  (local $temp3 i64)
  (local $temp2 i64)
  (local $temp1 i64)
  (local $temp0 i64)

  ;; split the ops
  (local.set $b (i64.and (local.get $a) (i64.const 4294967295)))
  (local.set $a (i64.shr_u (local.get $a) (i64.const 32)))

  (local.set $d (i64.and (local.get $c) (i64.const 4294967295)))
  (local.set $c (i64.shr_u (local.get $c) (i64.const 32)))

  (local.set $f (i64.and (local.get $e) (i64.const 4294967295)))
  (local.set $e (i64.shr_u (local.get $e) (i64.const 32)))

  (local.set $h (i64.and (local.get $g) (i64.const 4294967295)))
  (local.set $g (i64.shr_u (local.get $g) (i64.const 32)))

  (local.set $j (i64.and (local.get $i) (i64.const 4294967295)))
  (local.set $i (i64.shr_u (local.get $i) (i64.const 32)))

  (local.set $l (i64.and (local.get $k) (i64.const 4294967295)))
  (local.set $k (i64.shr_u (local.get $k) (i64.const 32)))

  (local.set $n (i64.and (local.get $m) (i64.const 4294967295)))
  (local.set $m (i64.shr_u (local.get $m) (i64.const 32)))

  (local.set $p (i64.and (local.get $o) (i64.const 4294967295)))
  (local.set $o (i64.shr_u (local.get $o) (i64.const 32)))
  ;; first row multiplication
  ;; p * h
  (local.set $temp0 (i64.mul (local.get $p) (local.get $h)))
  ;; p * g + carry
  (local.set $temp1 (i64.add (i64.mul (local.get $p) (local.get $g)) (i64.shr_u (local.get $temp0) (i64.const 32))))
  ;; p * f + carry
  (local.set $temp2 (i64.add (i64.mul (local.get $p) (local.get $f)) (i64.shr_u (local.get $temp1) (i64.const 32))))
  ;; p * e + carry
  (local.set $temp3 (i64.add (i64.mul (local.get $p) (local.get $e)) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; p * d + carry
  (local.set $temp4 (i64.add (i64.mul (local.get $p) (local.get $d)) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; p * c + carry
  (local.set $temp5  (i64.add (i64.mul (local.get $p) (local.get $c)) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; p * b + carry
  (local.set $temp6  (i64.add (i64.mul (local.get $p) (local.get $b)) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; p * a + carry
  (local.set $a  (i64.add (i64.mul (local.get $p) (local.get $a)) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; second row
  ;; o * h + $temp1 "pg"
  (local.set $temp1 (i64.add (i64.mul (local.get $o) (local.get $h)) (i64.and (local.get $temp1) (i64.const 4294967295))))
  ;; o * g + $temp2 "pf" + carry
  (local.set $temp2 (i64.add (i64.add (i64.mul (local.get $o) (local.get $g)) (i64.and (local.get $temp2) (i64.const 4294967295))) (i64.shr_u (local.get $temp1) (i64.const 32))))
  ;; o * f + $temp3 "pe" + carry
  (local.set $temp3 (i64.add (i64.add (i64.mul (local.get $o) (local.get $f)) (i64.and (local.get $temp3) (i64.const 4294967295))) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; o * e + $temp4  + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $o) (local.get $e)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; o * d + $temp5  + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $o) (local.get $d)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; o * c + $temp6  + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $o) (local.get $c)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; o * b + $a  + carry
  (local.set $a (i64.add (i64.add (i64.mul (local.get $o) (local.get $b)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))
  ;; third row - n
  ;; n * h + $temp2
  (local.set $temp2 (i64.add (i64.mul (local.get $n) (local.get $h)) (i64.and (local.get $temp2) (i64.const 4294967295))))
  ;; n * g + $temp3 + carry
  (local.set $temp3 (i64.add (i64.add (i64.mul (local.get $n) (local.get $g)) (i64.and (local.get $temp3) (i64.const 4294967295))) (i64.shr_u (local.get $temp2) (i64.const 32))))
  ;; n * f + $temp4 + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $n) (local.get $f)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; n * e + $temp5  + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $n) (local.get $e)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; n * d + $temp6  + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $n) (local.get $d)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; n * c + $a  + carry
  (local.set $a (i64.add (i64.add (i64.mul (local.get $n) (local.get $c)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))

  ;; forth row
  ;; m * h + $temp3
  (local.set $temp3 (i64.add (i64.mul (local.get $m) (local.get $h)) (i64.and (local.get $temp3) (i64.const 4294967295))))
  ;; m * g + $temp4 + carry
  (local.set $temp4 (i64.add (i64.add (i64.mul (local.get $m) (local.get $g)) (i64.and (local.get $temp4) (i64.const 4294967295))) (i64.shr_u (local.get $temp3) (i64.const 32))))
  ;; m * f + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $m) (local.get $f)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; m * e + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $m) (local.get $e)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; m * d + $a + carry
  (local.set $a (i64.add (i64.add (i64.mul (local.get $m) (local.get $d)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))

  ;; fith row
  ;; l * h + $temp4
  (local.set $temp4 (i64.add (i64.mul (local.get $l) (local.get $h)) (i64.and (local.get $temp4) (i64.const 4294967295))))
  ;; l * g + $temp5 + carry
  (local.set $temp5 (i64.add (i64.add (i64.mul (local.get $l) (local.get $g)) (i64.and (local.get $temp5) (i64.const 4294967295))) (i64.shr_u (local.get $temp4) (i64.const 32))))
  ;; l * f + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $l) (local.get $f)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; l * e + $a + carry
  (local.set $a (i64.add (i64.add (i64.mul (local.get $l) (local.get $e)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))

  ;; sixth row
  ;; k * h + $temp5
  (local.set $temp5 (i64.add (i64.mul (local.get $k) (local.get $h)) (i64.and (local.get $temp5) (i64.const 4294967295))))
  ;; k * g + $temp6 + carry
  (local.set $temp6 (i64.add (i64.add (i64.mul (local.get $k) (local.get $g)) (i64.and (local.get $temp6) (i64.const 4294967295))) (i64.shr_u (local.get $temp5) (i64.const 32))))
  ;; k * f + $a + carry
  (local.set $a (i64.add (i64.add (i64.mul (local.get $k) (local.get $f)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))))

  ;; seventh row
  ;; j * h + $temp6
  (local.set $temp6 (i64.add (i64.mul (local.get $j) (local.get $h)) (i64.and (local.get $temp6) (i64.const 4294967295))))
  ;; j * g + $a + carry

  ;; eigth row
  ;; i * h + $a
  (local.set $a (i64.add (i64.mul (local.get $i) (local.get $h)) (i64.and (i64.add (i64.add (i64.mul (local.get $j) (local.get $g)) (i64.and (local.get $a) (i64.const 4294967295))) (i64.shr_u (local.get $temp6) (i64.const 32))) (i64.const 4294967295))))

  ;; combine terms
  (local.set $a (i64.or (i64.shl (local.get $a) (i64.const 32)) (i64.and (local.get $temp6) (i64.const 4294967295))))
  (local.set $c (i64.or (i64.shl (local.get $temp5) (i64.const 32)) (i64.and (local.get $temp4) (i64.const 4294967295))))
  (local.set $e (i64.or (i64.shl (local.get $temp3) (i64.const 32)) (i64.and (local.get $temp2) (i64.const 4294967295))))
  (local.set $g (i64.or (i64.shl (local.get $temp1) (i64.const 32)) (i64.and (local.get $temp0) (i64.const 4294967295))))

  ;; save stack
  (i64.store (local.get $sp) (local.get $a))
  (i64.store (i32.sub (local.get $sp) (i32.const 8)) (local.get $c))
  (i64.store (i32.sub (local.get $sp) (i32.const 16)) (local.get $e))
  (i64.store (i32.sub (local.get $sp) (i32.const 24)) (local.get $g))
)

)
