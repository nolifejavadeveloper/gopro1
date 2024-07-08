package encoding

import (
	"errors"
	"unicode/utf16"
)

type (
	DataType interface {
		Read(buffer *Buffer) error
		Write(buffer *Buffer)
		Skip(buffer *Buffer) error
	}
	Byte      byte
	Varint    int32
	UShort    uint16
	Long      int64
	String    string
	ByteArray []byte
	Boolean   bool
)

func (b *Byte) Read(buffer *Buffer) error {
	newByte, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	*b = Byte(newByte)
	return nil
}

func (b Byte) Write(buffer *Buffer) {
	buffer.WriteByte(byte(b))
}

func (b Byte) Skip(buffer *Buffer) error {
	buffer.index++

	return nil
}

func (v *Varint) Read(buffer *Buffer) error {
	var result int
	var shift uint

	for {
		current, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		result |= int(current&0x7F) << shift

		if current&0x80 == 0 {
			break
		}
		shift += 7
	}

	*v = Varint(result)

	return nil
}

func (v Varint) Write(buffer *Buffer) {
	number := int32(v)
	for {
		if (number & ^0x7F) == 0 {
			buffer.WriteByte(byte(number))
			return
		}

		buffer.WriteByte(byte(number&0x7F | 0x80))

		number >>= 7
	}
}

func (v Varint) Skip(buffer *Buffer) error {
	for {
		current, err := buffer.ReadByte()
		buffer.index++
		if err != nil {
			return err
		}

		if current&0x80 == 0 {
			break
		}
	}
	return nil
}

func (v Varint) WriteIntoSlice(slice *[]byte) {
	number := int32(v)
	for {
		if (number & ^0x7F) == 0 {
			*slice = append(*slice, byte(number))
			return
		}

		*slice = append(*slice, byte(number&0x7F|0x80))

		number >>= 7
	}
}

func (v Varint) Len() int {
	switch {
	case v < 0:
		return 0
	case v < 1<<(7*1):
		return 1
	case v < 1<<(7*2):
		return 2
	case v < 1<<(7*3):
		return 3
	case v < 1<<(7*4):
		return 4
	default:
		return 5
	}
}

func (v *UShort) Read(buffer *Buffer) error {
	byte1, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	byte2, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	*v = UShort(uint16(byte1)<<8 | uint16(byte2))

	return nil
}

func (v UShort) Write(buffer *Buffer) {
	val := uint16(v)
	buffer.WriteByte(byte(val>>8), byte(val))
}

func (v UShort) Skip(buffer *Buffer) error {
	buffer.index += 2
	return nil
}

func (v *Long) Read(buffer *Buffer) error {
	var val int64
	for i := 0; i < 8; i++ {
		b, err := buffer.ReadByte()
		if err != nil {
			return err
		}
		val |= int64(b) << (56 - 8*i)
	}

	*v = Long(val)

	return nil
}

func (v Long) Write(buffer *Buffer) {
	val := int64(v)

	buffer.WriteByte(byte(val>>56), byte(val>>48), byte(val>>40), byte(val>>32),
		byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
}

func (v Long) Skip(buffer *Buffer) error {
	buffer.index += 8
	return nil
}

func (v *String) Read(buffer *Buffer) error {
	var length Varint
	err := length.Read(buffer)
	if err != nil {
		return err
	}

	if length < 0 {
		return errors.New("invalid string length")
	}

	bytes, err := buffer.ReadBytes(int(length))

	if err != nil {
		return err
	}

	str := string(bytes)

	*v = String(str)

	return nil
}

func (v String) Write(buffer *Buffer) {
	val := string(v)
	utf16Length := len(utf16.Encode([]rune(val)))

	// Write the length as a Varint
	Varint(utf16Length).Write(buffer)

	buffer.WriteByte([]byte(val)...)
}

func (v String) Skip(buffer *Buffer) error {
	var length Varint
	err := length.Read(buffer)
	if err != nil {
		return err
	}

	buffer.index += int(length)

	return nil
}

func (b *ByteArray) Read(buffer *Buffer) error {
	var length Varint
	err := length.Read(buffer)
	if err != nil {
		return err
	}

	bytes := make([]byte, length)
	copy(bytes, buffer.Data[buffer.index:buffer.index+int(length)])
	buffer.index += int(length)

	*b = bytes

	return nil
}

func (b ByteArray) Write(buffer *Buffer) {
	leng := Varint(len(b))
	leng.Write(buffer)
	buffer.WriteByte(b...)
}

func (b ByteArray) Skip(buffer *Buffer) error {
	panic("implement me")
}

func (b *Boolean) Read(buffer *Buffer) error {
	bol, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	if bol == 0 {
		*b = false
	} else {
		*b = true
	}

	return nil
}

func (b Boolean) Write(buffer *Buffer) {
	if b {
		buffer.WriteByte(1)
	} else {
		buffer.WriteByte(0)
	}
}

func (b Boolean) Skip(buffer *Buffer) error {
	buffer.index++
	return nil
}
