/*
 * @author Felix Zhang
 * @email Javagoshell@gmail.com
 * @datetime 2017-07-14 18:21
 */

package crawler

import (
	"container/list"
	"encoding/binary"
	"encoding/hex"
	"net"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/lingfei-zhang/dht-crawler/model"
	"github.com/lingfei-zhang/dht-crawler/util"
	"github.com/zeebo/bencode"
)

type dhtCrawler struct {
	knodeQueque chan *knode
	udpConn     *net.UDPConn
	tokenLen    int
	port        int
}

type knode struct {
	nid  string
	host string
	port int
}

var bootstrap_nodes = []knode{
	{host: "router.bittorrent.com", port: 6881},
	{host: "dht.transmissionbt.com", port: 6881},
	{host: "router.utorrent.com", port: 6881},
}

func NewDHTCrawler(queueSize int, tokenLen int, port int) *dhtCrawler {
	if tokenLen > 20 {
		glog.Error("tokenLen need less than 20")
		os.Exit(1)
	}

	dc := new(dhtCrawler)
	dc.knodeQueque = make(chan *knode, queueSize)
	dc.tokenLen = tokenLen
	dc.port = port
	return dc
}

func (dc *dhtCrawler) onConsumeMsg(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	switch msg.Y {
	case "r":
		if msg.R.Nodes != "" {
			dc.processFindNodeResponse(msg, remoteAddr)
		}
	case "q":
		if msg.Q == "get_peers" {
			dc.onGetPeersRequest(msg, remoteAddr)
		} else if msg.Q == "ping" {
			dc.onPing(msg, remoteAddr)
		} else if msg.Q == "announce_peer" {

		} else if msg.Q == "find_node" {
		}
	}

}
func (dc *dhtCrawler) onPing(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	dc.ok(msg, remoteAddr)
}

func (dc *dhtCrawler) onGetPeersRequest(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	infohash := msg.A.InfoHash
	glog.Infof("get peers:%s %s", hex.EncodeToString([]byte(infohash)), remoteAddr)
	token := infohash[:dc.tokenLen]
	smsg := new(model.KrpcMsg)
	smsg.T = msg.T
	smsg.Y = "r"
	r := new(model.RespR)
	r.Id = util.RandByteString(20)
	r.Token = token
	smsg.R = r
	dc.sendKrpcMsg(msg, remoteAddr)
}

func (dc *dhtCrawler) ok(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	smsg := new(model.KrpcMsg)
	smsg.T = msg.T
	smsg.Y = "r"
	r := new(model.RespR)
	r.Id = util.RandByteString(20)
	smsg.R = r
	dc.sendKrpcMsg(smsg, remoteAddr)
}

func decodeNodes(nodes string) *list.List {
	kl := list.New()
	length := len(nodes)
	if (length % 26) != 0 {
		return kl
	}
	for i := 0; i < length; i += 26 {
		var kn *knode = new(knode)
		kn.nid = nodes[i:i+20]
		kn.host = net.IP(nodes[i+20:i+24]).String()
		kn.port = int(binary.BigEndian.Uint16([]byte(nodes[i+24: i+26])))
		kl.PushBack(kn)
	}
	return kl

}

func (dc *dhtCrawler) processFindNodeResponse(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	if len(dc.knodeQueque) < cap(dc.knodeQueque)-50 {
		kl := decodeNodes(msg.R.Nodes)
		for e := kl.Front(); e != nil; e = e.Next() {
			k := e.Value.(*knode)
			if len(k.nid) != 20 {
				continue
			}
			if k.port < 1 || k.port > 65535 {
				continue
			}
			dc.knodeQueque <- k
		}
	}
}

func (dc *dhtCrawler) findNode(remoteAddr *net.UDPAddr) {
	msg := new(model.KrpcMsg)
	msg.T = util.RandByteString(2)
	msg.Y = "q"
	msg.Q = "find_node"
	a := new(model.QueryA)
	a.Id = util.RandByteString(20)
	a.Target = util.RandByteString(20)
	msg.A = a
	dc.sendKrpcMsg(msg, remoteAddr)
}

func (dc *dhtCrawler) monitor() {
	go func() {
		for range time.Tick(10 * time.Second) {
			glog.Info("the knodeQueque length:", len(dc.knodeQueque))
		}
	}()
}

func (dc *dhtCrawler) joinDHT() {
	for _, value := range bootstrap_nodes {
		ips, _ := net.LookupIP(value.host)
		for _, ip := range ips {
			if ip.To4() != nil {
				dc.findNode(&net.UDPAddr{IP: ips[0], Port: value.port})
				break
			}
		}
	}
}

func (dc *dhtCrawler) reJoinDHT() {
	go func() {
		for range time.Tick(3 * time.Second) {
			if len(dc.knodeQueque) == 0 {
				dc.joinDHT()
			}
		}
	}()
}

func (dc *dhtCrawler) autoFindNode() {
	go func() {
		for {
			kn := <-dc.knodeQueque
			dc.findNode(&net.UDPAddr{IP: net.ParseIP(kn.host), Port: kn.port})
		}
	}()
}

func (dc *dhtCrawler) sendKrpcMsg(msg *model.KrpcMsg, remoteAddr *net.UDPAddr) {
	bs, err := bencode.EncodeBytes(msg)
	if err != nil {
		glog.Error("bencode encode error: ", err)
	}
	_, err = dc.udpConn.WriteToUDP(bs, remoteAddr)
	if err != nil {
		glog.Error("send krpc msg error: ", err)
	}
}

func (dc *dhtCrawler) Start() {
	var err error
	dc.udpConn, err = net.ListenUDP("udp", &net.UDPAddr{Port: dc.port})
	if err != nil {
		glog.Error("listen udp ", err)
	}
	glog.Info("the udp server start listen ", dc.udpConn.LocalAddr())

	dc.reJoinDHT()
	dc.monitor()
	dc.autoFindNode()

	data := make([]byte, 1024)

	for {
		n, remoteAddr, err := dc.udpConn.ReadFromUDP(data)

		if err != nil {
			glog.Errorf("read: %s", err)
		}
		krpcMsg := new(model.KrpcMsg)

		err = bencode.DecodeBytes(data[:n], krpcMsg)
		if err != nil {
			glog.Error("decode error :", err)
		}
		go dc.onConsumeMsg(krpcMsg, remoteAddr)
	}
}
