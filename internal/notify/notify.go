package notify

import (
	"time"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog"
)

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
	log    zerolog.Logger
}

func New(silent bool, log zerolog.Logger) *Notifier {
	if silent {
		return &Notifier{silent: true, log: log}
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warn().Err(err).Msg("cannot connect to D-Bus session bus; notifications disabled")
		return &Notifier{silent: true, log: log}
	}
	inner, err := notify.New(conn)
	if err != nil {
		log.Warn().Err(err).Msg("cannot create notifier; notifications disabled")
		return &Notifier{silent: true, log: log}
	}
	return &Notifier{conn: conn, inner: inner, log: log}
}

func (n *Notifier) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

func (n *Notifier) Send(summary, body string, timeout time.Duration, urgency Urgency) *Handle {
	if n.silent || n.inner == nil {
		return nil
	}
	notif := notify.Notification{
		Summary:       summary,
		Body:          body,
		ExpireTimeout: timeout,
		Hints: map[string]dbus.Variant{
			"urgency": dbus.MakeVariant(byte(urgency)),
		},
	}
	id, err := n.inner.SendNotification(notif)
	if err != nil {
		n.log.Warn().Err(err).Msg("cannot send desktop notification")
		return nil
	}
	return &Handle{id: id, n: n}
}

// Replace closes handle and sends a fresh notification. Using ReplacesID
// does not reliably update urgency on most notification servers, so we
// close and recreate instead.
func (n *Notifier) Replace(handle *Handle, summary, body string, timeout time.Duration, urgency Urgency) *Handle {
	handle.Close()
	return n.Send(summary, body, timeout, urgency)
}
