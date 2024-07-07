package encoding

import (
	"io"
)

type Buffer struct {
	index int
	Data  []byte
}

func NewBuffer(data []byte) *Buffer {
	return &Buffer{Data: data, index: 0}
}

func (b *Buffer) Read(types1 ...DataType) error {
	for _, typ := range types1 {
		if err := typ.Read(b); err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) write(id byte, types1 ...DataType) error {
	Varint(int(id)).Write(b)
	for _, typ := range types1 {
		//write the data
		typ.Write(b)
	}

	return nil
}

func (b *Buffer) ReadByte() (byte, error) {
	if b.index > (len(b.Data) - 1) {
		return 0, io.EOF
	}

	bb := b.Data[b.index]
	b.index++

	return bb, nil
}

func (b *Buffer) ReadBytes(amount int) ([]byte, error) {
	if b.index > (len(b.Data) - 1) {
		return nil, io.EOF
	}

	bb := b.Data[b.index : b.index+amount]
	b.index += amount

	return bb, nil
}

func (b *Buffer) WriteByte(byt ...byte) {
	b.Data = append(b.Data, byt...)
}

func (b *Buffer) WriteWithLength(id byte, types1 ...DataType) error {
	err := b.write(id, types1...)
	if err != nil {
		return err
	}
	length := len(b.Data)
	lengthVarint := Varint(length)
	lengthLeg := lengthVarint.Len()

	result := make([]byte, 0, lengthLeg+length)

	lengthVarint.WriteIntoSlice(&result)

	result = append(result, b.Data...)

	b.Data = result

	return nil
}

func (b *Buffer) Skip(types ...DataType) error {
	for _, typ := range types {
		if err := typ.Skip(b); err != nil {
			return err
		}
	}

	return nil
}
