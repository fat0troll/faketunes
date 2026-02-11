package filesystem

import "sync/atomic"

func (f *FS) nextInode() uint64 {
	return atomic.AddUint64(&f.inodeCounter, 1)
}
