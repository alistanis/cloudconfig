package configlib

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"strings"

	"github.com/thisisfineio/go-cfg-gen"
)

var (
	CurrentUser       *user.User
	MetaConfigPath    string
	DefaultConfigPath string
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
	CurrentUser, err = user.Current()
	if err != nil {
		ExitError(err, ErrCodeCouldNotFindCurrentUser)
	}
	MetaConfigPath = filepath.Join(CurrentUser.HomeDir, ".cloudconfig.meta")
	DefaultConfigPath = filepath.Join(CurrentUser.HomeDir, ".cloudconfig")
}

func GenerateMetaConfig(r io.Reader, w io.Writer) (*MetaConfig, error) {
	c := NewMetaConfig(bufio.NewScanner(r), w)
	cfgen.SetReader(r)
	cfgen.SetWriter(w)
	c.GetFields()
	return c, nil
}

func NewMetaConfig(s *bufio.Scanner, w io.Writer) *MetaConfig {
	m := &MetaConfig{}
	m.Scanner = s
	m.Writer = w
	return m
}

func (c *MetaConfig) GetStorageType() {
	s := c.GetInput("Please enter a storage type for your config (local, s3)", nil, Local, S3)
	c.StorageType = s
}

func (c *MetaConfig) GetConfigPath() {
	s := c.GetInput("Please enter a storage path for your config. (default: $HOME/.cloudconfig)", c.IsValidPath)
	if s != "" {
		if strings.HasPrefix(s, "~/") {
			s = ReplaceHomeDir(s)
		}
		c.ConfigPath = s
		return
	}
	c.ConfigPath = DefaultConfigPath
}

func (c *MetaConfig) IsValidPath(path string) bool {
	switch c.StorageType {
	case Local:
		return c.ValidFilePath(path)
	case S3:
		return strings.HasPrefix(path, "s3://")
	}
	return false
}

func (c *MetaConfig) ValidFilePath(path string) bool {
	// for us, we'll consider empty string to be the default path
	if path == "" {
		return true
	}
	if strings.HasPrefix(path, "~/") {
		return c.ValidFilePath(ReplaceHomeDir(path))
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return true
	}
	// Write file and remove it
	if err := ioutil.WriteFile(path, []byte{}, 0644); err == nil {
		os.Remove(path)
		return true
	}

	return false
}

func ReplaceHomeDir(s string) string {
	return strings.Replace(s, "~/", CurrentUser.HomeDir+"/", -1)
}

func StringComp(s string, comps ...string) bool {
	for _, cs := range comps {
		if s == cs {
			return true
		}
	}
	return false
}

func (c *MetaConfig) GetInput(prompt string, compFunc func(s string) bool, comps ...string) string {
	c.WriteString(prompt + "\n")
	c.Scan()

	if compFunc != nil {
		if !compFunc(c.Text()) {
			return c.GetInput(prompt, compFunc)
		}
		return c.Text()
	}

	if !StringComp(c.Text(), comps...) {
		return c.GetInput(prompt, nil, comps...)
	}
	return c.Text()
}

func (c *MetaConfig) GetFields() {
	c.GetStorageType()
	c.GetConfigPath()
}

func LoadMetaConfig() (*MetaConfig, error) {
	data, err := ioutil.ReadFile(MetaConfigPath)
	if err != nil {
		return nil, err
	}
	c := &MetaConfig{}
	return c, json.Unmarshal(data, c)
}

type MetaConfig struct {
	*bufio.Scanner `json:"-"`
	io.Writer      `json:"-"`
	exitMsg        string
	StorageType    string          // StorageType is the storage type, either local or remote (s3, etc)
	ConfigPath     string          // ConfigPath is the path at which the config rests
	EncryptionType *EncryptionType // EncryptionType will not be nil if the config is encrypted
}

func (c *MetaConfig) GetEncryptionType() {
	s := c.GetInput("Would you like to encrypt your configuration file? (y, n, yes, no: case insensitive)", nil, "y", "n", "yes", "no")
	switch s {
	case "y", "yes":
		c.EncryptionType = &EncryptionType{}
		// TODO - implement EncPrivateKey
		c.EncryptionType.Type = c.GetInput("What kind of encryption would you like to use? (password, private key)", nil, Password, NoEncPrivateKey)

	}
}

func (c *MetaConfig) ExitMsg() string {
	return c.exitMsg
}

func (c *MetaConfig) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

func (c *MetaConfig) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(MetaConfigPath, data, 0644)
}

// Encryption types
const (
	Password        = "password"
	NoEncPrivateKey = "private key"
	EncPrivateKey   = "encrypted private key (password private key)"
)

type EncryptionType struct {
	Type        string // the type of encryption being used
	KeyLocation string // path to the key
	// Default for private key is ~/.ssh/cloudconfigkey
}

func (e *EncryptionType) GenerateKey(pass string) ([]byte, error) {
	return nil, nil
}

func ExitError(err error, code int) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}
