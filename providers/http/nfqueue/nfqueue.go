// Package nfqueue implements a HTTP provider for solving the HTTP-01 challenge using nfqueue
// by captureing http challange pacet in fly and answering it by ourself
// This solver needs a TCP server attached on request port, and need root or CAP_NET_ADMIN
package nfqueue

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-acme/lego/v4/log"

	gnfqueue "github.com/florianl/go-nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// HTTPProvider implements HTTPProvider for `http-01` challenge.
type HTTPProvider struct {
	port    string
	context context.Context
	cancel  context.CancelFunc
}

// a struct holds thing that packet handler should know about
type chalnfq struct {
	domain  string
	token   string
	keyAuth string
	queue   *gnfqueue.Nfqueue
}

var sopt = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

// NewHttpDpiProvider returns a HTTPProvider instance with a configured port.
func NewHttpDpiProvider(port string) (*HTTPProvider, error) {

	c := &HTTPProvider{
		port: port,
	}

	return c, nil
}

// craftkeyauthresponse carft acme challange response in HTTP level
func craftkeyauthresponse(keyAuth string) []byte {
	var reply []byte
	reply = fmt.Append(reply, "HTTP/1.1 200 OK\r\n")
	reply = fmt.Append(reply, "Content-Type: text/plain\r\n")
	reply = fmt.Append(reply, "server: go-acme-nfqueue\r\n")
	reply = fmt.Appendf(reply, "Content-Length: %d\r\n", len(keyAuth))
	reply = fmt.Append(reply, "\r\n", keyAuth)

	return reply
}

// setFirewallRule set rule in firewall INPUT chain so we can sniff on
// with --queue-bypass option even if this crash without clean webserver will listen
// iptables {on} INPUT -p tcp --dport {Port} -j NFQUEUE --queue-num 8555 --queue-bypass
func setFirewallRule(on bool, port string) error {
	// google's nft api is unstable, so we run command as-is
	var onoff string
	if on {
		onoff = "-I"
	} else {
		onoff = "-D"
	}
	out, err := exec.Command("iptables", onoff, "INPUT", "-p", "tcp", "--dport", port, "-j", "NFQUEUE", "--queue-num", "8555", "--queue-bypass").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	err = exec.Command("ip6tables", onoff, "INPUT", "-p", "tcp", "--dport", port, "-j", "NFQUEUE", "--queue-num", "8555", "--queue-bypass").Run()
	if err != nil {
		return fmt.Errorf("%s", out)
	}
	return nil
}

// craft packet
func craftReplyandSend(keyAuth string, inputpacket gopacket.Packet, dst net.IP) error {
	outbuffer := gopacket.NewSerializeBuffer()
	inputTcp := inputpacket.Layer(layers.LayerTypeTCP).(*layers.TCP)
	inputIPL := inputpacket.NetworkLayer()

	httplayer := gopacket.Payload(craftkeyauthresponse(keyAuth))
	tcplayer := &layers.TCP{
		// we reply back so reverse src and dst ports
		SrcPort: inputTcp.DstPort,
		DstPort: inputTcp.SrcPort,
		Ack:     inputTcp.Seq + uint32(len(inputTcp.Payload)),
		Seq:     inputTcp.Ack,
		Window:  1,
		PSH:     true,
		ACK:     true,
		// we want to finish TCP after this packet so set fin
		FIN: true,
	}
	// answer is same with same protocal, so we use input's layer
	tcplayer.SetNetworkLayerForChecksum(inputIPL)
	gopacket.SerializeLayers(outbuffer, sopt, tcplayer, httplayer)
	// send http reply
	sendPacket(outbuffer.Bytes(), &dst)

	// need to ACK for the server FIN so acme server can close connection
	outbuffer.Clear()
	tcplayer.ACK = true
	tcplayer.PSH = false
	tcplayer.Seq = tcplayer.Seq + uint32(len(httplayer.Payload())) + 1
	tcplayer.Ack = inputTcp.Seq + uint32(len(inputTcp.Payload)) + 1

	tcplayer.SetNetworkLayerForChecksum(inputIPL)
	gopacket.SerializeLayers(outbuffer, sopt, tcplayer)
	// sleep some time here so acme server sent its FIN+ACK when this arrives
	// alternatives are: 1. send rst here instead 2. actually trace connection
	time.Sleep(time.Millisecond * 10)
	sendPacket(outbuffer.Bytes(), &dst)

	return nil
}

func craftRSTbyte(inpkt gopacket.Packet) []byte {
	tcpl := inpkt.Layer(layers.LayerTypeTCP).(*layers.TCP)
	tcpl.SetNetworkLayerForChecksum(inpkt.NetworkLayer())
	ipl := inpkt.LayerClass(layers.LayerClassIPNetwork).(gopacket.SerializableLayer)
	buf := gopacket.NewSerializeBuffer()
	tcpl.RST = true
	tcpl.ACK = true
	gopacket.SerializeLayers(buf, sopt, ipl, tcpl)
	return buf.Bytes()
}

