package configlib

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// LoadMetaConfig loads a meta configuration from the MetaConfigPath, which by default is stored at $HOME/.cloudconfig.meta
func LoadMetaConfig() (*MetaConfig, error) {
	data, err := ioutil.ReadFile(MetaConfigPath)
	if err != nil {
		return nil, err
	}
	c := &MetaConfig{}
	return c, json.Unmarshal(data, c)
}

// MetaConfig refers to the cloudconfig application's internal configuration
type MetaConfig struct {
	*bufio.Scanner `json:"-"`
	io.Writer      `json:"-"`
	exitMsg        string
	StorageType    string          // StorageType is the storage type, either local or remote (s3, etc)
	ConfigPath     string          // ConfigPath is the path at which the config rests
	EncryptionType *EncryptionType // EncryptionType will not be nil if the config is encrypted
}

// GenerateMetaConfig generates a meta config by polling for information on r and w, then calling c.InputFields
func GenerateMetaConfig(r io.Reader, w io.Writer) (*MetaConfig, error) {
	c := NewMetaConfig(bufio.NewScanner(r), w)
	c.InputFields()
	return c, nil
}

// NewMetaConfig returns a new meta config with its Scanner and Writer initialized
func NewMetaConfig(s *bufio.Scanner, w io.Writer) *MetaConfig {
	m := &MetaConfig{}
	m.Scanner = s
	m.Writer = w
	return m
}

// ExitMsg returns the MetaConfig's exit message
func (c *MetaConfig) ExitMsg() string {
	return c.exitMsg
}

// GenerateEncryptionType gets input on c.Scanner and writes output on c.Writer in order to get information for encryption type
func (c *MetaConfig) GenerateEncryptionType() {
	s := c.GetInput("Would you like to encrypt your configuration file? (y, n, yes, no: case insensitive)", nil, "y", "n", "yes", "no")
	switch s {
	case "y", "yes":
		c.EncryptionType = &EncryptionType{}
		// TODO - implement EncPrivateKey
		c.EncryptionType.Type = c.GetInput("What kind of encryption would you like to use? (password, private key)", nil, Password, NoEncPrivateKey)

	}
}

// InputFields gets input for all of this MetaConfigs fields
func (c *MetaConfig) InputFields() {
	c.InputStorageType()
	c.InputConfigPath()
}

// GetInput takes a prompt, a comparison function/nil, and a potential list of string comparisons
// it validates input based on the compFunc or the strings input must match, and if successful returns the text
// will block until valid text is given
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

// InputStorageType prompts for the storage type of your cloud config
func (c *MetaConfig) InputStorageType() {
	s := c.GetInput("Please enter a storage type for your config (local, s3)", nil, Local, S3)
	c.StorageType = s
}

// InputConfigPath prompts for information regarding cloud config path
func (c *MetaConfig) InputConfigPath() {
	s := c.GetInput("Please enter a storage path for your config. (default: $HOME/.cloudconfig)", c.ValidPath)
	if s != "" {
		if strings.HasPrefix(s, "~/") {
			s = ReplaceHomeDir(s)
		}
		c.ConfigPath = s
		return
	}
	c.ConfigPath = DefaultConfigPath
}

// Save saves this file to disk
func (c *MetaConfig) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(MetaConfigPath, data, 0644)
}

// WriteString writes a string to the MetaConfig's internal writer
func (c *MetaConfig) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

// ValidPath checks to see if path is a valid path for local or s3 storage
func (c *MetaConfig) ValidPath(path string) bool {
	switch c.StorageType {
	case Local:
		return c.ValidFilePath(path)
	case S3:
		return strings.HasPrefix(path, "s3://")
	}
	return false
}

// ValidFilePath checks to see if this path is a valid local file path
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
