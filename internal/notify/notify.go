package notify

import (
	"time"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog"
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

func (n *Notifier) Send(message string, timeout time.Duration, urgency Urgency) *Handle {
	return n.Replace(nil, message, timeout, urgency)
}

// Replace updates an existing notification in place. If handle is nil a new
// notification is created. The old handle is invalidated.
func (n *Notifier) Replace(handle *Handle, message string, timeout time.Duration, urgency Urgency) *Handle {
	if n.silent || n.inner == nil {
		return nil
	}
	var replacesID uint32
	if handle != nil {
		replacesID = handle.id
	}
	notif := notify.Notification{
		AppName:       appName,
		ReplacesID:    replacesID,
		Summary:       appName,
		Body:          message,
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
