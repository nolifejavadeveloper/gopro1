package proto

import "gopro/core/proto/encoding"

type Packet struct {
	ID     byte
	buffer *encoding.Buffer
}

func Parse(buffer *encoding.Buffer) (*Packet, error) {
	var id encoding.Varint

	if err := id.Read(buffer); err != nil {
		return nil, err
	}

	return &Packet{ID: byte(id), buffer: buffer}, nil
}

func Write(id byte, types ...encoding.DataType) (*Packet, error) {
	pk := Packet{ID: id, buffer: encoding.NewBuffer([]byte{})}

	pk.ID = id
	err := pk.Write(types...)
	if err != nil {
		return nil, err
	}

	return &pk, nil
}

func (packet *Packet) Read(types ...encoding.DataType) error {
	return packet.buffer.Read(types...)
}

func (packet *Packet) Write(types ...encoding.DataType) error {
	return packet.buffer.WriteWithLength(packet.ID, types...)
}

func (packet *Packet) Bytes() []byte {
	return packet.buffer.Data
}

func (packet *Packet) Skip(types ...encoding.DataType) error {
	err := packet.buffer.Skip(types...)
	if err != nil {
		return err
	}

	return nil
}
