package encryption

const (
	NoEncryption = EncryptionState(iota)
	PrivateKey
	SharedKey
)

type EncryptionState byte
