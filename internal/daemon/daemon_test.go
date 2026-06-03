package daemon

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseQueueFile(t *testing.T) {
	t.Run("with args and body", func(t *testing.T) {
		data := []byte("-t --read-envelope-from\nFrom: a@b.com\nSubject: Hello\n\nBody\n")
		args, msg, err := parseQueueFile(data)
		require.NoError(t, err)
		assert.Equal(t, "-t --read-envelope-from", args)
		assert.Equal(t, []byte("From: a@b.com\nSubject: Hello\n\nBody\n"), msg)
	})

	t.Run("empty args", func(t *testing.T) {
		data := []byte("\nFrom: a@b.com\n")
		args, msg, err := parseQueueFile(data)
		require.NoError(t, err)
		assert.Equal(t, "", args)
		assert.Equal(t, []byte("From: a@b.com\n"), msg)
	})

	t.Run("missing newline", func(t *testing.T) {
		_, _, err := parseQueueFile([]byte("no newline"))
		assert.Error(t, err)
	})
}

func TestExtractSubject(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		msg := []byte("From: a@b.com\r\nSubject: Hello World\r\nTo: c@d.com\r\n\r\nBody\r\n")
		assert.Equal(t, "Hello World", extractSubject(msg))
	})

	t.Run("absent", func(t *testing.T) {
		msg := []byte("From: a@b.com\r\nTo: c@d.com\r\n\r\nBody\r\n")
		assert.Equal(t, "<no subject>", extractSubject(msg))
	})

	t.Run("empty message", func(t *testing.T) {
		assert.Equal(t, "<no subject>", extractSubject([]byte{}))
	})
}

func TestBuildCmd(t *testing.T) {
	t.Run("default msmtp path", func(t *testing.T) {
		d := &Daemon{Config: Config{ConfigFile: "/home/user/.msmtprc", MsmtpPath: "msmtp"}}
		base := []string{"msmtp", "--debug", "-C", "/home/user/.msmtprc"}
		assert.Equal(t, base, d.buildCmd(""))
		assert.Equal(t, append(base, "-t", "--read-envelope-from"), d.buildCmd("-t --read-envelope-from"))
	})

	t.Run("custom msmtp path", func(t *testing.T) {
		d := &Daemon{Config: Config{ConfigFile: "/home/user/.msmtprc", MsmtpPath: "/usr/bin/msmtp"}}
		base := []string{"/usr/bin/msmtp", "--debug", "-C", "/home/user/.msmtprc"}
		assert.Equal(t, base, d.buildCmd(""))
	})
}

func TestSendEnabled(t *testing.T) {
	t.Run("no send-mail-file configured", func(t *testing.T) {
		d := &Daemon{Config: Config{}}
		assert.True(t, d.sendEnabled())
	})

	t.Run("file exists", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "send-mail-file-*")
		require.NoError(t, err)
		f.Close()
		d := &Daemon{Config: Config{SendMailFile: f.Name()}}
		assert.True(t, d.sendEnabled())
	})

	t.Run("file absent", func(t *testing.T) {
		d := &Daemon{Config: Config{SendMailFile: "/nonexistent/offlinemsmtp-test-file"}}
		assert.False(t, d.sendEnabled())
	})
}
