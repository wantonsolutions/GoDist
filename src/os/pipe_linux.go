// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import "dara"
import "runtime"
import "syscall"

// Pipe returns a connected pair of Files; reads from r return bytes written to w.
// It returns the files and an error, if any.
func Pipe() (r *File, w *File, err error) {
	var p [2]int
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
		runtime.Dara_Debug_Print(func() { println("[PIPE]") })
		retInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(retInfo1.String[:], r.name)
		retInfo2 := dara.GeneralType{Type: dara.FILE}
        copy(retInfo2.String[:], w.name)
		retInfo3 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_PIPE2, 0, 3, [10]dara.GeneralType{}, [10]dara.GeneralType{retInfo1, retInfo2, retInfo3}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_PIPE2, syscallInfo)
	}
	e := syscall.Pipe2(p[0:], syscall.O_CLOEXEC)
	// pipe2 was added in 2.6.27 and our minimum requirement is 2.6.23, so it
	// might not be implemented.
	if e == syscall.ENOSYS {
		// See ../syscall/exec.go for description of lock.
		syscall.ForkLock.RLock()
		e = syscall.Pipe(p[0:])
		if e != nil {
			syscall.ForkLock.RUnlock()
			return nil, nil, NewSyscallError("pipe", e)
		}
		syscall.CloseOnExec(p[0])
		syscall.CloseOnExec(p[1])
		syscall.ForkLock.RUnlock()
	} else if e != nil {
		return nil, nil, NewSyscallError("pipe2", e)
	}

	return newFile(uintptr(p[0]), "|0", kindPipe), newFile(uintptr(p[1]), "|1", kindPipe), nil
}


