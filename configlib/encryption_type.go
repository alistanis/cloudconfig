package configlib

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
