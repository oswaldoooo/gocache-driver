package driver_tools_v2

import (
	"io"
	"net"
	"strconv"
	"strings"

	driver_tools "github.com/oswaldoooo/gocache-driver/basics"
)

type Data map[string]any
type CacheDB_V2 struct {
	host        string
	port        int
	passwd      string
	database    string
	checkonline bool
	connector   io.ReadWriteCloser
	buffer_size int
}

const ( //signal
	PING = 0x10
)

var (
	NET_ERROR   = Str_Error("network error")
	NOT_ALLOWED = Str_Error("not allowed")
)

// this return never be nil
func NewCacheDB_V2(host string, port int, passwd, database string) *CacheDB_V2 {
	return &CacheDB_V2{host: host, port: port, passwd: passwd, database: database}
}
func (s *CacheDB_V2) Connect() error {

	ip, err := ToNetIP(s.host)
	if err != nil {
		return err
	}
	var con net.Conn
	con, err = net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: s.port})
	if err == nil {
		s.connector = con
		err = s.Ping()
		if err == nil {
			s.checkonline = true
		}
	}
	if err != nil {
		err = NET_ERROR
		s.Close()
	}
	return err
}
func (s *CacheDB_V2) Ping() error {
	content := []byte{}
	return writeTo(s.connector, PING, content)
}
func (s *CacheDB_V2) PingX() error { //advance ping for make sure can visit db
	var (
		err     error
		content []byte
	)
	if len(s.passwd) > 0 {
		data_source := make(Data)
		data_source["passwd"] = s.passwd
		if len(s.database) > 0 {
			data_source["database"] = s.database
		}
		content, err = Encode(&data_source)
		if err != nil {
			Logger.Println("[error] encode datasource failed", err.Error())
		}
	}
	err = writeTo(s.connector, PING, content)
	if err == nil && len(content) > 0 {
		//need wait response
		err = Read(s.connector, func(u uint8, b []byte) error {
			var err error
			switch u {
			case OK:
				err = nil
			case ERROR:
				err = NOT_ALLOWED
			default:
				err = Str_Error("unknown command")
			}
			return err
		})
	}
	return err
}
func (s *CacheDB_V2) Close() error {
	var err error
	err = s.connector.Close()
	return err
}
func (s *CacheDB_V2) SetKey(key, value string) (err error) {
	msg := &driver_tools.Message{DB: s.database, Key: key, Value: []byte(value), Act: 31}
	_, err = Pack_V2(msg, s.connector)
	return
}

func (s *CacheDB_V2) CompareAndSetKey(key, value string) (version uint32, err error) {
	msg := &driver_tools.Message{DB: s.database, Key: key, Value: []byte(value), Act: 11}
	var resbytes []byte
	resbytes, err = Pack_V2(msg, s.connector)
	if err == nil {
		var ans map[string]any = make(map[string]any)
		err = Decode(resbytes, &ans)
		if err == nil {
			rawversion, _ := strconv.ParseUint(ans["version"].(string), 10, 32)
			version = uint32(rawversion)
		}
	}
	return
}
func (s *CacheDB_V2) GetKeys(keys ...string) ([]string, error) {
	msg := &driver_tools.Message{DB: s.database, Key: strings.Join(keys, " "), Act: 32}
	resbytes, err := Pack_V2(msg, s.connector)
	if err == nil {
		var restr []string
		err = Decode(resbytes, &restr)
		if err == nil {
			if len(restr) > 0 {
				return restr, nil
			} else {
				return nil, nil
			}
		}
	}
	return nil, err
}

func (s *CacheDB_V2) GetKeysContain(subkey string) (resmap map[string][]byte, err error) {
	msg := &driver_tools.Message{Act: 35, DB: s.database, Key: subkey}
	resbytes, err := Pack_V2(msg, s.connector)
	if err == nil {
		err = Decode(resbytes, &resmap)
	}
	return
}
func (s *CacheDB_V2) GetAllKeys() (resmap map[string][]byte, err error) {
	msg := &driver_tools.Message{DB: s.database, Act: 34}
	resbytes, err := Pack_V2(msg, s.connector)
	if err == nil {
		err = Decode(resbytes, &resmap)
	}
	return
}
func (s *CacheDB_V2) DeleteKeys(keys ...string) (err error) {
	msg := &driver_tools.Message{DB: s.database, Key: strings.Join(keys, " "), Act: 33}
	_, err = Pack_V2(msg, s.connector)
	return
}
func (s *CacheDB_V2) CreateDB() (err error) {
	msg := &driver_tools.Message{DB: s.database, Act: 36}
	_, err = Pack_V2(msg, s.connector)
	return
}
func (s *CacheDB_V2) DropDB() (err error) {
	msg := &driver_tools.Message{DB: s.database, Act: 38}
	_, err = Pack_V2(msg, s.connector)
	return
}
func (s *CacheDB_V2) Save() (err error) {
	msg := &driver_tools.Message{DB: s.database, Act: 37}
	_, err = Pack_V2(msg, s.connector)
	return
}
