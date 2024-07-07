package event

import "gopro/core/component"

type LoginStartEvent struct {
	Username       string
	Declined       bool
	DeclinedReason *component.TextComponent
}

func NewLoginStartEvent(user string) *LoginStartEvent {
	return &LoginStartEvent{Username: user, Declined: false}
}

func (e *LoginStartEvent) Name() string {
	return "LoginStartEvent"
}

func (e *LoginStartEvent) Reject(declinedReason *component.TextComponent) {
	e.Declined = true
	e.DeclinedReason = declinedReason
}
