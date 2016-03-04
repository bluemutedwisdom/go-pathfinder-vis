package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	cjdns "github.com/ehmry/go-cjdns/admin"
)

func cjdnsAdmin(addr string, pwd string) (*cjdns.Conn, error) {
	a := strings.Split(addr, ":")
	port, err := strconv.Atoi(a[1])
	if err != nil {
		return nil, err
	}

	c := &cjdns.CjdnsAdminConfig{
		Addr:     a[0],
		Port:     port,
		Password: pwd,
	}
	admin, err := cjdns.Connect(c)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

type PFVis struct {
	Cjdns *cjdns.Conn
	Links []*cjdns.StoreLink
}

type Link struct {
	Parent, Child    string
	Cost, BestParent int
}

func (pfv *PFVis) Refresh() error {
	links := []*cjdns.StoreLink{}
	for i := 0; i < 1000000; i++ {
		link, err := pfv.Cjdns.NodeStore_getLink("", i)
		if err != nil && err.Error() == "not_found" {
			pfv.Links = links
			return nil
		}
		if err != nil {
			return err
		}
		links = append(links, link)
	}
	return nil
}

func (pfv *PFVis) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write(bytes.NewBufferString("asdsadsdsada").Bytes())
}

func main() {
	laddr := flag.String("l", ":8080", "Network address to listen on")
	caddr := flag.String("c", "127.0.0.1:11234", "Network address of cjdns admin API")
	pwd := flag.String("p", "NONE", "Password of cjdns admin API")
	flag.Parse()

	c, err := cjdnsAdmin(*caddr, *pwd)
	if err != nil {
		log.Fatal(err.Error())
	}

	pfv := &PFVis{Cjdns: c}

	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Printf("asking %s for links", *caddr)
			pfv.Refresh()
			log.Printf("links: %s", pfv.Links)
		}
	}()

	log.Printf("listening on %s\n", *laddr)
	err = http.ListenAndServe(*laddr, pfv)
	if err != nil {
		log.Fatal(err.Error())
	}
}
