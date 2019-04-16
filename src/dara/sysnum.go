package dara

// Dara Syscall Numbers. These are the numbers used for identification of syscalls by the scheduler.
const (
	DSYS_READ = iota
	DSYS_WRITE
	DSYS_OPEN
	DSYS_CLOSE
	DSYS_STAT
	DSYS_FSTAT
	DSYS_LSTAT
	DSYS_LSEEK
	DSYS_PREAD64
	DSYS_PWRITE64
	DSYS_GETPAGESIZE
	DSYS_EXECUTABLE
	DSYS_GETPID
	DSYS_GETPPID
	DSYS_GETWD
	DSYS_READDIR
	DSYS_READDIRNAMES
	DSYS_WAIT4
	DSYS_KILL
	DSYS_GETUID
	DSYS_GETEUID
	DSYS_GETGID
	DSYS_GETEGID
	DSYS_GETGROUPS
	DSYS_EXIT
	DSYS_RENAME
	DSYS_TRUNCATE
	DSYS_UNLINK
	DSYS_RMDIR
	DSYS_LINK
	DSYS_SYMLINK
	DSYS_PIPE2
	DSYS_MKDIR
	DSYS_CHDIR
	DSYS_UNSETENV
	DSYS_GETENV
	DSYS_SETENV
	DSYS_CLEARENV
	DSYS_ENVIRON
	DSYS_TIMENOW
	DSYS_READLINK
	DSYS_CHMOD
	DSYS_FCHMOD
	DSYS_CHOWN
	DSYS_LCHOWN
	DSYS_FCHOWN
	DSYS_FTRUNCATE
	DSYS_FSYNC
	DSYS_UTIMES
	DSYS_FCHDIR
	DSYS_SETDEADLINE
	DSYS_SETREADDEADLINE
	DSYS_SETWRITEDEADLINE
	DSYS_NET_READ
	DSYS_NET_WRITE
	DSYS_NET_CLOSE
	DSYS_NET_SETDEADLINE
	DSYS_NET_SETREADDEADLINE
	DSYS_NET_SETWRITEDEADLINE
	DSYS_NET_SETREADBUFFER
	DSYS_NET_SETWRITEBUFFER
	DSYS_SOCKET
	DSYS_LISTEN_TCP
    DSYS_SLEEP
)

