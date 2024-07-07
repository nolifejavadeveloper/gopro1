package packets

import (
	"gopro/core/component"
	"gopro/core/proto"
	"gopro/core/proto/encoding"
)

type LoginStart struct {
	Name encoding.String
}

type Disconnect struct {
	Reason encoding.String
}

type EncryptionRequest struct {
	PublicKey   encoding.ByteArray
	VerifyToken encoding.ByteArray
	ShouldAuth  encoding.Boolean
}

type EncryptionResponse struct {
	SharedSecret encoding.ByteArray
	VerifyToken  encoding.ByteArray
}

func NewLoginStart(pk *proto.Packet) (*LoginStart, error) {
	var ls LoginStart

	err := pk.Read(&ls.Name)
	return &ls, err
}

func NewDisconnect(component *component.TextComponent) (*Disconnect, error) {
	s, err := component.Serialize()
	if err != nil {
		return &Disconnect{}, err
	}

	return &Disconnect{Reason: encoding.String(s)}, nil
}

func NewEncryptionRequest(pub []byte, verifyToken []byte) *EncryptionRequest {
	return &EncryptionRequest{
		PublicKey:   pub,
		VerifyToken: verifyToken,
		ShouldAuth:  true,
	}
}

func NewEncryptionResponse(pk *proto.Packet) (*EncryptionResponse, error) {
	var er EncryptionResponse

	err := pk.Read(
		&er.SharedSecret,
		&er.VerifyToken,
	)

	return &er, err
}
