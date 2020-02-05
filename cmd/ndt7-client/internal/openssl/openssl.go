// +build ndt7openssl

// Package openssl contains an OpenSSL 1.1+ dialer.
package openssl

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	// #include <poll.h>
	//
	// #include <errno.h>
	// #include <limits.h>
	// #include <stdint.h>
	// #include <stdlib.h>
	// #include <string.h>
	//
	// #include <openssl/err.h>
	// #include <openssl/ssl.h>
	//
	// #cgo LDFLAGS: -lssl -lcrypto
	//
	// static int pollwrapper(int fd, short events, int timeout) {
	//   struct pollfd pfd;
	//   pfd.fd = fd;
	//   pfd.events = events;
	//   pfd.revents = 0;
	//   int rv = poll(&pfd, 1, timeout);
	//   if (rv < 0) {
	//     return -errno;
	//   }
	//   return rv;
	// }
	"C"
)
import (
	"fmt"
	"io/ioutil"
)

// Dialer dials connections using OpenSSL.
type Dialer struct {
	CABundlePath       string
	Dialer             *net.Dialer
	InsecureSkipVerify bool
}

// ErrCannotGuessCABundlePath is returned when we cannot guess the path of the
// CA bundle in the current system.
var ErrCannotGuessCABundlePath = errors.New("openssl: cannot guess CA bundle path")

func guessCABundlePath(readFile func(filename string) ([]byte, error)) (string, error) {
	// This list of CA bundle path locations was copied from
	// Measurement Kit m4 directory.
	candidates := []string{
		"/etc/pki/tls/certs/ca-bundle.crt",
		"/etc/ssl/cert.pem",
		"/etc/ssl/certs/ca-certificates.crt",
		"/usr/local/etc/openssl/cert.pem",
		"/usr/local/share/certs/ca-root.crt",
		"/usr/share/ssl/certs/ca-bundle.crt",
	}
	for _, name := range candidates {
		if _, err := readFile(name); err == nil {
			return name, nil
		}
	}
	return "", ErrCannotGuessCABundlePath
}

// NewDialer creates a new dialer. It may fail if we cannot guess the
// proper CA bundle path for this system.
func NewDialer() *Dialer {
	return &Dialer{Dialer: new(net.Dialer)}
}

// Dial dials an OpenSSL network connection.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext dials an OpenSSL network connection using a context.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("openssl: only tcp is supported")
	}
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	tcpconn := conn.(*net.TCPConn) // don't see why this could fail
	return d.newconn(tcpconn)
}

func (d *Dialer) newconn(conn *net.TCPConn) (net.Conn, error) {
	localAddr := conn.LocalAddr()
	remoteAddr := conn.RemoteAddr()
	fp, err := conn.File()
	conn.Close() // fp contains a cloned socket so we can close conn
	if err != nil {
		return nil, err
	}
	ssl, err := d.newssl(fp.Fd())
	if err != nil {
		fp.Close()
		return nil, err
	}
	return &opensslconn{
		filep:      fp,
		ssl:        ssl,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}, nil
}

func (d *Dialer) newssl(fd uintptr) (*C.struct_ssl_st, error) {
	method := C.TLS_client_method()
	if method == nil {
		return nil, errors.New("TLS_method failed")
	}
	ctx := C.SSL_CTX_new(method)
	if ctx == nil {
		return nil, errors.New("SSL_CTX_new failed")
	}
	defer C.SSL_CTX_free(ctx) // note that it's reference counted
	if d.InsecureSkipVerify == true {
		path := d.CABundlePath
		if path == "" {
			var err error
			path, err = guessCABundlePath(ioutil.ReadFile)
			if err != nil {
				return nil, err
			}
		}
		cpath := C.CString(path)
		defer C.free(unsafe.Pointer(cpath))
		retval := C.SSL_CTX_load_verify_locations(ctx, cpath, nil)
		if retval != 1 {
			return nil, errors.New("SSL_CTX_load_verify_locations failed")
		}
		// TODO(bassosimone): we also need to properly enable
		// the verification of remote certificates
	}
	ssl := C.SSL_new(ctx)
	if ssl == nil {
		return nil, errors.New("SSL_CTX_new failed")
	}
	if C.SSL_set_fd(ssl, C.int(fd)) != 1 {
		C.SSL_free(ssl)
		return nil, errors.New("SSL_set_fd failed")
	}
	C.SSL_set_connect_state(ssl)
	return ssl, nil
}

