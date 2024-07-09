package core

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/rs/zerolog"
	"gopro/core/proto"
	"gopro/core/proto/encoding"
	"gopro/core/proto/encryption"
	"net"
)

type Conn struct {
	isActive bool
	Conn     net.Conn
	Logger   zerolog.Logger

	State           byte
	ProtocolVersion byte
	Threshold       int

	encryptedState encryption.EncryptionState
	sharedSecret   []byte
	encrypter      cipher.Stream
	decrypter      cipher.Stream

	currentHandler PacketHandler
}

type PacketHandler interface {
	Handle(packet *proto.Packet)
}

func Wrap(conn net.Conn, logger zerolog.Logger, deps *HandlerDependency) *Conn {
	wrapped := &Conn{Conn: conn, State: 0, Logger: logger}
	wrapped.currentHandler = newHandshakeHandler(deps, wrapped)
	return wrapped

}

func (c *Conn) StartEncrypting(sharedSecret []byte) {
	c.sharedSecret = sharedSecret
	c.encryptedState = encryption.SharedKey
	block, err := aes.NewCipher(c.sharedSecret)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Error creating AES cipher")
		c.Close()
		return
	}

	c.decrypter = cipher.NewCFBDecrypter(block, c.sharedSecret)
	c.encrypter = cipher.NewCFBEncrypter(block, c.sharedSecret)
}

func (c *Conn) Read() (*proto.Packet, error) {
	if c.isActive {
		return nil, nil
	}

	buffer := encoding.NewBuffer(make([]byte, 4096))

	//read packet data
	l, err := c.Conn.Read(buffer.Data)
	if err != nil {
		return nil, err
	}

	buffer.Data = buffer.Data[:l]

	if c.encryptedState == encryption.SharedKey {
		c.decrypt(buffer.Data)
	}

	var leng encoding.Varint
	err = leng.Read(buffer)
	if err != nil {
		c.Logger.Error().Err(err).Msg("error reading packet length")
		c.Close()
		return nil, err
	}

	if int(leng) > len(buffer.Data) {
		c.Logger.Error().Msg(fmt.Sprintf("incomplete packet received, expected %d but got %d", leng, len(buffer.Data)))
		c.Close()
		return nil, err
	}

	packet, err := proto.Parse(buffer)
	if err != nil {
		c.Logger.Error().Err(err).Msg("error parsing packet")
		c.Close()
		return nil, err
	}

	return packet, nil
}

func (c *Conn) SwitchState(b byte) {
	c.State = b
}

func (c *Conn) SwitchPacketHandler(handler PacketHandler) {
	c.currentHandler = handler
}
func (c *Conn) decrypt(bytearr []byte) {
	decrypted := make([]byte, len(bytearr))
	c.decrypter.XORKeyStream(decrypted, bytearr)
	copy(bytearr, decrypted)
}

func (c *Conn) SendPacket(pk *proto.Packet) error {
	_, err := c.Conn.Write(pk.Bytes())
	return err
}

func (c *Conn) Close() {
	if !c.isActive {
		return
	}

	err := c.Conn.Close()
	if err != nil {
		c.Logger.Error().Err(err).Msg("Error closing connection")
	}
}
