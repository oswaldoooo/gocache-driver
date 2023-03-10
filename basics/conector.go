package driver_tools

import (
	"encoding/json"
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
}

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func New(hostname string, port int, passwd, db string) *CacheDB {
	return &CacheDB{host: hostname, port: port, passwd: passwd, database: db}
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
func (s *CacheDB) GetKeys(keys ...string) ([]string, error) {
	msg := &Message{DB: s.database, Key: strings.Join(keys, " "), Act: 32}
	resbytes, err := Pack(msg, s.checkonline, s.connector, 1*GB)
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
