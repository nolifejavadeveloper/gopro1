package core

import (
	"bytes"
	"crypto/rand"
	"github.com/rs/zerolog"
	"gopro/core/component"
	"gopro/core/proto"
	"gopro/core/proto/packets"
)

type loginHandler struct {
	deps   *HandlerDependency
	conn   *Conn
	logger zerolog.Logger

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
	_, err := packets.NewLoginStart(packet)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "start").Msg("Error while reading packet, closing connection")
		h.conn.Close()
		return
	}

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

	packet := packets.NewEncryptionRequest(h.deps.Public, token)
	toSend, err := proto.Write(0x01, &packet.PublicKey, &packet.VerifyToken, &packet.ShouldAuth)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_request").Msg("Error while writing packet, closing connection")
		h.conn.Close()
		return
	}

	err = h.conn.SendPacket(toSend)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "encryption_request").Msg("Error while sending packet, closing connection")
		h.conn.Close()
		return
	}
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
		h.logger.Error().Err(err).Str("packet", "encryption_response").Msg("Error while reading packet, closing connection")
		h.conn.Close()
	}

	if bytes.Equal(h.token, es.VerifyToken) {
		h.logger.Debug().Msg("Verify token matched")
	} else {
		h.logger.Debug().Msg("Verify token isn't matched")
		h.conn.Close()
		return
	}

	//release the memory
	h.token = nil

}
