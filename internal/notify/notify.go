package notify

import (
	"time"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog/log"
)

const appName = "offlinemsmtp"

type Urgency byte

const (
	UrgencyLow      Urgency = 0
	UrgencyNormal   Urgency = 1
	UrgencyCritical Urgency = 2
)

// Handle is a reference to a displayed notification that can be closed.
type Handle struct {
	id uint32
	n  *Notifier
}

func (h *Handle) Close() {
	if h == nil || h.n == nil || h.n.inner == nil {
		return
	}
	_, _ = h.n.inner.CloseNotification(h.id)
}

type Notifier struct {
	conn   *dbus.Conn
	inner  notify.Notifier
	silent bool
}

func New(silent bool) *Notifier {
	if silent {
		return &Notifier{silent: true}
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warn().Err(err).Msg("cannot connect to D-Bus session bus; notifications disabled")
		return &Notifier{silent: true}
	}
	inner, err := notify.New(conn)
	if err != nil {
		log.Warn().Err(err).Msg("cannot create notifier; notifications disabled")
		return &Notifier{silent: true}
	}
	return &Notifier{conn: conn, inner: inner}
}

func (n *Notifier) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

func (n *Notifier) Send(message string, timeout time.Duration, urgency Urgency) *Handle {
	if n.silent || n.inner == nil {
		return nil
	}
	notif := notify.Notification{
		AppName:       appName,
		Summary:       appName,
		Body:          message,
		ExpireTimeout: timeout,
		Hints: map[string]dbus.Variant{
			"urgency": dbus.MakeVariant(byte(urgency)),
		},
	}
	id, err := n.inner.SendNotification(notif)
	if err != nil {
		log.Warn().Err(err).Msg("cannot send desktop notification")
		return nil
	}
	return &Handle{id: id, n: n}
}
