//go:build !unix && !windows

// @index Errno classification fallback for platforms without syscall.Errno constants, such as plan9 and wasip1.
package trace

// @intent keep ConvertSystemError compiling on platforms that do not define the errno constants.
// @ensures always reports false so the caller falls back to its portable rules.
func convertErrno(_ error) (error, bool) {
	return nil, false
}
