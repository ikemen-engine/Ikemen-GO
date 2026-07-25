// IKEMEN:VOID memory write pass-through for Raw Execution cheapie exploits.
//
// Entry points:
//   - voidMemPassThroughWrite / voidMemPassThroughRead
//   - voidMemHealthWritePassThrough (health/life OOB writes, no validation)
//   - voidMemTryUnsafeWrite / voidMemTryUnsafeRead (recover-wrapped unsafe access)
//
// Implementation lives in the mem.go section of supernull_var.go (same package).
package main
