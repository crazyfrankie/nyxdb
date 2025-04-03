package nyx

type DB struct {
	mm  *memTable   // our latest in-memory table (active-written)
	imm []*memTable // add here only AFTER pushing to flushChan.

	nextMemfd int // Initialized through openMemTables.

	opt *option
}
