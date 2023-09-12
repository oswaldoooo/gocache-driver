package driver_tools

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

var (
	CON_ERR = fmt.Errorf("connection error")
	Erron   = []string{}
	rwmutex sync.RWMutex //control erron read and write
)

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
				} else {
					Erron = append(Erron, err.Error())
					err = CON_ERR
				}
			} else {
				Erron = append(Erron, err.Error())
				err = CON_ERR
			}
		}
	} else {
		err = fmt.Errorf("please connect cache first")
	}
	return
}
func SimplePack(msg *Message, checkonline bool, con net.Conn) (err error) {
	_, err = Pack(msg, checkonline, con, 1*MB)
	return
}
