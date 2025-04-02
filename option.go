package nyx

type option struct {
	Dir            string // Database home directory (holds SSTable)
	ValueDir       string // Directory for large values
	MemTableSize   int64  // MemTable size threshold (Flush if exceeded)
	SyncWrites     bool   // Whether each write is immediately flushed to disk
	ValueThreshold int64  // Threshold value, above which the value is written to ValueDir instead of Dir.
}
type Option func(*option)

var defaultMemTableOpt = &option{
	MemTableSize:   64 << 20, // 64 MB
	SyncWrites:     false,
	ValueThreshold: 1 << 20, // 1 MB
}

// WithSyncWrites returns a new Options value with SyncWrites set to the given value.
//
// Nyx does all writes via mmap. So, all writes can survive process crashes or k8s environments
// with SyncWrites set to false.
//
// When set to true, Badger would call an additional msync after writes to flush mmap buffer over to
// disk to survive hard reboots. Most users of Nyx should not need to do this.
//
// The default value of SyncWrites is false.
func WithSyncWrites(val bool) Option {
	return func(opt *option) {
		opt.SyncWrites = val
	}
}

// WithDir returns a new Options value with Dir set to the given value.
//
// Dir is the path of the directory where key data will be stored in.
// If it doesn't exist, Nyx will try to create it for you.
// This is set automatically to be the path given to `DefaultOptions`.
func WithDir(path string) Option {
	return func(opt *option) {
		opt.Dir = path
	}
}

// WithValueDir returns a new Options value with ValueDir set to the given value.
//
// ValueDir is the path of the directory where value data will be stored in.
// If it doesn't exist, Nyx will try to create it for you.
// This is set automatically to be the path given to `DefaultOptions`.
func WithValueDir(path string) Option {
	return func(opt *option) {
		opt.ValueDir = path
	}
}

// WithValueThreshold returns a new Options value with ValueThreshold set to the given value.
//
// ValueThreshold sets the threshold used to decide whether a value is stored directly in the LSM
// tree or separately in the log value files.
//
// The default value of ValueThreshold is 1 MB
func WithValueThreshold(val int64) Option {
	return func(opt *option) {
		opt.ValueThreshold = val
	}
}
