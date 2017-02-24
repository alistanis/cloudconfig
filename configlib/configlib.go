package configlib

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"strings"
)

var (
	CurrentUser       *user.User
	MetaConfigPath    string
	DefaultConfigPath string
	DefaultConfigType = "local"
)

// error codes
const (
	Unknown                        = -1
	ErrCodeCouldNotFindCurrentUser = iota + 1
)

// Storage types
const (
	Local = "local"
	S3    = "s3"
)

func init() {
	var err error
	// Get the current user and exit if we can't
	CurrentUser, err = user.Current()
	if err != nil {
		ExitError(err, ErrCodeCouldNotFindCurrentUser)
	}
	// Set default paths
	MetaConfigPath = filepath.Join(CurrentUser.HomeDir, ".cloudconfig.meta")
	DefaultConfigPath = filepath.Join(CurrentUser.HomeDir, ".cloudconfig")
}

// ReplaceHomeDir replaces any matching prefixes
func ReplaceHomeDir(s string) string {
	for _, r := range homeShortcuts {
		s = strings.Replace(s, r, CurrentUser.HomeDir+string(os.PathSeparator), -1)
	}
	return s
}

// StringComp compares to see if s is in comps
func StringComp(s string, comps ...string) bool {
	for _, cs := range comps {
		if s == cs {
			return true
		}
	}
	return false
}

// ExitError prints an error to stderr and exits with the given code
func ExitError(err error, code int) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}
