// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package os

import (
	"dara"
	"internal/poll"
	"runtime"
	"syscall"
)

// fixLongPath is a noop on non-Windows platforms.
func fixLongPath(path string) string {
	return path
}

func rename(oldname, newname string) error {
	fi, err := Lstat(newname)
	if err == nil && fi.IsDir() {
		// There are two independent errors this function can return:
		// one for a bad oldname, and one for a bad newname.
		// At this point we've determined the newname is bad.
		// But just in case oldname is also bad, prioritize returning
		// the oldname error because that's what we did historically.
		if _, err := Lstat(oldname); err != nil {
			if pe, ok := err.(*PathError); ok {
				err = pe.Err
			}
			return &LinkError{"rename", oldname, newname, err}
		}
		return &LinkError{"rename", oldname, newname, syscall.EEXIST}
	}
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[RENAME] : ")
		    print(oldname)
		    print(" ")
		    println(newname)
        })
		argInfo1 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo1.String[:], oldname)
		argInfo2 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo2.String[:], newname)
		retInfo := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_RENAME, 2, 1, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_RENAME, syscallInfo)
	}
	err = syscall.Rename(oldname, newname)
	if err != nil {
		return &LinkError{"rename", oldname, newname, err}
	}
	return nil
}

// file is the real representation of *File.
// The extra level of indirection ensures that no clients of os
// can overwrite this data, which could cause the finalizer
// to close the wrong file descriptor.
type file struct {
	pfd         poll.FD
	name        string
	dirinfo     *dirInfo // nil unless directory being read
	nonblock    bool     // whether we set nonblocking mode
	stdoutOrErr bool     // whether this is stdout or stderr
}

// Fd returns the integer Unix file descriptor referencing the open file.
// The file descriptor is valid only until f.Close is called or f is garbage collected.
// On Unix systems this will cause the SetDeadline methods to stop working.
func (f *File) Fd() uintptr {
	if f == nil {
		return ^(uintptr(0))
	}

	// If we put the file descriptor into nonblocking mode,
	// then set it to blocking mode before we return it,
	// because historically we have always returned a descriptor
	// opened in blocking mode. The File will continue to work,
	// but any blocking operation will tie up a thread.
	if f.nonblock {
		f.pfd.SetBlocking()
	}

	return uintptr(f.pfd.Sysfd)
}

// NewFile returns a new File with the given file descriptor and
// name. The returned value will be nil if fd is not a valid file
// descriptor.
func NewFile(fd uintptr, name string) *File {
	return newFile(fd, name, kindNewFile)
}

// newFileKind describes the kind of file to newFile.
type newFileKind int

const (
	kindNewFile newFileKind = iota
	kindOpenFile
	kindPipe
)

// newFile is like NewFile, but if called from OpenFile or Pipe
// (as passed in the kind parameter) it tries to add the file to
// the runtime poller.
func newFile(fd uintptr, name string, kind newFileKind) *File {
	fdi := int(fd)
	if fdi < 0 {
		return nil
	}
	f := &File{&file{
		pfd: poll.FD{
			Sysfd:         fdi,
			IsStream:      true,
			ZeroReadIsEOF: true,
		},
		name:        name,
		stdoutOrErr: fdi == 1 || fdi == 2,
	}}

	// Don't try to use kqueue with regular files on FreeBSD.
	// It crashes the system unpredictably while running all.bash.
	// Issue 19093.
	if runtime.GOOS == "freebsd" && kind == kindOpenFile {
		kind = kindNewFile
	}

	pollable := kind == kindOpenFile || kind == kindPipe
	if err := f.pfd.Init("file", pollable); err != nil {
		// An error here indicates a failure to register
		// with the netpoll system. That can happen for
		// a file descriptor that is not supported by
		// epoll/kqueue; for example, disk files on
		// GNU/Linux systems. We assume that any real error
		// will show up in later I/O.
	} else if pollable {
		// We successfully registered with netpoll, so put
		// the file into nonblocking mode.
		if err := syscall.SetNonblock(fdi, true); err == nil {
			f.nonblock = true
		}
	}

	runtime.SetFinalizer(f.file, (*file).close)
	return f
}

// Auxiliary information if the File describes a directory
type dirInfo struct {
	buf  []byte // buffer for directory I/O
	nbuf int    // length of buf; return value from Getdirentries
	bufp int    // location of next record in buf.
}

// epipecheck raises SIGPIPE if we get an EPIPE error on standard
// output or standard error. See the SIGPIPE docs in os/signal, and
// issue 11845.
func epipecheck(file *File, e error) {
	if e == syscall.EPIPE && file.stdoutOrErr {
		sigpipe()
	}
}

// DevNull is the name of the operating system's ``null device.''
// On Unix-like systems, it is "/dev/null"; on Windows, "NUL".
const DevNull = "/dev/null"

