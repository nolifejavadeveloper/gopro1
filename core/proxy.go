package core

import (
	"github.com/rs/zerolog"
	"gopro/core/event"
	"gopro/core/proto/encryption"
	"io"
	"net"
	"os"
	"time"
)

type Proxy struct {
	debug  bool
	logger zerolog.Logger

	eventBus *event.Bus
	keypair  *encryption.Keypair
}

type HandlerDependency struct {
	EventBus *event.Bus
	Keypair  *encryption.Keypair
}

func NewProxy(debug bool) *Proxy {
	return &Proxy{debug: debug, logger: createLogger(debug), eventBus: event.NewEventBus()}
}

func start(options startupOptions) {
	proxy := NewProxy(options.debug)

	keypair, err := encryption.MakeKeypairBytes()
	if err != nil {
		proxy.logger.Panic().Err(err).Msg("Failed to make keypair")
	}

	proxy.keypair = keypair

	if options.debug {
		proxy.logger.Debug().Msg("Debug mode enabled")
	}

	proxy.loadPlugins()
	err = proxy.listen(":25565")
	if err != nil {
		proxy.logger.Panic().Err(err).Msg("Failed to start listener")
	}
}

func (p *Proxy) Shutdown() {
	p.shutdownPlugins()
}

func (p *Proxy) listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic("Error while starting listener: " + err.Error())
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			p.logger.Panic().Err(err).Msg("Failed to close listener")
		}
	}(listener)
	p.logger.Info().Msgf("Listening on %s", addr)

	for {

		conn, err := listener.Accept()
		if err != nil {
			p.logger.Info().Err(err).Msg("Error accepting connection")
			continue
		}

		go p.handleConnection(conn)
	}
}

func (p *Proxy) loadPlugins() {
	for _, plugin := range Plugins {
		err := plugin.Init(p)
		if err != nil {
			p.logger.Error().Err(err).Msg("Error loading plugin")
		}
	}
}

func (p *Proxy) shutdownPlugins() {
	for _, plugin := range Plugins {
		err := plugin.Shutdown(p)
		if err != nil {
			p.logger.Error().Err(err).Msg("Error shutting down plugin")
		}
	}
}

func createLogger(debug bool) zerolog.Logger {
	var level zerolog.Level
	if debug {
		level = zerolog.DebugLevel
	} else {
		level = zerolog.InfoLevel
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).Level(level).With().Timestamp().Logger()
	return logger
}

func (p *Proxy) handleConnection(conn net.Conn) {
	wrapped := Wrap(conn, p.logger.With().Str("address", conn.RemoteAddr().String()).Logger(), &HandlerDependency{EventBus: p.eventBus, Keypair: p.keypair})

	defer func() {
		wrapped.Close()
	}()

	wrapped.Logger.Debug().Msg("New connection")

	p.handlePackets(wrapped)
}

func (p *Proxy) handlePackets(conn *Conn) {
	for {
		packet, err := conn.Read()
		if err != nil {
			if err == io.EOF {
				conn.Logger.Debug().Msg("Connection closed")
				conn.Close()
			}
			return
		}

		if packet == nil {
			return
		}

		conn.currentHandler.Handle(packet)

		if err != nil {
			conn.Logger.Error().Err(err).Msg("Error handling packet, connection closed")
			conn.Close()
			return
		}
	}
}