type opensslconn struct {
	filep      *os.File
	ssl        *C.struct_ssl_st
	localAddr  net.Addr
	mu         sync.Mutex
	remoteAddr net.Addr
	rdeadline  time.Time
	wdeadline  time.Time
}

func (c *opensslconn) Read(b []byte) (n int, err error) {
	return c.readwrite(func() C.int {
		return C.SSL_read(c.ssl, unsafe.Pointer(&b[0]), C.int(len(b)))
	})
}

func (c *opensslconn) Write(b []byte) (int, error) {
	total := len(b)
	for len(b) > 0 {
		n, err := c.writeonce(b)
		if err != nil {
			return 0, err
		}
		b = b[n:]
	}
	return total, nil
}

func (c *opensslconn) writeonce(b []byte) (int, error) {
	return c.readwrite(func() C.int {
		return C.SSL_write(c.ssl, unsafe.Pointer(&b[0]), C.int(len(b)))
	})
}

func (c *opensslconn) Close() error {
	// In OpenSSL 1.1+, SSL_set_fd does not take ownership of the
	// file descriptor hence here we aren't closing twice
	C.SSL_free(c.ssl)
	return c.filep.Close()
}

func (c *opensslconn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *opensslconn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *opensslconn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *opensslconn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rdeadline = t
	return nil
}

func (c *opensslconn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wdeadline = t
	return nil
}

func (c *opensslconn) readwrite(fn func() C.int) (int, error) {
	for {
		var (
			errcode   C.int
			rdeadline time.Time
			retval    C.int
			wdeadline time.Time
		)
		// OpenSSL is quite not thread safe
		c.mu.Lock()
		rdeadline, wdeadline = c.rdeadline, c.wdeadline
		if retval = fn(); retval <= 0 {
			errcode = C.SSL_get_error(c.ssl, retval)
		}
		c.mu.Unlock()
		if retval > 0 {
			return int(retval), nil
		}
		switch errcode {
		case C.SSL_ERROR_WANT_READ:
			if err := c.poll(C.POLLIN, rdeadline); err != nil {
				return 0, err
			}
		case C.SSL_ERROR_WANT_WRITE:
			if err := c.poll(C.POLLOUT, wdeadline); err != nil {
				return 0, err
			}
		default:
			return 0, c.errcodeToErr(errcode)
		}
	}
}

func (c *opensslconn) errcodeToErr(errcode C.int) (err error) {
	switch errcode {
	case C.SSL_ERROR_NONE:
		err = errors.New("SSL_ERROR_NONE")
	case C.SSL_ERROR_ZERO_RETURN:
		err = errors.New("SSL_ERROR_ZERO_RETURN")
	case C.SSL_ERROR_WANT_READ:
		err = errors.New("SSL_ERROR_WANT_READ")
	case C.SSL_ERROR_WANT_WRITE:
		err = errors.New("SSL_ERROR_WANT_WRITE")
	case C.SSL_ERROR_WANT_CONNECT:
		err = errors.New("SSL_ERROR_WANT_CONNECT")
	case C.SSL_ERROR_WANT_ACCEPT:
		err = errors.New("SSL_ERROR_WANT_ACCEPT")
	case C.SSL_ERROR_WANT_X509_LOOKUP:
		err = errors.New("SSL_ERROR_WANT_X509_LOOKUP")
	case C.SSL_ERROR_SYSCALL:
		err = errors.New("SSL_ERROR_SYSCALL")
	case C.SSL_ERROR_SSL:
		err = errors.New("SSL_ERROR_SSL")
	default:
		err = errors.New("SSL_ERROR_UNKNOWN")
	}
	for {
		opensslcode := C.ERR_get_error()
		if opensslcode == 0 {
			break
		}
		reason := C.GoString(C.ERR_reason_error_string(opensslcode))
		err = fmt.Errorf("%s: %w", reason, err)
	}
	return err
}

func (c *opensslconn) poll(flags C.short, deadline time.Time) error {
	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return context.DeadlineExceeded
		}
		var timeout time.Duration = -1
		if !deadline.IsZero() {
			timeout = deadline.Sub(time.Now()) * time.Millisecond
		}
		if timeout > C.INT_MAX {
			timeout = C.INT_MAX
		}
		retval := C.pollwrapper(C.int(c.filep.Fd()), flags, C.int(timeout))
		if retval > 0 {
			return nil
		}
		if retval < 0 && -retval != C.EINTR {
			return syscall.Errno(-retval)
		}
	}
}