// openFileNolog is the Unix implementation of OpenFile.
func openFileNolog(name string, flag int, perm FileMode) (*File, error) {
	chmod := false
	if !supportsCreateWithStickyBit && flag&O_CREATE != 0 && perm&ModeSticky != 0 {
		if _, err := Stat(name); IsNotExist(err) {
			chmod = true
		}
	}

	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[OPEN] : ")
		    print(name + " ")
		    print(flag)
		    print(" ")
		    println(perm)
        })
		argInfo1 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo1.String[:], name)
		argInfo2 := dara.GeneralType{Type: dara.INTEGER, Integer: flag}
		argInfo3 := dara.GeneralType{Type: dara.INTEGER, Integer: int(perm)}
		retInfo1 := dara.GeneralType{Type: dara.POINTER, Unsupported: dara.UNSUPPORTEDVAL}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_OPEN, 3, 2, [10]dara.GeneralType{argInfo1, argInfo2, argInfo3}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_OPEN, syscallInfo)
	}
	var r int
	for {
		var e error
		r, e = syscall.Open(name, flag|syscall.O_CLOEXEC, syscallMode(perm))
		if e == nil {
			break
		}

		// On OS X, sigaction(2) doesn't guarantee that SA_RESTART will cause
		// open(2) to be restarted for regular files. This is easy to reproduce on
		// fuse file systems (see http://golang.org/issue/11180).
		if runtime.GOOS == "darwin" && e == syscall.EINTR {
			continue
		}

		return nil, &PathError{"open", name, e}
	}

	// open(2) itself won't handle the sticky bit on *BSD and Solaris
	if chmod {
		Chmod(name, perm)
	}

	// There's a race here with fork/exec, which we are
	// content to live with. See ../syscall/exec_unix.go.
	if !supportsCloseOnExec {
		syscall.CloseOnExec(r)
	}

	return newFile(uintptr(r), name, kindOpenFile), nil
}

// Close closes the File, rendering it unusable for I/O.
// It returns an error, if any.
func (f *File) Close() error {
	if f == nil {
		return ErrInvalid
	}
	return f.file.close()
}

func (file *file) close() error {
	if file == nil {
		return syscall.EINVAL
	}
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[CLOSE] : ")
		    println(file.name)
        })
		argInfo := dara.GeneralType{Type: dara.FILE}
        copy(argInfo.String[:], file.name)
		retInfo := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_CLOSE, 1, 1, [10]dara.GeneralType{argInfo}, [10]dara.GeneralType{retInfo}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_CLOSE, syscallInfo)
	}
	var err error
	if e := file.pfd.Close(); e != nil {
		if e == poll.ErrFileClosing {
			e = ErrClosed
		}
		err = &PathError{"close", file.name, e}
	}

	// no need for a finalizer anymore
	runtime.SetFinalizer(file, nil)
	return err
}

// read reads up to len(b) bytes from the File.
// It returns the number of bytes read and an error, if any.
func (f *File) read(b []byte) (n int, err error) {
	n, err = f.pfd.Read(b)
	runtime.KeepAlive(f)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[READ] : ")
		    println(f.file.name)
        })
		argInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(argInfo1.String[:], f.name)
		argInfo2 := dara.GeneralType{Type: dara.ARRAY, Integer: len(b)}
		retInfo1 := dara.GeneralType{Type: dara.INTEGER, Integer: n}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_READ, 2, 2, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_READ, syscallInfo)
	}
	return n, err
}

// pread reads len(b) bytes from the File starting at byte offset off.
// It returns the number of bytes read and the error, if any.
// EOF is signaled by a zero count with err set to nil.
func (f *File) pread(b []byte, off int64) (n int, err error) {
	n, err = f.pfd.Pread(b, off)
	runtime.KeepAlive(f)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[PREAD] : ")
		    print(f.file.name)
		    print(" ")
		    println(off)
        })
		argInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(argInfo1.String[:], f.name)
		argInfo2 := dara.GeneralType{Type: dara.ARRAY, Integer: len(b)}
		argInfo3 := dara.GeneralType{Type: dara.INTEGER64, Integer64: off}
		retInfo1 := dara.GeneralType{Type: dara.INTEGER, Integer: n}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_PREAD64, 3, 2, [10]dara.GeneralType{argInfo1, argInfo2, argInfo3}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_PREAD64, syscallInfo)
	}
	return n, err
}

// write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
func (f *File) write(b []byte) (n int, err error) {
	n, err = f.pfd.Write(b)
	runtime.KeepAlive(f)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[WRITE] : ")
		    print(f.file.name)
		    print(" ")
		    println(string(b[:len(b)]))
        })
		argInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(argInfo1.String[:], f.name)
		argInfo2 := dara.GeneralType{Type: dara.ARRAY, Integer: len(b)}
		retInfo1 := dara.GeneralType{Type: dara.INTEGER, Integer: n}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_WRITE, 2, 2, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_WRITE, syscallInfo)
	}
	return n, err
}

