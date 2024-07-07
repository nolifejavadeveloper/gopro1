package core

import (
	"github.com/rs/zerolog"
	"gopro/core/event"
	"gopro/core/proto"
	"gopro/core/proto/packets"
)

type statusHandler struct {
	deps   *HandlerDependency
	conn   *Conn
	logger zerolog.Logger
}

func newsStatusHandler(deps *HandlerDependency, conn *Conn) *statusHandler {
	return &statusHandler{deps: deps, conn: conn, logger: conn.Logger.With().Str("handler", "status").Logger()}
}

func (h *statusHandler) Handle(packet *proto.Packet) {
	switch packet.ID {
	case 0x00:
		{
			h.handleStatusRequest()
		}
	case 0x01:
		{
			h.handlePing(packet)
		}
	}
}

func (h *statusHandler) handleStatusRequest() {
	h.logger.Debug().Msg("Handling Status Request")
	e := event.NewServerStatusRequestEvent()
	h.deps.EventBus.Trigger(e)

	packet, err := packets.NewStatusResponse(e.Response)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "request").Msg("Error while initializing packet, closing connection")
		h.conn.Close()
		return
	}

	toSend, err := proto.Write(0x00, &packet.JSONResponse)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "response").Msg("Error while writing packet, closing connection")
		h.conn.Close()
		return
	}

	err = h.conn.SendPacket(toSend)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "response").Msg("Error while sending packet, closing connection")
		h.conn.Close()
		return
	}

}

func (h *statusHandler) handlePing(packet *proto.Packet) {
	h.logger.Debug().Msg("Handling Status Ping")
	ping, err := packets.NewStatusPing(packet)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "ping").Msg("Error while reading packet, closing connection")
		h.conn.Close()
		return
	}

	toSend, err := proto.Write(0x01, &ping.Payload)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "pong").Msg("Error while writing packet, closing connection")
		h.conn.Close()
		return
	}

	err = h.conn.SendPacket(toSend)
	if err != nil {
		h.logger.Error().Err(err).Str("packet", "pong").Msg("Error while sending packet, closing connection")
		h.conn.Close()
	}
}
