package clients

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	clients   sync.Map
	gloableID uint32 = 0
)

// Client Client
type Client struct {
	ID          uint32
	Name        string
	ConnectTime time.Time
	Queue       chan QueueData
}

// QueueData QueueData
type QueueData struct {
	PeerID uint32
	Data   []byte
}

// GetEntry GetEntry
func (c *Client) GetEntry() string {
	return fmt.Sprintf("%s,%d,%d\n", c.Name, c.ID, 1)
}

// WriteResponse WriteResponse
func (c *Client) WriteResponse(w http.ResponseWriter, peerid uint32, data []byte) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Header().Set("Pragma", fmt.Sprintf("%d", peerid))
	w.Write(data)
}

// TimeOut TimeOut
func (c *Client) TimeOut() bool {
	return c.Queue == nil && time.Now().Sub(c.ConnectTime) > 30*time.Second
}

// AddMember AddMember
func AddMember(w http.ResponseWriter, r *http.Request) {
	c := &Client{
		ID:          atomic.AddUint32(&gloableID, 1),
		Name:        r.URL.RawQuery,
		ConnectTime: time.Now(),
	}
	NotifyOtherMember(c)
	clients.Store(c.ID, c)
	resp := BuildResponseForNewMember(c)
	c.WriteResponse(w, c.ID, []byte(resp))
}

// BuildResponseForNewMember BuildResponseForNewMember
func BuildResponseForNewMember(c *Client) string {
	var (
		resp = c.GetEntry()
	)
	clients.Range(func(k, v interface{}) bool {
		if k == c.ID {
			return true
		}
		resp += v.(*Client).GetEntry()
		return true
	})
	return resp
}

// NotifyOtherMember NotifyOtherMember
func NotifyOtherMember(other *Client) {
	clients.Range(func(k, v interface{}) bool {
		if k == other.ID {
			return true
		}
		data := QueueData{
			PeerID: v.(*Client).ID,
			Data:   []byte(other.GetEntry()),
		}
		v.(*Client).Queue <- data
		return true
	})
}

// GetClient GetClient
func GetClient(id uint32) *Client {
	v, ok := clients.Load(id)
	if !ok {
		return nil
	}
	c, _ := v.(*Client)
	return c
}

// GetPeer GetPeer
func GetPeer(r *http.Request) *Client {
	peerID := r.FormValue("peer_id")
	id, _ := strconv.Atoi(peerID)
	return GetClient(uint32(id))
}

// GetOtherPeer GetOtherPeer
func GetOtherPeer(r *http.Request) *Client {
	peerID := r.FormValue("to")
	id, _ := strconv.Atoi(peerID)
	return GetClient(uint32(id))
}

// RemovePeer RemovePeer
func RemovePeer(r *http.Request) {
	peerID := r.FormValue("peer_id")
	id, _ := strconv.Atoi(peerID)
	clients.Delete(uint32(id))
}

// ProcessTimeOut ProcessTimeOut
func ProcessTimeOut() {
	clients.Range(func(k, v interface{}) bool {
		if v.(*Client).TimeOut() {
			fmt.Printf("peer[%d] timeout", k)
			clients.Delete(k)
		}
		return true
	})
}

// GC GC
func GC() {
	ProcessTimeOut()
	time.AfterFunc(5*time.Second, ProcessTimeOut)
}

func init() {
	go GC()
}