// pwrite writes len(b) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any.
func (f *File) pwrite(b []byte, off int64) (n int, err error) {
    // DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[PWRITE] : ")
		    print(f.file.name)
		    print(" ")
		    print(string(b[:len(b)]))
		    print(" ")
		    println(off)
        })
		argInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(argInfo1.String[:], f.name)
		argInfo2 := dara.GeneralType{Type: dara.ARRAY, Integer: len(b)}
		argInfo3 := dara.GeneralType{Type: dara.INTEGER64, Integer64: off}
		retInfo1 := dara.GeneralType{Type: dara.INTEGER, Integer: n}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_PWRITE64, 3, 2, [10]dara.GeneralType{argInfo1, argInfo2, argInfo3}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_PWRITE64, syscallInfo)
	}
	n, err = f.pfd.Pwrite(b, off)
	runtime.KeepAlive(f)
	return n, err
}

// seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
func (f *File) seek(offset int64, whence int) (ret int64, err error) {
	ret, err = f.pfd.Seek(offset, whence)
	runtime.KeepAlive(f)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[SEEK] : ")
		    print(f.file.name)
		    print(" ")
		    print(offset)
		    print(" ")
		    println(whence)
        })
		argInfo1 := dara.GeneralType{Type: dara.FILE}
        copy(argInfo1.String[:], f.name)
		argInfo2 := dara.GeneralType{Type: dara.INTEGER64, Integer64: offset}
		argInfo3 := dara.GeneralType{Type: dara.INTEGER, Integer: whence}
		retInfo1 := dara.GeneralType{Type: dara.INTEGER64, Integer64: ret}
		retInfo2 := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_LSEEK, 3, 2, [10]dara.GeneralType{argInfo1, argInfo2, argInfo3}, [10]dara.GeneralType{retInfo1, retInfo2}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_LSEEK, syscallInfo)
	}
	return ret, err
}

// Truncate changes the size of the named file.
// If the file is a symbolic link, it changes the size of the link's target.
// If there is an error, it will be of type *PathError.
func Truncate(name string, size int64) error {
    // DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[TRUNCATE] : ")
		    print(name)
		    print(" ")
		    println(size)
        })
		argInfo1 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo1.String[:], name)
		argInfo2 := dara.GeneralType{Type: dara.INTEGER64, Integer64: size}
		retInfo := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_TRUNCATE, 2, 1, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_TRUNCATE, syscallInfo)
	}
	if e := syscall.Truncate(name, size); e != nil {
		return &PathError{"truncate", name, e}
	}
	return nil
}

// Remove removes the named file or directory.
// If there is an error, it will be of type *PathError.
func Remove(name string) error {
	// System call interface forces us to know
	// whether name is a file or directory.
	// Try both: it is cheaper on average than
	// doing a Stat plus the right one
	e := syscall.Unlink(name)
	if e == nil {
		return nil
	}
	e1 := syscall.Rmdir(name)
	if e1 == nil {
		return nil
	}

	// Both failed: figure out which error to return.
	// OS X and Linux differ on whether unlink(dir)
	// returns EISDIR, so can't use that. However,
	// both agree that rmdir(file) returns ENOTDIR,
	// so we can use that to decide which error is real.
	// Rmdir might also return ENOTDIR if given a bad
	// file path, like /etc/passwd/foo, but in that case,
	// both errors will be ENOTDIR, so it's okay to
	// use the error from unlink.
	if e1 != syscall.ENOTDIR {
		e = e1
	}
	return &PathError{"remove", name, e}
}

func tempDir() string {
	dir := Getenv("TMPDIR")
	if dir == "" {
		if runtime.GOOS == "android" {
			dir = "/data/local/tmp"
		} else {
			dir = "/tmp"
		}
	}
	return dir
}

// Link creates newname as a hard link to the oldname file.
// If there is an error, it will be of type *LinkError.
func Link(oldname, newname string) error {
	e := syscall.Link(oldname, newname)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[LINK] : ")
		    print(oldname)
		    print(" ")
		    println(newname)
        })
		argInfo1 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo1.String[:], oldname)
		argInfo2 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo2.String[:], newname)
		retInfo := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_LINK, 2, 1, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_LINK, syscallInfo)
	}
	if e != nil {
		return &LinkError{"link", oldname, newname, e}
	}
	return nil
}

// Symlink creates newname as a symbolic link to oldname.
// If there is an error, it will be of type *LinkError.
func Symlink(oldname, newname string) error {
	e := syscall.Symlink(oldname, newname)
	// DARA Instrumentation
	if runtime.Is_dara_profiling_on() {
        runtime.Dara_Debug_Print(func() {
		    print("[LINK] : ")
		    print(oldname)
		    print(" ")
		    println(newname)
        })
		argInfo1 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo1.String[:], oldname)
		argInfo2 := dara.GeneralType{Type: dara.STRING}
        copy(argInfo2.String[:], newname)
		retInfo := dara.GeneralType{Type: dara.ERROR, Unsupported: dara.UNSUPPORTEDVAL}
		syscallInfo := dara.GeneralSyscall{dara.DSYS_SYMLINK, 2, 1, [10]dara.GeneralType{argInfo1, argInfo2}, [10]dara.GeneralType{retInfo}}
		runtime.Report_Syscall_To_Scheduler(dara.DSYS_SYMLINK, syscallInfo)
	}
	if e != nil {
		return &LinkError{"symlink", oldname, newname, e}
	}
	return nil
}
