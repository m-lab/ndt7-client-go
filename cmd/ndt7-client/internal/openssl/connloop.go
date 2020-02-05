// +build ndt7openssl

package openssl

import (
	"os"

	// #include <errno.h>
	// #include <poll.h>
	//
	// #include <openssl/err.h>
	// #include <openssl/ssl.h>
	//
	// #cgo LDFLAGS: -lssl -lcrypto
	//
	// /* pollwrapper wraps poll to return the value of errno to the
	//    caller because CGO cannot access errno. */
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
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// MessageType is the type of a message you send to the connection loop.
type MessageType int

const (
	// MessageClose is a request to close the stream.
	MessageClose = MessageType(iota)

	// MessageRead is a request to read bytes.
	MessageRead

	// MessageWrite is a request to write bytes.
	MessageWrite
)

// Message is a message you send to the connection loop.
type Message struct {
	Count    int
	Data     []byte
	Deadline time.Time
	Done     chan interface{}
	Error    error
	Type     MessageType
}

// NewMessage creates a new empty message.
func NewMessage() *Message {
	return &Message{
		Done: make(chan interface{}, 1),
	}
}

// RunConnLoop runs the connection loop. Takes ownership of the |fp| and |ssl| args.
func RunConnLoop(fp *os.File, ssl *C.struct_ssl_st, inch <-chan *Message) {
	for msg := range inch {
		switch msg.Type {
		case MessageClose:
			// In OpenSSL 1.1+, SSL_set_fd does not take ownership of the
			// file descriptor hence here we aren't closing twice
			msg.Error = fp.Close()
			C.SSL_free(ssl)
			msg.Done <- true
			return
		case MessageRead:
			msg.Count, msg.Error = read(ssl, msg.Data, msg.Deadline)
			msg.Done <- true
		case MessageWrite:
			msg.Count, msg.Error = write(ssl, msg.Data, msg.Deadline)
			msg.Done <- true
		}
	}
}

func read(
	ssl *C.struct_ssl_st,
	data []byte,
	deadline time.Time,
) (int, error) {
	return readwrite(ssl, data, deadline, doread)
}

func doread(
	ssl *C.struct_ssl_st,
	data []byte,
) C.int {
	return C.SSL_read(ssl, unsafe.Pointer(&data[0]), C.int(len(data)))
}

func write(
	ssl *C.struct_ssl_st,
	data []byte,
	deadline time.Time,
) (int, error) {
	total := len(data)
	for len(data) > 0 {
		count, err := writeonce(ssl, data, deadline)
		if err != nil {
			return 0, err
		}
		data = data[count:]
	}
	return total, nil
}

func writeonce(
	ssl *C.struct_ssl_st,
	data []byte,
	deadline time.Time,
) (int, error) {
	return readwrite(ssl, data, deadline, dowrite)
}

func dowrite(
	ssl *C.struct_ssl_st,
	data []byte,
) C.int {
	return C.SSL_write(ssl, unsafe.Pointer(&data[0]), C.int(len(data)))
}

func readwrite(
	ssl *C.struct_ssl_st,
	data []byte,
	deadline time.Time,
	fn func(*C.struct_ssl_st, []byte) C.int,
) (int, error) {
	for {
		var retval C.int
		if retval = fn(ssl, data); retval > 0 {
			return int(retval), nil
		}
		switch errcode := C.SSL_get_error(ssl, retval); errcode {
		case C.SSL_ERROR_WANT_READ:
			if err := poll(ssl, C.POLLIN, deadline); err != nil {
				return 0, err
			}
		case C.SSL_ERROR_WANT_WRITE:
			if err := poll(ssl, C.POLLOUT, deadline); err != nil {
				return 0, err
			}
		default:
			return 0, errcodeToErr(errcode)
		}
	}
}

func errcodeToErr(errcode C.int) (err error) {
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

func poll(ssl *C.struct_ssl_st, flags C.short, deadline time.Time) error {
	fd := C.SSL_get_fd(ssl)
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
		retval := C.pollwrapper(fd, flags, C.int(timeout))
		if retval > 0 {
			return nil
		}
		if retval < 0 && -retval != C.EINTR {
			return syscall.Errno(-retval)
		}
	}
}