// sendPacket sends packet: TODO: call cleanup if errors out
func sendPacket(packet []byte, DstIP *net.IP) error {
	var err error
	con, err := net.Dial("ip:6", DstIP.String())
	if err != nil {
		return err
	}
	_, err = con.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

// handlePacket handles packet input
func (q chalnfq) handlePacket(qupkt gnfqueue.Attribute) int {
	id := *qupkt.PacketID
	dopt := gopacket.DecodeOptions{
		NoCopy: true,
		Lazy:   false,
	}
	var ipLType gopacket.LayerType
	// Hwprotocol here is ethernet frame protocol header
	if *qupkt.HwProtocol == 0x0800 {
		//ipv4
		ipLType = layers.LayerTypeIPv4
	} else if *qupkt.HwProtocol == 0x86DD {
		ipLType = layers.LayerTypeIPv6
	} else {
		q.queue.SetVerdict(id, gnfqueue.NfAccept)
		return 0
	}
	packetin := gopacket.NewPacket(*qupkt.Payload, ipLType, dopt)
	// Get actual TCP data from this layer
	tcpLayer := packetin.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		q.queue.SetVerdict(id, gnfqueue.NfAccept)
		return 0
	}
	inputTcp := tcpLayer.(*layers.TCP)
	// get destination IP here, this is sent from other side, so src is other side
	otherend := net.IP(packetin.NetworkLayer().NetworkFlow().Src().Raw())
	// this should be HTTP payload as this is webserver
	httpPayload, err := http.ReadRequest(bufio.NewReader((bytes.NewReader(inputTcp.LayerPayload()))))
	if err != nil {
		q.queue.SetVerdict(id, gnfqueue.NfAccept)
		return 0
	}
	// check this request ask for token
	chalPath := fmt.Sprintf("/.well-known/acme-challenge/%s", q.token)
	if httpPayload.URL.Path == chalPath {
		// we got the token!
		// forge our new reply
		log.Infof("[%s] Injecting key authentication", q.domain)
		err := craftReplyandSend(q.keyAuth, packetin, otherend)
		if err != nil {
			return 0
		}
		// mark incomming packet as RST so backend server ignore and close session
		rstpk := craftRSTbyte(packetin)
		err = q.queue.SetVerdictModPacket(id, gnfqueue.NfAccept, rstpk)
		if err != nil {
			fmt.Print("modpacket err", err)
		}
		// packet sent, end of function
		return 0
	} else {
		q.queue.SetVerdict(id, gnfqueue.NfAccept)
		return 0
	}

}

// serve runs server by sniffing packets on firewall and inject response into it.
func (w *HTTPProvider) serve(domain, token, keyAuth string) error {
	// run nfqueue start
	config := gnfqueue.Config{
		NfQueue:      8555,
		MaxPacketLen: 0xFFFF,
		MaxQueueLen:  0xFF,
		Copymode:     gnfqueue.NfQnlCopyPacket,
		Flags:        gnfqueue.NfQaCfgFlagFailOpen,
		WriteTimeout: 15 * time.Millisecond,
	}
	nf, err := gnfqueue.Open(&config)
	if err != nil {
		return err
	}
	defer nf.Close()

	h := chalnfq{
		token:   token,
		domain:  domain,
		keyAuth: keyAuth,
		queue:   nf,
	}

	// error here would mean we couldn't capture packet, notthing to act about
	ignoreerr := func(err error) int {
		log.Print(err)
		return 0
	}

	// Register function to listen on nflqueue queue
	err = nf.RegisterWithErrorFunc(w.context, h.handlePacket, ignoreerr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Block till the context expires
	<-w.context.Done()
	return nil
}

func (w *HTTPProvider) Present(domain, token, keyAuth string) error {
	// test if OS is linux, otherwise no point running this nfqueue is linux thing
	if runtime.GOOS != "linux" {
		return fmt.Errorf("[%s] http-nfq provider is only for linux", domain)
	}
	// test if there is a webserver on port requested
	con, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%s", w.port), time.Second)
	if err != nil {
		return fmt.Errorf("[%s] http-nfq needs a webserver watching on requested, port %s", domain, w.port)
	} else {
		con.Close()
	}
	// if there is residuel firewall rule from old run remove it, ignore error
	setFirewallRule(false, w.port)

	// try set actuall firewall rule needed
	err = setFirewallRule(true, w.port)
	if err != nil {
		return fmt.Errorf("[nfqueue] fail to set firewal rule, error : %s", err.Error())
	}
	w.context, w.cancel = context.WithCancel(context.Background())
	go w.serve(domain, token, keyAuth)
	return nil
}

// CleanUp removes the firewall rule created for the challenge.
// solve should removed it already but just do be safe:
func (w *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	setFirewallRule(false, w.port)
	// tell nfqueue to shut down
	w.cancel()
	return nil
}
