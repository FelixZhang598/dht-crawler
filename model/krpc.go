/*
 * @author Felix Zhang
 * @email Javagoshell@gmail.com
 * @datetime 2017-07-14 17:45
 */

package model


type QueryA struct {
	Id          string `bencode:"id"`
	Target      string `bencode:"target,omitempty"`
	InfoHash    string `bencode:"info_hash,omitempty"`
	ImpliedPort int `bencode:"implied_port,omitempty"`
	Port        int `bencode:"port,omitempty"`
	Token       string `bencode:"token,omitempty"`
}

type RespR struct {
	Id    string `bencode:"id"`
	Nodes  string `bencode:"nodes,omitempty"`
	Token string `bencode:"token,omitempty"`
}

type KrpcMsg struct {
	T string `bencode:"t"`
	Y string `bencode:"y"`
	V string `bencode:"v,omitempty"`

	Q string `bencode:"q,omitempty"`
	A *QueryA  `bencode:"a,omitempty"`
	R *RespR  `bencode:"r,omitempty"`
}
