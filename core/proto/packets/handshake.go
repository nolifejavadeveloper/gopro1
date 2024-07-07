package packets

import (
	"gopro/core/proto"
	"gopro/core/proto/encoding"
)

type Handshake struct {
	Protocol  encoding.Varint
	NextState encoding.Varint
}

func NewHandshake(p *proto.Packet) (*Handshake, error) {
	var h Handshake

	var skippedAddress encoding.String
	var skippedPort encoding.UShort

	if err := p.Read(
		&h.Protocol,
	); err != nil {
		return nil, err
	}

	if err := p.Skip(
		&skippedAddress,
		&skippedPort,
	); err != nil {
		return nil, err
	}

	if err := p.Read(
		&h.NextState,
	); err != nil {
		return nil, err
	}

	return &h, nil
}
