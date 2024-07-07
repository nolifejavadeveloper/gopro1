package packets

import (
	"encoding/json"
	"gopro/core/proto"
	"gopro/core/proto/encoding"
	"gopro/core/proto/status"
)

type StatusResponse struct {
	JSONResponse encoding.String
}

type StatusPing struct {
	Payload encoding.Long
}

func NewStatusResponse(response *status.Response) (*StatusResponse, error) {
	val, err := json.Marshal(response)

	if err != nil {
		return nil, err
	}

	return &StatusResponse{JSONResponse: encoding.String(val)}, nil
}

func NewStatusPing(pk *proto.Packet) (*StatusPing, error) {
	var sp StatusPing

	if err := pk.Read(
		&sp.Payload,
	); err != nil {
		return &sp, err
	}

	return &sp, nil
}
