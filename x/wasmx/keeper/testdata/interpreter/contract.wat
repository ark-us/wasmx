
(module
  (import "ewasm" "CALLDATASIZE" (func $CALLDATASIZE ))
  (import "ewasm" "PUSH" (func $PUSH  (param i64 i64 i64 i64)))
  (import "ewasm" "MSTORE" (func $MSTORE ))
  (import "ewasm" "RETURN" (func $RETURN ))
  (import "ewasm" "GLOBAL_GET_SP" (func $GLOBAL_GET_SP (result i32)))
  (import "ewasm" "GLOBAL_SET_SP" (func $GLOBAL_SET_SP (param i32)))

(type $et12 (func))
(func $ewasm_ewasm_1 (export "ewasm_ewasm_1") (type $et12) (nop))

  (global $cb_dest (mut i32) (i32.const 0))
  (global $init (mut i32) (i32.const 0))

  (func $instantiate (export "instantiate"))
  (func $codesize (export "codesize") (result i32) (i32.const 9))
  (func $main
    (export "main")
    (local $jump_dest i32) (local $jump_map_switch i32)
    (local.set $jump_dest (i32.const -1))

    (block $done
      (loop $loop

  (block $0
    (if
      (i32.eqz (global.get $init))
      (then
        (global.set $init (i32.const 1))
        (br $0))
      (else
        ;; the callback dest can never be in the first block
        (if (i32.eq (global.get $cb_dest) (i32.const 0))
          (then
            (unreachable)
          )
          (else
            ;; return callback destination and zero out $cb_dest
            (local.set $jump_map_switch (global.get $cb_dest))
            (global.set $cb_dest (i32.const 0))
            (br_table $0  (local.get $jump_map_switch))
          ))))) (if (i32.gt_s (call $GLOBAL_GET_SP) (i32.const 32672))
                 (then (unreachable)))(call $CALLDATASIZE)
(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const 32)))
(call $PUSH (i64.const 0)(i64.const 0)(i64.const 0)(i64.const 0))(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const 32)))
(call $MSTORE)
(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const -64)))
(call $PUSH (i64.const 0)(i64.const 0)(i64.const 0)(i64.const 32))(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const 32)))
(call $PUSH (i64.const 0)(i64.const 0)(i64.const 0)(i64.const 0))(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const 32)))
(call $RETURN) (br $done)
(call $GLOBAL_SET_SP (i32.add (call $GLOBAL_GET_SP) (i32.const -64)))
)))

(func $evm_bytecode (export "evm_bytecode") (result i32 i32 i32)
    i32.const 33832
    i32.const 0
    i32.const 0)


)
