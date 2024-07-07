package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
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
	privateKey     *rsa.PrivateKey
	sharedSecret   []byte
	encrypter      cipher.Stream
	decrypter      cipher.Stream

	currentHandler PacketHandler
}

type PacketHandler interface {
	Handle(packet *proto.Packet)
}

func Wrap(conn net.Conn, logger zerolog.Logger, deps *HandlerDependency, key *rsa.PrivateKey) *Conn {
	wrapped := &Conn{Conn: conn, State: 0, Logger: logger, privateKey: key}
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

	// Read packet data
	packetLength, err := c.readLength()
	if err != nil {
		return nil, err
	}

	packetData := make([]byte, packetLength)
	err = c.decrypt(packetData)
	if err != nil {
		return nil, err
	}

	packet, err := proto.Parse(packetData)

	if err != nil {
		c.Logger.Error().Err(err).Msg("Error reading packet data")
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

func (c *Conn) readLength() (int, error) {
	buffer := encoding.NewBuffer(make([]byte, 5))
	_, err := c.Conn.Read(buffer.Data)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Error reading length from connection")
		return 0, err
	}

	err = c.decrypt(buffer.Data)
	if err != nil {
		return 0, err
	}

	var varint encoding.Varint
	err = varint.Read(buffer)

	return int(varint), err
}

func (c *Conn) decrypt(bytearr []byte) error {
	switch c.encryptedState {
	case encryption.PrivateKey:
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, bytearr)
		if err != nil {
			c.Logger.Error().Err(err).Msg("Error decrypting data with private key")
			return err
		}
		copy(bytearr, decrypted)
		return nil

	case encryption.SharedKey:
		decrypted := make([]byte, len(bytearr))
		c.decrypter.XORKeyStream(decrypted, bytearr)
		copy(bytearr, decrypted)
		return nil
	}

	return nil
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
