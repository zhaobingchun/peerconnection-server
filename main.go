package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zhaobingchun/peerconnection-server/clients"
)

func init() {
	http.HandleFunc("/sign_in", func(w http.ResponseWriter, r *http.Request) {
		clients.AddMember(w, r)
	})
	http.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("wait: %+v\n", r.URL.RawQuery)
		c := clients.GetPeer(r)
		if c != nil {
			if c.Queue == nil {
				c.Queue = make(chan clients.QueueData, 10)
			}
			resp := <-c.Queue
			fmt.Printf("peer[%d] recv resp[%s]\n", c.ID, string(resp.Data))
			c.WriteResponse(w, resp.PeerID, resp.Data)
		}
	})
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("message: %+v\n", r.URL.RawQuery)
		c := clients.GetPeer(r)
		other := clients.GetOtherPeer(r)
		if other != nil {
			bys, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			fmt.Printf("recv bys[%s]\n", string(bys))
			other.Queue <- clients.QueueData{
				PeerID: c.ID,
				Data:   bys,
			}
		}
		w.Write([]byte(""))
	})
	http.HandleFunc("/sign_out", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("sign_out: %+v\n", r.URL.RawQuery)
		clients.RemovePeer(r)
		w.Write([]byte(""))
	})
	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("quit: %+v\n", r.URL.RawQuery)
	})
}

func main() {
	http.ListenAndServe(":8888", nil)
}
