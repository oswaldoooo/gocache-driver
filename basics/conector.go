package driver_tools

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type CacheDB struct {
	host        string
	port        int
	passwd      string
	database    string
	checkonline bool
	connector   net.Conn
	buffer_size int
}

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func New(hostname string, port int, passwd, db string) *CacheDB {
	return &CacheDB{host: hostname, port: port, passwd: passwd, database: db, buffer_size: 10 * MB}
}
func (s *CacheDB) Connect() error {
	con, err := net.Dial("tcp", s.host+":"+strconv.Itoa(s.port))
	if err == nil {
		s.checkonline = true
		s.connector = con
	} else {
		s.checkonline = false
	}
	return err
}
func (s *CacheDB) SetKey(key, value string) (err error) {
	msg := &Message{DB: s.database, Key: key, Value: []byte(value), Act: 31}
	err = SimplePack(msg, s.checkonline, s.connector)
	return
}

func (s *CacheDB) CompareAndSetKey(key, value string) (version uint32, err error) {
	msg := &Message{DB: s.database, Key: key, Value: []byte(value), Act: 11}
	var resbytes []byte
	resbytes, err = Pack(msg, s.checkonline, s.connector, s.buffer_size)
	if err == nil {
		if len(resbytes) == 2 {
			version = uint32(resbytes[0])*256 + uint32(resbytes[1])
		} else {
			err = fmt.Errorf("unknown error,origin datapacket %v", resbytes)
		}
	}
	return
}
func (s *CacheDB) GetKeys(keys ...string) ([]string, error) {
	msg := &Message{DB: s.database, Key: strings.Join(keys, " "), Act: 32}
	resbytes, err := Pack(msg, s.checkonline, s.connector, s.buffer_size)
	if err == nil {
		var restr []string
		err = json.Unmarshal(resbytes, &restr)
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
func (s *CacheDB) GetKeysContain(subkey string) (resmap map[string][]byte, err error) {
	msg := &Message{Act: 341, DB: s.database, Key: subkey}
	resbytes, err := Pack(msg, s.checkonline, s.connector, s.buffer_size)
	if err == nil {
		err = json.Unmarshal(resbytes, &resmap)
	}
	return
}
func (s *CacheDB) GetAllKeys() (resmap map[string][]byte, err error) {
	msg := &Message{DB: s.database, Act: 322}
	resbytes, err := Pack(msg, s.checkonline, s.connector, s.buffer_size)
	if err == nil {
		err = json.Unmarshal(resbytes, &resmap)
	}
	return
}
func (s *CacheDB) DeleteKeys(keys ...string) (err error) {
	msg := &Message{DB: s.database, Key: strings.Join(keys, " "), Act: 33}
	err = SimplePack(msg, s.checkonline, s.connector)
	return
}
func (s *CacheDB) CreateDB() (err error) {
	msg := &Message{DB: s.database, Act: 310}
	err = SimplePack(msg, s.checkonline, s.connector)
	return
}
func (s *CacheDB) DropDB() (err error) {
	msg := &Message{DB: s.database, Act: 330}
	err = SimplePack(msg, s.checkonline, s.connector)
	return
}
func (s *CacheDB) Save() (err error) {
	msg := &Message{DB: s.database, Act: 320}
	err = SimplePack(msg, s.checkonline, s.connector)
	return
}
func (s *CacheDB) Close() {
	s.connector.Close()
}
