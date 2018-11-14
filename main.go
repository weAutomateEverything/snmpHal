package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	g "github.com/soniah/gosnmp"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	newListener("6162")
	newListener("162")

	errs := make(chan error, 2)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()
	log.Printf("terminated: %v", <-errs)
}

func newListener(port string) {
	tl := g.NewTrapListener()
	tl.OnNewTrap = handleTrap
	tl.Params = g.Default
	tl.Params.Logger = log.New(os.Stdout, "", 0)
	go func() {
		err := tl.Listen("0.0.0.0:"+port)
		if err != nil {
			log.Panicf("error in listen: %s", err)
		}
	}()
}

func handleTrap(packet *g.SnmpPacket, addr *net.UDPAddr) {
	l := log.New(os.Stdout, "", 0)

	l.Println("-----------------------")
	l.Printf("Address: %v", addr.String())

	body, _ := json.Marshal(packet)
	l.Printf("Body: %v", string(body))
	l.Printf("Community: %v", packet.Community)
	l.Println("-----------------------")
	var b bytes.Buffer
	for _, v := range packet.Variables {

		switch v.Type {
		case g.OctetString:
			c := v.Value.([]byte)
			if len(c) < 2 {
				continue
			}
			s := string(c)
			if strings.Index(s, "[") != -1 {
				continue
			}
			b.WriteString(string(c))
			b.WriteString("\n")

		}
	}
	msg := b.String()
	msg = strings.Replace(msg, "_", "\\_", -1)
	msg = strings.Replace(msg, "*", "\\*", -1)
	group, err := getGroupFromCommunity(packet.Community)
	if err != nil {
		log.Println(err.Error())
		return
	}
	http.Post(fmt.Sprintf("%v/api/alert/%v",os.Getenv("HAL"), group), "content/text", strings.NewReader(msg))
}

func getGroupFromCommunity(community string) (string, error) {
	community = strings.ToUpper(community)
	if strings.HasPrefix(community, "T") {
		return community[1:], nil
	}
	return "", fmt.Errorf("currently we support telegram messages, so the communit needs to start with a T followed by the group ID. For example T203123897. Received %v", community)

}
