package testutil

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"sync"
	"testing"
)

// SMTPServer is a minimal SMTP server listening on 127.0.0.1. It supports AUTH PLAIN
// (checking the given credentials) and records all the messages it receives.
type SMTPServer struct {
	Host string
	Port int

	user string
	pass string

	mu            sync.Mutex
	messages      []Message
	failRecipient string
}

// NewSMTPServer starts a server requiring the given credentials, stopped when the test ends.
func NewSMTPServer(t *testing.T, user, pass string) *SMTPServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to start the SMTP server: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	addr := listener.Addr().(*net.TCPAddr)
	s := &SMTPServer{Host: "127.0.0.1", Port: addr.Port, user: user, pass: pass}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.handle(conn)
		}
	}()

	return s
}

// Messages returns a copy of all the messages received so far.
func (s *SMTPServer) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message(nil), s.messages...)
}

// FailFor makes the server reject emails to the given address.
func (s *SMTPServer) FailFor(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failRecipient = address
}

func (s *SMTPServer) handle(conn net.Conn) {
	c := textproto.NewConn(conn)
	defer c.Close()

	authenticated := false
	reply := func(format string, args ...any) { _ = c.PrintfLine(format, args...) }

	reply("220 localhost santa test server")
	for {
		line, err := c.ReadLine()
		if err != nil {
			return
		}
		command, arg, _ := strings.Cut(line, " ")

		switch strings.ToUpper(command) {
		case "EHLO":
			reply("250-localhost")
			reply("250 AUTH PLAIN")
		case "HELO", "NOOP", "RSET":
			reply("250 OK")
		case "AUTH":
			mechanism, response, _ := strings.Cut(arg, " ")
			decoded, err := base64.StdEncoding.DecodeString(response)
			if strings.ToUpper(mechanism) != "PLAIN" || err != nil {
				reply("504 5.5.4 Unrecognized authentication type")
				continue
			}
			// PLAIN is `authzid\x00user\x00pass`
			parts := strings.Split(string(decoded), "\x00")
			if len(parts) == 3 && parts[1] == s.user && parts[2] == s.pass {
				authenticated = true
				reply("235 2.7.0 Authentication successful")
			} else {
				reply("535 5.7.8 Username and Password not accepted")
			}
		case "MAIL":
			if !authenticated {
				reply("530 5.7.0 Authentication required")
				continue
			}
			reply("250 OK")
		case "RCPT":
			s.mu.Lock()
			fail := s.failRecipient != "" && strings.Contains(arg, "<"+s.failRecipient+">")
			s.mu.Unlock()
			if fail {
				reply("550 5.1.1 Mailbox unavailable")
			} else {
				reply("250 OK")
			}
		case "DATA":
			reply("354 Go ahead")
			data, err := c.ReadDotBytes()
			if err != nil {
				return
			}
			message, err := parseMessage(data)
			if err != nil {
				reply("554 5.6.0 Unable to parse the message: %v", err)
				continue
			}
			s.mu.Lock()
			s.messages = append(s.messages, message)
			s.mu.Unlock()
			reply("250 OK")
		case "QUIT":
			reply("221 Bye")
			return
		default:
			reply("502 5.5.2 Command not implemented")
		}
	}
}

func parseMessage(data []byte) (Message, error) {
	m, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return Message{}, err
	}

	decoder := new(mime.WordDecoder)
	header := func(name string) string {
		value, err := decoder.DecodeHeader(m.Header.Get(name))
		if err != nil {
			return m.Header.Get(name)
		}
		return value
	}

	var body io.Reader = m.Body
	if strings.EqualFold(m.Header.Get("Content-Transfer-Encoding"), "quoted-printable") {
		body = quotedprintable.NewReader(body)
	}
	html, err := io.ReadAll(body)
	if err != nil {
		return Message{}, err
	}

	return Message{
		From:    header("From"),
		To:      header("To"),
		ReplyTo: header("Reply-To"),
		Subject: header("Subject"),
		// the SMTP data always ends with a line break, which isn't part of the body
		HTML: strings.TrimRight(string(html), "\r\n"),
	}, nil
}
