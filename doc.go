// Package flowmatic contains easy-to-use generic helpers for structured concurrency.
//
// Comparison of simple helpers:
//
//	       Tasks       Cancels Context?   Collect results?
//	Do     Different   No                 No
//	All    Different   On error           No
//	Race   Different   On success         No
//	Each   Same        No                 No
//	Map    Same        On error           Yes
//
// ManageTasks and TaskPool allow for advanced concurrency patterns.
package flowmatic

// MaxProcs means use GOMAXPROCS workers when doing tasks.
const MaxProcs = -1
