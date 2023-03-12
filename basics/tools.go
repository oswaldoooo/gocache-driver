package driver_tools

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

func (s *CacheDB) connect() {
	con, _ := net.Dial("tcp", s.host+":"+strconv.Itoa(s.port))
	defer con.Close()

}
func Pack(msg *Message, checkonline bool, con net.Conn, wrsize int) (resbytes []byte, err error) {
	var rply = new(ReplayStatus)
	buff := make([]byte, wrsize)
	if checkonline {
		sendmes, err := json.Marshal(msg)
		if err == nil {
			_, err = con.Write(sendmes)
			if err == nil {
				lang, err := con.Read(buff)
				if err == nil {
					err = json.Unmarshal(buff[:lang], rply)
					if err == nil {
						switch rply.StatusCode {
						case 200:
							err = nil
							resbytes = rply.Content
						case 400:
							err = fmt.Errorf(string(rply.Content))
						default:
							err = fmt.Errorf("unknown status")
						}
					}
				}
			}
		}
	} else {
		err = fmt.Errorf("please connect cache first")
	}
	buff = make([]byte, 0) //手动释放内存
	return
}
func SimplePack(msg *Message, checkonline bool, con net.Conn) (err error) {
	_, err = Pack(msg, checkonline, con, 1*MB)
	return
}
