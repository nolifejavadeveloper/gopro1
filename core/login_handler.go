package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog"
	"gopro/core/component"
	"gopro/core/proto"
	"gopro/core/proto/encoding"
	"gopro/core/proto/encryption"
	"gopro/core/proto/packets"
	"io"
	"net/http"
	"os"
	"strings"
)

type loginHandler struct {
	deps   *HandlerDependency
	conn   *Conn
	logger zerolog.Logger

	username string

	token []byte
}

func newLoginHandler(deps *HandlerDependency, conn *Conn) *loginHandler {
	return &loginHandler{deps: deps, conn: conn, logger: conn.Logger.With().Str("handler", "login").Logger()}
}

func (h *loginHandler) Handle(packet *proto.Packet) {
	switch packet.ID {
	case 0x00:
		{
			h.handleLoginStart(packet)
		}
	case 0x01:
		{
			h.handleEncryptionResponse(packet)
		}
	}
}

func (h *loginHandler) disconnect(reason *component.TextComponent) {
	packet, err := packets.NewDisconnect(reason)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "disconnect").Msg("Error while initializing packet, closing connection")
		h.conn.Close()
		return
	}
	toSend, err := proto.Write(0x00, &packet.Reason)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "disconnect").Msg("Error while writing packet, closing connection")
		h.conn.Close()
		return
	}

	err = h.conn.SendPacket(toSend)

	if err != nil {
		h.logger.Error().Err(err).Str("packet", "disconnect").Msg("Error while sending packet, closing connection")
		h.conn.Close()
		return
	}
}

func (h *loginHandler) handleLoginStart(packet *proto.Packet) {
	h.logger.Debug().Msg("Handling Login Start")
	ls, err := packets.NewLoginStart(packet)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "start").Msg("Error while reading packet, closing connection")
		h.conn.Close()
		return
	}

	h.username = string(ls.Name)

	h.writeEncryptionRequest()
}

func (h *loginHandler) writeEncryptionRequest() {
	h.logger.Debug().Msg("Writing Encryption Request")
	token, err := h.generateVerifyToken()
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_request").Msg("Error while generating verify token, closing connection")
		h.conn.Close()
		return
	}

	h.token = token

	packet := packets.NewEncryptionRequest(h.deps.Keypair.Public, token)
	toSend, err := proto.Write(0x01, &packet.ServerId, &packet.PublicKey, &packet.VerifyToken)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_request").Msg("Error while writing packet, closing connection")
		h.conn.Close()
		return
	}

	//TODO: figure out whether to include the length of the public key inside the pubkey object instead of getting the len every time

	err = h.conn.SendPacket(toSend)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_request").Msg("Error while sending packet, closing connection")
		h.conn.Close()
		return
	}

	h.conn.encryptedState = encryption.PrivateKey
}

func (h *loginHandler) generateVerifyToken() ([]byte, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (h *loginHandler) handleEncryptionResponse(packet *proto.Packet) {
	es, err := packets.NewEncryptionResponse(packet)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_response").Msg("error while reading packet, closing connection")
		h.conn.Close()
	}

	h.decrypt(&es.SharedSecret)
	h.decrypt(&es.VerifyToken)

	if bytes.Equal(h.token, es.VerifyToken) {
		h.logger.Debug().Msg("verify token matched")
	} else {
		h.logger.Debug().Msg("verify token isn't matched")
		h.conn.Close()
		return
	}

	//change the encryption state
	h.conn.encryptedState = encryption.SharedKey
	h.conn.sharedSecret = es.SharedSecret

	//release the memory
	h.token = nil

	hash := h.makeHash()
	res, err := http.Get(url + "username=" + h.username + "&serverId=" + hash)
	fmt.Println(err)
	fmt.Println(res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}

func (h *loginHandler) decrypt(ba *encoding.ByteArray) []byte {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, h.deps.Keypair.Private, *ba)
	if err != nil {
		h.logger.Error().Err(err).Msg("error decrypting encryption response data")
		h.conn.Close()
		return nil
	}

	*ba = decrypted

	return decrypted
}

func (h *loginHandler) makeHash() string {
	sha := sha1.New()

	sha.Write(h.conn.sharedSecret)
	sha.Write(h.deps.Keypair.Public)
	hash := sha.Sum(nil)

	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}

//func (h *loginHandler) auth() bool {
//
//}
