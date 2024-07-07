package core

import (
	"github.com/rs/zerolog"
	"gopro/core/proto"
	"gopro/core/proto/packets"
)

type handshakeHandler struct {
	deps   *HandlerDependency
	conn   *Conn
	logger zerolog.Logger
}

func newHandshakeHandler(deps *HandlerDependency, conn *Conn) *handshakeHandler {
	return &handshakeHandler{deps: deps, conn: conn, logger: conn.Logger.With().Str("handler", "handshake").Logger()}
}

func (h *handshakeHandler) Handle(packet *proto.Packet) {
	handshakePacket, err := packets.NewHandshake(packet)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read handshake packet")
		return
	}

	nextState := byte(handshakePacket.NextState)

	h.conn.ProtocolVersion = byte(handshakePacket.Protocol)

	var handler PacketHandler
	switch nextState {
	case 1:
		handler = newsStatusHandler(h.deps, h.conn)
	case 2:
		handler = newLoginHandler(h.deps, h.conn)
	default:
		{
			h.logger.Error().Int("intent", int(nextState)).Msg("invalid handshake intent")
			h.conn.Close()
		}
	}

	h.conn.SwitchState(nextState)
	h.conn.SwitchPacketHandler(handler)
}
