package driver_tools_v2

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	driver_tools "github.com/oswaldoooo/gocache-driver/basics"
)

const ( //status code
	OK    = 0x01
	ERROR = 0x02
)

var (
	Bigendian    = new(BigEndian)
	Littleendian = new(LittleEndian)
	Decode       func([]byte, any) error
	Encode       func(any) ([]byte, error)
	Logger       *log.Logger = log.New(os.Stderr, "[error]", log.Lshortfile|log.Ltime)
)

type endian interface {
	Uint(src []byte, n uint) uint64
	Write(v uint64, n int, p []byte) error
}

type BigEndian struct {
}
type LittleEndian struct {
}

func (s *BigEndian) Uint(src []byte, n uint) uint64 {
	step := n
	var length uint = uint(len(src))
	if length < step {
		step = length
	}
	var ans uint64
	if step > 0 {
		var i uint
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "error pos", i, "src length", len(src))
				os.Exit(1)
			}
		}()
		for i = 0; i < step; i++ {
			ans *= 256
			ans += uint64(src[i])
		}
	}
	return ans
}
func (s *BigEndian) Write(v uint64, n int, p []byte) error {
	if len(p) < n {
		return Str_Error("p out of range n")
	}
	var i int
	for i < n {
		p[n-1-i] = uint8(v % 256)
		if v > 0 {
			v /= 256
		}
		i++
	}
	return nil
}

func (s *LittleEndian) Uint(src []byte, n uint) uint64 {
	step := n
	var length uint = uint(len(src))
	if length < step {
		step = length
	}
	var ans uint64
	if step > 1 {
		var i int
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "error pos", i, "src length", len(src), "step", step)
				os.Exit(1)
			}
		}()
		for i = int(step - 1); i >= 0; i-- {
			ans *= 256
			ans += uint64(src[i])
		}
	}
	return ans
}
func (s *LittleEndian) Write(v uint64, n int, p []byte) error {
	if len(p) < n {
		return Str_Error("p out of range n")
	}
	var i int
	for i < n {
		p[i] = uint8(v % 256)
		if v > 0 {
			v /= 256
		}
		i++
	}
	return nil
}

type Reader struct {
	cache_buffer []byte
	mux          sync.Mutex
}

func (s *Reader) Read(reader io.Reader, n int, end endian) uint64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	// lang, err := reader.Read(s.cache_buffer[0:n])
	err := read(reader, s.cache_buffer[0:n])
	if err == nil {
		if n == 1 {
			return uint64(s.cache_buffer[0])
		}
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "[panic error] lang", n, len(s.cache_buffer), r)
				os.Exit(1)
			}
		}()
		return end.Uint(s.cache_buffer[0:n], uint(n))
	}
	return 0
}
func (s *Reader) RawRead(reader io.Reader, n int) []byte {
	s.mux.Lock()
	defer s.mux.Unlock()
	var (
		ans []byte = nil
	)
	err := read(reader, s.cache_buffer[:n])
	if err == nil {
		ans = make([]byte, n)
		copy(ans, s.cache_buffer[:n])
	}
	return ans
}
func NewReader(size uint64) *Reader {
	return &Reader{cache_buffer: make([]byte, size)}
}

type Str_Error string

func (s Str_Error) Error() string {
	return string(s)
}
func writeTo(out io.Writer, code uint8, content []byte) error {
	newcontent := make([]byte, len(content)+3)
	newcontent[0] = code
	err := Bigendian.Write(uint64(len(content)), 2, newcontent[1:3])
	if err == nil && len(content) > 0 {
		copy(newcontent[3:], content)
		_, err = out.Write(newcontent)
	}
	return err
}
func Read(in io.Reader, callback func(uint8, []byte) error) error {
	var (
		err  error
		data []byte
	)
	reader := NewReader(10 << 10)
	code := reader.Read(in, 1, Bigendian)
	data_len := reader.Read(in, 2, Bigendian)
	if data_len > 0 {
		data = reader.RawRead(in, int(data_len))
	}
	err = callback(uint8(code), data)
	return err
}

// read full byte array
func read(in io.Reader, p []byte) error {
	var (
		err   error
		n     int
		start int
		lang  = len(p)
	)
	n, err = in.Read(p[start:])
	for err == nil && start < lang {
		start += n
		if start >= lang {
			break
		}
		n, err = in.Read(p[start:])
		if n == 0 {
			err = io.EOF
		}
	}
	return err
}

func Pack_V2(msg *driver_tools.Message, con io.ReadWriter) (resbytes []byte, err error) {
	var content []byte
	content, err = Encode(msg)
	if err == nil {
		err = writeTo(con, uint8(msg.Act), content)
		if err == nil {
			// var reply *driver_tools.ReplayStatus = new(driver_tools.ReplayStatus)
			err = Read(con, func(u uint8, b []byte) error {
				var err error
				switch u {
				case OK:
					err = nil
					resbytes = b
				case ERROR:
					err = Str_Error(b)
				default:
					Logger.Println("unknow code", u)
					err = Str_Error("unknown status")
				}
				return err
			})
		}
	}

	return
}

func ToNetIP[T ~string](s T) (ip net.IP, err error) {
	arr := strings.Split(string(s), ".")
	if len(arr) != 4 {
		err = Str_Error(s + " is not ip address")
		return
	}
	ip = make(net.IP, 4)
	var vl uint64
	for i := 0; i < 4; i++ {
		vl, err = strconv.ParseUint(arr[i], 10, 8)
		if err == nil {
			ip[i] = byte(vl)
		} else {
			return
		}
	}
	return
}

func WriteTo[T any](out io.Writer, code uint8, v *T) error {
	content, err := Encode(v)
	if err == nil {
		err = writeTo(out, code, content)
	}
	return err
}
func ReadFrom[T any](in io.Reader, v *T) error {
	var rawbytes []byte
	err := Read(in, func(u uint8, b []byte) error {
		var err error
		switch u {
		case OK:
			if len(b) > 0 {
				rawbytes = b
			}
		case ERROR:
			err = Str_Error(b)
		default:
			err = Str_Error("unknown status" + strconv.FormatUint(uint64(u), 8))
		}
		return err
	})
	if err == nil && len(rawbytes) > 0 {
		err = Decode(rawbytes, v)
	}
	return err
}
