package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	cjdns "github.com/ehmry/go-cjdns/admin"
)

func cjdnsConn(addr string, pwd string) (*cjdns.Conn, error) {
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
	Cjdns   *cjdns.Conn
	Graph   string
	Updated time.Time
}

func (pfv *PFVis) Refresh() error {
	nodeName := func(addr string) string {
		parts := strings.Split(addr, ".")
		return parts[5]
	}
	graph := "digraph pathfinder {\n"
	nodes := map[string]string{}
	links := ""
	for i := 0; i < 1000000; i++ {
		link, err := pfv.Cjdns.NodeStore_getLink("", i)
		if err != nil && err.Error() == "not_found" {
			break
		}
		if err != nil {
			return err
		}
		parent := nodeName(link.Parent)
		child := nodeName(link.Child)
		nodes[parent] = link.Parent
		nodes[child] = link.Child
		links += fmt.Sprintf("\t%s -> %s [label=\"cost:%d\"];\n", parent, child, link.LinkCost)
	}
	for label, node := range nodes {
		graph += fmt.Sprintf("\t%s [label=\"%s\"];\n", label, node)
	}
	pfv.Graph = graph + links + "}"
	pfv.Updated = time.Now()
	return nil
}

func (pfv *PFVis) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write(bytes.NewBufferString(pfv.Graph).Bytes())
}

func main() {
	laddr := flag.String("l", ":8080", "Network address to listen on")
	caddr := flag.String("c", "127.0.0.1:11234", "Network address of cjdns admin API")
	pwd := flag.String("p", "NONE", "Password of cjdns admin API")
	flag.Parse()

	c, err := cjdnsConn(*caddr, *pwd)
	if err != nil {
		log.Fatal(err.Error())
	}

	pfv := &PFVis{Cjdns: c}

	go func() {
		log.Printf("asking %s for links", *caddr)
		for {
			time.Sleep(2 * time.Second)
			if err = pfv.Refresh(); err != nil {
				log.Printf("error: %s", err)
			}
		}
	}()

	log.Printf("listening on %s\n", *laddr)
	err = http.ListenAndServe(*laddr, pfv)
	if err != nil {
		log.Fatal(err.Error())
	}
}
