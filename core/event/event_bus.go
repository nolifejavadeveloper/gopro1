package event

import (
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"sort"
	"sync"
)

type handlerWrapper struct {
	handler  reflect.Value
	priority int
}

type Bus struct {
	logger    zerolog.Logger
	listeners map[string][]handlerWrapper
	mu        sync.RWMutex
}

func NewEventBus() *Bus {
	return &Bus{
		listeners: make(map[string][]handlerWrapper),
		//FIX THUS
		logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

func (eb *Bus) Subscribe(priority int, handler interface{}) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	if handlerType.Kind() != reflect.Func || handlerType.NumIn() != 1 {
		eb.logger.Panic().Msgf("handler must be a function with exactly one argument")
	}

	//first argument
	firstArgument := handlerType.In(0)
	eventName := reflect.Zero(firstArgument).Interface().(Event).Name()

	eb.listeners[eventName] = append(eb.listeners[eventName], handlerWrapper{
		handler:  handlerValue,
		priority: priority,
	})

	//log the event registration
	eb.logger.Info().Str("event", eventName).Int("priority", priority).Msg("Event hooked")

	//sort handlers by priority
	sort.Slice(eb.listeners[eventName], func(i, j int) bool {
		return eb.listeners[eventName][i].priority > eb.listeners[eventName][j].priority
	})
}

func (eb *Bus) Trigger(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, ok := eb.listeners[event.Name()]; ok {
		go func(handlers []handlerWrapper) {
			for _, wrapper := range handlers {
				eventValue := reflect.ValueOf(event)
				if wrapper.handler.Type().In(0) == eventValue.Type() {
					wrapper.handler.Call([]reflect.Value{eventValue})
				}
			}
		}(handlers)
	}
}
