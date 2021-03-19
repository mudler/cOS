package cos_test

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	ssh "golang.org/x/crypto/ssh"
)

const (
	grubSwap    = "dev=$(blkid -L COS_STATE);	mount -o rw,remount $dev; mount $dev /boot/grub2; sed -i 's/set default=.*/set default=%s/' /boot/grub2/grub.cfg; mount -o ro,remount $dev"
	Passive     = 0
	Active      = iota
	Recovery    = iota
	UnknownBoot = iota
)

type SUT struct {
	Host     string
	Username string
	Password string
}

func NewSUT(Host, Username, Password string) *SUT {
	user := os.Getenv("COS_USER")
	if Username == "" {
		user = "root"
	}
	pass := os.Getenv("COS_PASS")
	if Password == "" {
		pass = "cos"
	}

	host := os.Getenv("COS_HOST")
	if Host == "" {
		host = "127.0.0.1:2222"
	}
	return &SUT{
		Host:     host,
		Username: user,
		Password: pass,
	}
}

func (s *SUT) ChangeBoot(b int) error {

	var bootEntry string

	switch b {
	case Active:
		bootEntry = "cos"
	case Passive:
		bootEntry = "fallback"
	case Recovery:
		bootEntry = "recovery"
	}

	_, err := s.command(fmt.Sprintf(grubSwap, bootEntry), false)
	if err != nil {
		return err
	}
	return nil
}

// Reset runs reboots cOS into Recovery and runs cos-reset.
// It will boot back the system from the Active partition afterwards
func (s *SUT) Reset() {
	err := s.ChangeBoot(Recovery)
	Expect(err).ToNot(HaveOccurred())

	s.Reboot()

	Expect(s.BootFrom()).To(Equal(Recovery))
	out, err := s.command("cos-reset", false)
	Expect(err).ToNot(HaveOccurred())
	Expect(out).Should(ContainSubstring("Installing"))

	err = s.ChangeBoot(Active)
	Expect(err).ToNot(HaveOccurred())

	s.Reboot()

	ExpectWithOffset(1, s.BootFrom()).To(Equal(Active))
}

// BootFrom returns the booting partition of the SUT
func (s *SUT) BootFrom() int {
	out, err := s.command("cat /proc/cmdline", false)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	switch {
	case strings.Contains(out, "COS_ACTIVE"):
		return Active
	case strings.Contains(out, "COS_PASSIVE"):
		return Passive
	case strings.Contains(out, "COS_SYSTEM"):
		return Recovery
	default:
		return UnknownBoot
	}
}

func (s *SUT) EventuallyConnects(t ...int) {
	dur := 120
	if len(t) > 0 {
		dur = t[0]
	}
	Eventually(func() error {
		out, err := s.command("echo ping", true)
		if out == "ping\n" {
			return nil
		}
		return err
	}, time.Duration(time.Duration(dur)*time.Second), time.Duration(5*time.Second)).ShouldNot(HaveOccurred())
}

// Command sends a command to the SUIT and waits for reply
func (s *SUT) Command(cmd string) (string, error) {
	return s.command(cmd, false)
}

func (s *SUT) command(cmd string, timeout bool) (string, error) {
	client, err := s.connectToHost(timeout)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), errors.Wrap(err, string(out))
	}

	return string(out), err
}

// Reboot reboots the system under test
func (s *SUT) Reboot() {
	s.command("reboot", true)
	time.Sleep(10 * time.Second)
	s.EventuallyConnects()
}

func (s *SUT) connectToHost(timeout bool) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:    s.Username,
		Auth:    []ssh.AuthMethod{ssh.Password(s.Password)},
		Timeout: 30 * time.Second, // max time to establish connection
	}

	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := DialWithDeadline("tcp", s.Host, sshConfig, timeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// DialWithDeadline Dials SSH with a deadline to avoid Read timeouts
func DialWithDeadline(network string, addr string, config *ssh.ClientConfig, timeout bool) (*ssh.Client, error) {
	conn, err := net.DialTimeout(network, addr, config.Timeout)
	if err != nil {
		return nil, err
	}
	if config.Timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(config.Timeout))
		conn.SetWriteDeadline(time.Now().Add(config.Timeout))
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	if !timeout {
		conn.SetReadDeadline(time.Time{})
		conn.SetWriteDeadline(time.Time{})
	}

	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for range t.C {
			_, _, err := c.SendRequest("keepalive@golang.org", true, nil)
			if err != nil {
				return
			}
		}
	}()
	return ssh.NewClient(c, chans, reqs), nil
}
