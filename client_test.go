package grout

import (
	"encoding/binary"
	"errors"
	"net"
	"testing"
)

// multiHandler wraps a hello handshake followed by a custom handler for the
// next request. It supports both single-response and stream-response patterns.
func multiHandler(serverVersion string, next func(conn net.Conn, id, msgT, payloadLen uint32, payload []byte)) func(net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()

		// Handle hello.
		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		helloPayload, err := readPayload(conn, helloReqSize)
		if err != nil {
			return
		}
		_ = helloPayload

		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], serverVersion)
		if err := writeResponse(conn, id, 0, resp); err != nil {
			return
		}

		// Handle next request.
		id2, msgT2, pLen2, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		payload2, err := readPayload(conn, pLen2)
		if err != nil {
			return
		}
		next(conn, id2, msgT2, pLen2, payload2)
	}
}

func connectToMock(t *testing.T, handler func(net.Conn)) *Client {
	t.Helper()
	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestDHCPStart(t *testing.T) {
	var gotIfaceID uint16
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		if msgT != msgTypeDHCPStart {
			t.Errorf("msgType = 0x%x, want 0x%x", msgT, msgTypeDHCPStart)
		}
		gotIfaceID = binary.LittleEndian.Uint16(payload[0:2])
		writeResponse(conn, id, 0, nil)
	}))

	err := client.DHCPStart(7)
	if err != nil {
		t.Fatalf("DHCPStart: %v", err)
	}
	if gotIfaceID != 7 {
		t.Errorf("ifaceID = %d, want 7", gotIfaceID)
	}
}

func TestDHCPStop(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.DHCPStop(3); err != nil {
		t.Fatalf("DHCPStop: %v", err)
	}
}

func TestDHCPList_Empty(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	statuses, err := client.DHCPList()
	if err != nil {
		t.Fatalf("DHCPList: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("got %d statuses, want 0", len(statuses))
	}
}

func TestDHCPList_WithEntries(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, dhcpStatusSize)
		binary.LittleEndian.PutUint16(entry[0:2], 5)
		entry[2] = byte(DHCPStateBound)
		copy(entry[4:8], []byte{10, 0, 0, 1})
		copy(entry[8:12], []byte{10, 0, 0, 50})
		binary.LittleEndian.PutUint32(entry[12:16], 3600)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	statuses, err := client.DHCPList()
	if err != nil {
		t.Fatalf("DHCPList: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("got %d statuses, want 1", len(statuses))
	}
	if statuses[0].IfaceID != 5 {
		t.Errorf("IfaceID = %d, want 5", statuses[0].IfaceID)
	}
}

func TestLogLevelList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, logEntrySize)
		copy(entry[0:logNameLen], "grout.main")
		binary.LittleEndian.PutUint32(entry[logNameLen:logNameLen+4], 7)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	entries, err := client.LogLevelList(false)
	if err != nil {
		t.Fatalf("LogLevelList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Name != "grout.main" {
		t.Errorf("Name = %q, want %q", entries[0].Name, "grout.main")
	}
}

func TestLogLevelSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.LogLevelSet("grout.*", 5); err != nil {
		t.Fatalf("LogLevelSet: %v", err)
	}
}

func TestLogPacketsSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.LogPacketsSet(true); err != nil {
		t.Fatalf("LogPacketsSet: %v", err)
	}
}

func TestStatsGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, statSize)
		copy(entry[0:statNameLen], "eth_rx")
		binary.LittleEndian.PutUint64(entry[statNameLen:statNameLen+8], 1)
		binary.LittleEndian.PutUint64(entry[statNameLen+8:statNameLen+16], 500)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	stats, err := client.StatsGet(StatsGetRequest{Flags: StatsFlagSW})
	if err != nil {
		t.Fatalf("StatsGet: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("got %d stats, want 1", len(stats))
	}
	if stats[0].Name != "eth_rx" {
		t.Errorf("Name = %q, want %q", stats[0].Name, "eth_rx")
	}
}

func TestStatsReset(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.StatsReset(); err != nil {
		t.Fatalf("StatsReset: %v", err)
	}
}

func TestGraphDump(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, []byte("digraph { a -> b }"))
	}))

	dump, err := client.GraphDump(GraphDumpRequest{Format: 0})
	if err != nil {
		t.Fatalf("GraphDump: %v", err)
	}
	if dump != "digraph { a -> b }" {
		t.Errorf("dump = %q", dump)
	}
}

func TestGraphConfigGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, graphConfSize)
		binary.LittleEndian.PutUint16(resp[0:2], 256)
		binary.LittleEndian.PutUint16(resp[2:4], 32)
		writeResponse(conn, id, 0, resp)
	}))

	conf, err := client.GraphConfigGet()
	if err != nil {
		t.Fatalf("GraphConfigGet: %v", err)
	}
	if conf.RxBurstMax != 256 {
		t.Errorf("RxBurstMax = %d, want 256", conf.RxBurstMax)
	}
	if conf.VectorMax != 32 {
		t.Errorf("VectorMax = %d, want 32", conf.VectorMax)
	}
}

func TestGraphConfigSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.GraphConfigSet(GraphConf{RxBurstMax: 128, VectorMax: 16}); err != nil {
		t.Fatalf("GraphConfigSet: %v", err)
	}
}

func TestPacketTraceSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.PacketTraceSet(5, true, false); err != nil {
		t.Fatalf("PacketTraceSet: %v", err)
	}
}

func TestPacketTraceClear(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.PacketTraceClear(); err != nil {
		t.Fatalf("PacketTraceClear: %v", err)
	}
}

func TestPacketTraceDump(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, []byte("trace data"))
	}))

	data, err := client.PacketTraceDump(10)
	if err != nil {
		t.Fatalf("PacketTraceDump: %v", err)
	}
	if string(data) != "trace data" {
		t.Errorf("data = %q", data)
	}
}

func TestEventSubscribe(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		if msgT != msgTypeEventSubscribe {
			t.Errorf("msgType = 0x%x, want EventSubscribe", msgT)
		}
		evType := binary.LittleEndian.Uint32(payload[0:4])
		if evType != EventAll {
			t.Errorf("evType = 0x%x, want EventAll", evType)
		}
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.EventSubscribe(EventAll, false); err != nil {
		t.Fatalf("EventSubscribe: %v", err)
	}
}

func TestEventUnsubscribe(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.EventUnsubscribe(); err != nil {
		t.Fatalf("EventUnsubscribe: %v", err)
	}
}

func TestAffinityRxQList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, rxQueueMapSize)
		binary.LittleEndian.PutUint16(entry[0:2], 1)
		binary.LittleEndian.PutUint16(entry[2:4], 0)
		binary.LittleEndian.PutUint16(entry[4:6], 3)
		binary.LittleEndian.PutUint16(entry[6:8], 1)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	maps, err := client.AffinityRxQList()
	if err != nil {
		t.Fatalf("AffinityRxQList: %v", err)
	}
	if len(maps) != 1 {
		t.Fatalf("got %d maps, want 1", len(maps))
	}
	if maps[0].CPUID != 3 {
		t.Errorf("CPUID = %d, want 3", maps[0].CPUID)
	}
}

func TestAffinityRxQSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.AffinityRxQSet(1, 0, 4); err != nil {
		t.Fatalf("AffinityRxQSet: %v", err)
	}
}

func TestAffinityCPUGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := encodeCPUAffinity(CPUSet{CPUs: []uint16{0}}, CPUSet{CPUs: []uint16{1, 2}})
		writeResponse(conn, id, 0, resp)
	}))

	aff, err := client.AffinityCPUGet()
	if err != nil {
		t.Fatalf("AffinityCPUGet: %v", err)
	}
	if len(aff.Control.CPUs) != 1 {
		t.Errorf("Control CPUs len = %d, want 1", len(aff.Control.CPUs))
	}
	if len(aff.Datapath.CPUs) != 2 {
		t.Errorf("Datapath CPUs len = %d, want 2", len(aff.Datapath.CPUs))
	}
}

func TestAffinityCPUSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.AffinityCPUSet(CPUSet{CPUs: []uint16{0}}, CPUSet{CPUs: []uint16{1, 2}}); err != nil {
		t.Fatalf("AffinityCPUSet: %v", err)
	}
}

func TestConntrackList_Empty(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	entries, err := client.ConntrackList()
	if err != nil {
		t.Fatalf("ConntrackList: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestConntrackFlush(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.ConntrackFlush(); err != nil {
		t.Fatalf("ConntrackFlush: %v", err)
	}
}

func TestConntrackConfigGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := encodeConntrackConfig(ConntrackConfig{MaxCount: 65536, TimeoutTCPEstablished: 7200})
		writeResponse(conn, id, 0, resp)
	}))

	cfg, err := client.ConntrackConfigGet()
	if err != nil {
		t.Fatalf("ConntrackConfigGet: %v", err)
	}
	if cfg.MaxCount != 65536 {
		t.Errorf("MaxCount = %d, want 65536", cfg.MaxCount)
	}
}

func TestConntrackConfigSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.ConntrackConfigSet(ConntrackConfig{MaxCount: 32768}); err != nil {
		t.Fatalf("ConntrackConfigSet: %v", err)
	}
}

func TestFDBAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.FDBAdd(FDBEntry{BridgeID: 1, MAC: MustParseEtherAddr("aa:bb:cc:dd:ee:ff")}, false)
	if err != nil {
		t.Fatalf("FDBAdd: %v", err)
	}
}

func TestFDBDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.FDBDel(1, MustParseEtherAddr("aa:bb:cc:dd:ee:ff"), 100, true); err != nil {
		t.Fatalf("FDBDel: %v", err)
	}
}

func TestFDBFlush(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.FDBFlush(FDBFlushRequest{BridgeID: 1}); err != nil {
		t.Fatalf("FDBFlush: %v", err)
	}
}

func TestFDBList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := encodeFDBEntry(FDBEntry{BridgeID: 1, MAC: MustParseEtherAddr("aa:bb:cc:dd:ee:ff"), Flags: FDBFlagStatic}, false)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	entries, err := client.FDBList(FDBListRequest{BridgeID: 1})
	if err != nil {
		t.Fatalf("FDBList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Flags != FDBFlagStatic {
		t.Errorf("Flags = %d, want Static", entries[0].Flags)
	}
}

func TestFDBConfigGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, fdbConfigSize)
		binary.LittleEndian.PutUint32(resp[0:4], 10000)
		binary.LittleEndian.PutUint32(resp[4:8], 42)
		writeResponse(conn, id, 0, resp)
	}))

	cfg, err := client.FDBConfigGet()
	if err != nil {
		t.Fatalf("FDBConfigGet: %v", err)
	}
	if cfg.MaxEntries != 10000 {
		t.Errorf("MaxEntries = %d, want 10000", cfg.MaxEntries)
	}
	if cfg.UsedEntries != 42 {
		t.Errorf("UsedEntries = %d, want 42", cfg.UsedEntries)
	}
}

func TestFDBConfigSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.FDBConfigSet(50000); err != nil {
		t.Fatalf("FDBConfigSet: %v", err)
	}
}

func TestFloodAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.FloodAdd(FloodEntry{Type: FloodTypeVTEP, VNI: 42, Addr: L3AddrFromIP4(IP4Addr{10, 0, 0, 1})}, false)
	if err != nil {
		t.Fatalf("FloodAdd: %v", err)
	}
}

func TestFloodDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.FloodDel(FloodEntry{Type: FloodTypeVTEP, VNI: 42}, true)
	if err != nil {
		t.Fatalf("FloodDel: %v", err)
	}
}

func TestFloodList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := encodeFloodEntry(FloodEntry{Type: FloodTypeVTEP, VNI: 99, Addr: L3AddrFromIP4(IP4Addr{10, 0, 0, 1})}, false)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	entries, err := client.FloodList(FloodTypeVTEP, 1)
	if err != nil {
		t.Fatalf("FloodList: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].VNI != 99 {
		t.Errorf("VNI = %d, want 99", entries[0].VNI)
	}
}

func TestDNAT44Add(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.DNAT44Add(DNAT44Policy{IfaceID: 1, Match: MustParseIP4("10.0.0.1"), Replace: MustParseIP4("192.168.1.1")}, false)
	if err != nil {
		t.Fatalf("DNAT44Add: %v", err)
	}
}

func TestDNAT44Del(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.DNAT44Del(1, MustParseIP4("10.0.0.1"), true); err != nil {
		t.Fatalf("DNAT44Del: %v", err)
	}
}

func TestDNAT44List(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	policies, err := client.DNAT44List(0)
	if err != nil {
		t.Fatalf("DNAT44List: %v", err)
	}
	if len(policies) != 0 {
		t.Errorf("got %d policies, want 0", len(policies))
	}
}

func TestSNAT44Add(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.SNAT44Add(SNAT44Policy{IfaceID: 1, Net: MustParseIP4Net("10.0.0.0/8"), Replace: MustParseIP4("1.2.3.4")}, false)
	if err != nil {
		t.Fatalf("SNAT44Add: %v", err)
	}
}

func TestSNAT44Del(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.SNAT44Del(SNAT44Policy{IfaceID: 1}, true); err != nil {
		t.Fatalf("SNAT44Del: %v", err)
	}
}

func TestSNAT44List(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	policies, err := client.SNAT44List()
	if err != nil {
		t.Fatalf("SNAT44List: %v", err)
	}
	if len(policies) != 0 {
		t.Errorf("got %d policies, want 0", len(policies))
	}
}

func TestSRv6TunSrcSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		if msgT != msgTypeSRv6TunSrcSet {
			t.Errorf("msgType = 0x%x, want SRv6TunSrcSet", msgT)
		}
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.SRv6TunSrcSet(MustParseIP6("fd00::1")); err != nil {
		t.Fatalf("SRv6TunSrcSet: %v", err)
	}
}

func TestSRv6TunSrcClear(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.SRv6TunSrcClear(); err != nil {
		t.Fatalf("SRv6TunSrcClear: %v", err)
	}
}

func TestSRv6TunSrcShow(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		addr := MustParseIP6("fd00::1")
		writeResponse(conn, id, 0, addr[:])
	}))

	addr, err := client.SRv6TunSrcShow()
	if err != nil {
		t.Fatalf("SRv6TunSrcShow: %v", err)
	}
	if addr.String() != "fd00::1" {
		t.Errorf("addr = %s, want fd00::1", addr)
	}
}

func TestSRv6TunSrcShow_NotFound(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	_, err := client.SRv6TunSrcShow()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestIP4AddrAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP4AddrAdd(IP4IfAddr{IfaceID: 1, Addr: MustParseIP4Net("10.0.0.1/24")}, false); err != nil {
		t.Fatalf("IP4AddrAdd: %v", err)
	}
}

func TestIP4AddrDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP4AddrDel(IP4IfAddr{IfaceID: 1, Addr: MustParseIP4Net("10.0.0.1/24")}, true); err != nil {
		t.Fatalf("IP4AddrDel: %v", err)
	}
}

func TestIP4AddrFlush(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP4AddrFlush(1); err != nil {
		t.Fatalf("IP4AddrFlush: %v", err)
	}
}

func TestIP4RouteAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.IP4RouteAdd(IP4RouteAddRequest{
		VRFID:  VRFDefaultID,
		Dest:   MustParseIP4Net("0.0.0.0/0"),
		NH:     MustParseIP4("10.0.0.1"),
		Origin: NHOriginStatic,
	})
	if err != nil {
		t.Fatalf("IP4RouteAdd: %v", err)
	}
}

func TestIP4RouteDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP4RouteDel(VRFDefaultID, MustParseIP4Net("10.0.0.0/24"), true); err != nil {
		t.Fatalf("IP4RouteDel: %v", err)
	}
}

func TestIP4FIBDefaultSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP4FIBDefaultSet(65536); err != nil {
		t.Fatalf("IP4FIBDefaultSet: %v", err)
	}
}

func TestIP6AddrAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6AddrAdd(IP6IfAddr{IfaceID: 1, Addr: MustParseIP6Net("fd00::1/64")}, false); err != nil {
		t.Fatalf("IP6AddrAdd: %v", err)
	}
}

func TestIP6AddrDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6AddrDel(IP6IfAddr{IfaceID: 1, Addr: MustParseIP6Net("fd00::1/64")}, true); err != nil {
		t.Fatalf("IP6AddrDel: %v", err)
	}
}

func TestIP6AddrFlush(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6AddrFlush(1); err != nil {
		t.Fatalf("IP6AddrFlush: %v", err)
	}
}

func TestIP6RouteAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	err := client.IP6RouteAdd(IP6RouteAddRequest{
		VRFID:  VRFDefaultID,
		Dest:   MustParseIP6Net("fd00::/64"),
		NH:     MustParseIP6("fd00::1"),
		Origin: NHOriginStatic,
	})
	if err != nil {
		t.Fatalf("IP6RouteAdd: %v", err)
	}
}

func TestIP6RouteDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6RouteDel(VRFDefaultID, MustParseIP6Net("fd00::/64"), true); err != nil {
		t.Fatalf("IP6RouteDel: %v", err)
	}
}

func TestIP6RASet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6RASet(RAConf{Enabled: true, IfaceID: 1, Interval: 600, Lifetime: 1800}); err != nil {
		t.Fatalf("IP6RASet: %v", err)
	}
}

func TestIP6RAClear(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6RAClear(1); err != nil {
		t.Fatalf("IP6RAClear: %v", err)
	}
}

func TestNHAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.NexthopAdd(Nexthop{Type: NHTypeBlackhole}); err != nil {
		t.Fatalf("NexthopAdd: %v", err)
	}
}

func TestNHDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.NexthopDel(42); err != nil {
		t.Fatalf("NexthopDel: %v", err)
	}
}

func TestNHConfigGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := encodeNHConfig(NHConfig{MaxCount: 1024})
		writeResponse(conn, id, 0, resp)
	}))

	cfg, err := client.NexthopConfigGet()
	if err != nil {
		t.Fatalf("NexthopConfigGet: %v", err)
	}
	if cfg.MaxCount != 1024 {
		t.Errorf("MaxCount = %d, want 1024", cfg.MaxCount)
	}
}

func TestNHConfigSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.NexthopConfigSet(NHConfig{MaxCount: 2048}); err != nil {
		t.Fatalf("NexthopConfigSet: %v", err)
	}
}

func TestRequestError(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0xfffffffe, nil) // -ENOENT (negated errno 2)
	}))

	err := client.DHCPStart(1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestInterfaceGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, ifaceBaseSize)
		binary.LittleEndian.PutUint16(resp[0:2], 5)
		resp[2] = byte(IfaceTypeVRF)
		binary.LittleEndian.PutUint16(resp[8:10], 1500)
		copy(resp[18:18+ifaceNameLen], "vrf0")
		writeResponse(conn, id, 0, resp)
	}))

	iface, err := client.InterfaceGet(5)
	if err != nil {
		t.Fatalf("InterfaceGet: %v", err)
	}
	if iface.ID != 5 {
		t.Errorf("ID = %d, want 5", iface.ID)
	}
	if iface.Name != "vrf0" {
		t.Errorf("Name = %q, want %q", iface.Name, "vrf0")
	}
}

func TestInterfaceSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.InterfaceSet(InterfaceSetRequest{IfaceID: 1, MTU: 9000, SetAttrs: IfaceSetMTU}); err != nil {
		t.Fatalf("InterfaceSet: %v", err)
	}
}

func TestInterfaceStatsGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, ifaceStatsSize)
		binary.LittleEndian.PutUint16(entry[0:2], 3)
		binary.LittleEndian.PutUint64(entry[8:16], 1000)
		binary.LittleEndian.PutUint64(entry[16:24], 64000)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	stats, err := client.InterfaceStatsGet()
	if err != nil {
		t.Fatalf("InterfaceStatsGet: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("got %d stats, want 1", len(stats))
	}
	if stats[0].RxPackets != 1000 {
		t.Errorf("RxPackets = %d, want 1000", stats[0].RxPackets)
	}
}

func TestInterfaceMACAdd(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.InterfaceMACAdd(1, MustParseEtherAddr("aa:bb:cc:dd:ee:ff")); err != nil {
		t.Fatalf("InterfaceMACAdd: %v", err)
	}
}

func TestInterfaceMACDel(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.InterfaceMACDel(1, MustParseEtherAddr("aa:bb:cc:dd:ee:ff")); err != nil {
		t.Fatalf("InterfaceMACDel: %v", err)
	}
}

func TestInterfaceMACList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, ifaceMACSize)
		binary.LittleEndian.PutUint16(entry[0:2], 1)
		binary.LittleEndian.PutUint16(entry[2:4], 1) // ref_count
		entry[4] = 1                                   // primary
		mac := MustParseEtherAddr("aa:bb:cc:dd:ee:ff")
		copy(entry[6:12], mac[:])
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	macs, err := client.InterfaceMACList(1)
	if err != nil {
		t.Fatalf("InterfaceMACList: %v", err)
	}
	if len(macs) != 1 {
		t.Fatalf("got %d MACs, want 1", len(macs))
	}
	if !macs[0].Primary {
		t.Error("expected primary MAC")
	}
}

func TestInterfaceMACSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.InterfaceMACSet(1, MustParseEtherAddr("aa:bb:cc:dd:ee:ff")); err != nil {
		t.Fatalf("InterfaceMACSet: %v", err)
	}
}

func TestIP4AddrList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := encodeIP4IfAddr(IP4IfAddr{IfaceID: 1, Addr: MustParseIP4Net("10.0.0.1/24")}, false)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	addrs, err := client.IP4AddrList(0, 1)
	if err != nil {
		t.Fatalf("IP4AddrList: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("got %d addrs, want 1", len(addrs))
	}
	if addrs[0].Addr.PrefixLen != 24 {
		t.Errorf("PrefixLen = %d, want 24", addrs[0].Addr.PrefixLen)
	}
}

func TestIP4RouteGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, nexthopBaseSize)
		resp[0] = byte(NHTypeBlackhole)
		binary.LittleEndian.PutUint32(resp[4:8], 42)
		writeResponse(conn, id, 0, resp)
	}))

	nh, err := client.IP4RouteGet(VRFDefaultID, MustParseIP4("10.0.0.0"))
	if err != nil {
		t.Fatalf("IP4RouteGet: %v", err)
	}
	if nh.Type != NHTypeBlackhole {
		t.Errorf("Type = %d, want Blackhole", nh.Type)
	}
}

func TestIP4RouteList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	routes, err := client.IP4RouteList(VRFDefaultID, 100)
	if err != nil {
		t.Fatalf("IP4RouteList: %v", err)
	}
	if len(routes) != 0 {
		t.Errorf("got %d routes, want 0", len(routes))
	}
}

func TestIP4FIBInfoList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, fib4InfoSize)
		binary.LittleEndian.PutUint16(entry[0:2], 1)
		binary.LittleEndian.PutUint32(entry[4:8], 65536)
		binary.LittleEndian.PutUint32(entry[8:12], 100)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	infos, err := client.IP4FIBInfoList(VRFDefaultID)
	if err != nil {
		t.Fatalf("IP4FIBInfoList: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("got %d infos, want 1", len(infos))
	}
	if infos[0].MaxRoutes != 65536 {
		t.Errorf("MaxRoutes = %d, want 65536", infos[0].MaxRoutes)
	}
}

func TestIP6AddrList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := encodeIP6IfAddr(IP6IfAddr{IfaceID: 1, Addr: MustParseIP6Net("fd00::1/64")}, false)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	addrs, err := client.IP6AddrList(0, 1)
	if err != nil {
		t.Fatalf("IP6AddrList: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("got %d addrs, want 1", len(addrs))
	}
}

func TestIP6RouteGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, nexthopBaseSize)
		resp[0] = byte(NHTypeBlackhole)
		binary.LittleEndian.PutUint32(resp[4:8], 42)
		writeResponse(conn, id, 0, resp)
	}))

	nh, err := client.IP6RouteGet(VRFDefaultID, MustParseIP6("fd00::"))
	if err != nil {
		t.Fatalf("IP6RouteGet: %v", err)
	}
	if nh.Type != NHTypeBlackhole {
		t.Errorf("Type = %d, want Blackhole", nh.Type)
	}
}

func TestIP6RouteList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	routes, err := client.IP6RouteList(VRFDefaultID, 100)
	if err != nil {
		t.Fatalf("IP6RouteList: %v", err)
	}
	if len(routes) != 0 {
		t.Errorf("got %d routes, want 0", len(routes))
	}
}

func TestIP6FIBDefaultSet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeResponse(conn, id, 0, nil)
	}))

	if err := client.IP6FIBDefaultSet(65536); err != nil {
		t.Fatalf("IP6FIBDefaultSet: %v", err)
	}
}

func TestIP6FIBInfoList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		writeStreamEnd(conn, id)
	}))

	infos, err := client.IP6FIBInfoList(VRFDefaultID)
	if err != nil {
		t.Fatalf("IP6FIBInfoList: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("got %d infos, want 0", len(infos))
	}
}

func TestIP6RAShow(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := encodeRAConf(RAConf{Enabled: true, IfaceID: 1, Interval: 600, Lifetime: 1800})
		writeResponse(conn, id, 0, resp)
		writeStreamEnd(conn, id)
	}))

	confs, err := client.IP6RAShow(1)
	if err != nil {
		t.Fatalf("IP6RAShow: %v", err)
	}
	if len(confs) != 1 {
		t.Fatalf("got %d confs, want 1", len(confs))
	}
	if !confs[0].Enabled {
		t.Error("expected Enabled")
	}
	if confs[0].Interval != 600 {
		t.Errorf("Interval = %d, want 600", confs[0].Interval)
	}
}

func TestNexthopGet(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		resp := make([]byte, nexthopBaseSize)
		resp[0] = byte(NHTypeBlackhole)
		binary.LittleEndian.PutUint32(resp[4:8], 42)
		writeResponse(conn, id, 0, resp)
	}))

	nh, err := client.NexthopGet(42)
	if err != nil {
		t.Fatalf("NexthopGet: %v", err)
	}
	if nh.Type != NHTypeBlackhole {
		t.Errorf("Type = %d, want Blackhole", nh.Type)
	}
}

func TestNexthopList(t *testing.T) {
	client := connectToMock(t, multiHandler("grout-test", func(conn net.Conn, id, msgT, pLen uint32, payload []byte) {
		entry := make([]byte, nexthopBaseSize)
		entry[0] = byte(NHTypeBlackhole)
		binary.LittleEndian.PutUint32(entry[4:8], 1)
		writeResponse(conn, id, 0, entry)
		writeStreamEnd(conn, id)
	}))

	nhs, err := client.NexthopList()
	if err != nil {
		t.Fatalf("NexthopList: %v", err)
	}
	if len(nhs) != 1 {
		t.Fatalf("got %d nexthops, want 1", len(nhs))
	}
}

func TestInterfaceGetByName_Found(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()

		// Handle hello.
		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, helloReqSize)
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], "grout-test")
		writeResponse(conn, id, 0, resp)

		// Handle InterfaceList (stream).
		id2, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, 0)

		entry := make([]byte, ifaceBaseSize)
		binary.LittleEndian.PutUint16(entry[0:2], 3)
		entry[2] = byte(IfaceTypePort)
		copy(entry[18:18+ifaceNameLen], "p0")
		writeResponse(conn, id2, 0, entry)
		writeStreamEnd(conn, id2)
	}

	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	iface, err := client.InterfaceGetByName("p0")
	if err != nil {
		t.Fatalf("InterfaceGetByName: %v", err)
	}
	if iface.ID != 3 {
		t.Errorf("ID = %d, want 3", iface.ID)
	}
}

func TestIP4Ping(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()

		// Hello.
		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, helloReqSize)
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], "grout-test")
		writeResponse(conn, id, 0, resp)

		// ICMPSend.
		id2, _, pLen2, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, pLen2)
		writeResponse(conn, id2, 0, nil)

		// ICMPRecv.
		id3, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, 0)
		icmp := make([]byte, 20)
		icmp[2] = 64 // ttl
		binary.LittleEndian.PutUint16(icmp[6:8], 1)
		copy(icmp[8:12], []byte{10, 0, 0, 1})
		binary.LittleEndian.PutUint64(icmp[12:20], 500000)
		writeResponse(conn, id3, 0, icmp)
	}

	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	resp, err := client.IP4Ping(IP4PingRequest{
		VRFID: VRFDefaultID,
		Addr:  MustParseIP4("10.0.0.1"),
		TTL:   64,
	})
	if err != nil {
		t.Fatalf("IP4Ping: %v", err)
	}
	if resp.TTL != 64 {
		t.Errorf("TTL = %d, want 64", resp.TTL)
	}
}

func TestIP6Ping(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()

		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, helloReqSize)
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], "grout-test")
		writeResponse(conn, id, 0, resp)

		id2, _, pLen2, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, pLen2)
		writeResponse(conn, id2, 0, nil)

		id3, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, 0)
		icmp := make([]byte, 32)
		icmp[0] = 129 // echo reply
		icmp[2] = 64
		src := MustParseIP6("fe80::1")
		copy(icmp[8:24], src[:])
		binary.LittleEndian.PutUint64(icmp[24:32], 1000000)
		writeResponse(conn, id3, 0, icmp)
	}

	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	resp, err := client.IP6Ping(IP6PingRequest{
		VRFID: VRFDefaultID,
		Addr:  MustParseIP6("fe80::1"),
		TTL:   64,
	})
	if err != nil {
		t.Fatalf("IP6Ping: %v", err)
	}
	if resp.TTL != 64 {
		t.Errorf("TTL = %d, want 64", resp.TTL)
	}
}

func TestEventRecv(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()

		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, helloReqSize)
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], "grout-test")
		writeResponse(conn, id, 0, resp)

		// EventSubscribe.
		id2, _, pLen2, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, pLen2)
		writeResponse(conn, id2, 0, nil)

		// Send an event (event header: ev_type(4) + payload_len(8)).
		var evHdr [eventHeaderSize]byte
		binary.LittleEndian.PutUint32(evHdr[0:4], 42)
		binary.LittleEndian.PutUint64(evHdr[4:12], 4)
		conn.Write(evHdr[:])
		conn.Write([]byte("test"))
	}

	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	if err := client.EventSubscribe(EventAll, false); err != nil {
		t.Fatalf("EventSubscribe: %v", err)
	}

	ev, err := client.EventRecv()
	if err != nil {
		t.Fatalf("EventRecv: %v", err)
	}
	if ev.EvType != 42 {
		t.Errorf("EvType = %d, want 42", ev.EvType)
	}
	if string(ev.Payload) != "test" {
		t.Errorf("Payload = %q, want %q", ev.Payload, "test")
	}
}

func TestEncodeIfaceAdd_VLAN(t *testing.T) {
	req := InterfaceAddRequest{
		Type:     IfaceTypeVLAN,
		Name:     "vlan100",
		ParentID: 5,
		VLANID:   100,
	}
	buf := encodeIfaceAdd(req)
	if len(buf) < ifaceBaseSize+4 {
		t.Fatalf("buf too short: %d", len(buf))
	}
	parentID := binary.LittleEndian.Uint16(buf[ifaceBaseSize : ifaceBaseSize+2])
	vlanID := binary.LittleEndian.Uint16(buf[ifaceBaseSize+2 : ifaceBaseSize+4])
	if parentID != 5 {
		t.Errorf("ParentID = %d, want 5", parentID)
	}
	if vlanID != 100 {
		t.Errorf("VLANID = %d, want 100", vlanID)
	}
}

func TestEncodeIfaceAdd_IPIP(t *testing.T) {
	req := InterfaceAddRequest{
		Type:   IfaceTypeIPIP,
		Name:   "ipip0",
		Local:  MustParseIP4("10.0.0.1"),
		Remote: MustParseIP4("10.0.0.2"),
	}
	buf := encodeIfaceAdd(req)
	if len(buf) < ifaceBaseSize+8 {
		t.Fatalf("buf too short: %d", len(buf))
	}
}

func TestEncodeIfaceAdd_VXLAN(t *testing.T) {
	req := InterfaceAddRequest{
		Type:       IfaceTypeVXLAN,
		Name:       "vxlan0",
		VNI:        42,
		EncapVRFID: 1,
		DstPort:    4789,
		VXLANLocal: L3AddrFromIP4(IP4Addr{10, 0, 0, 1}),
	}
	buf := encodeIfaceAdd(req)
	if len(buf) < ifaceBaseSize+8+l3AddrWireSize {
		t.Fatalf("buf too short: %d", len(buf))
	}
	vni := binary.LittleEndian.Uint32(buf[ifaceBaseSize : ifaceBaseSize+4])
	if vni != 42 {
		t.Errorf("VNI = %d, want 42", vni)
	}
}

func TestDecodeIface_VRF(t *testing.T) {
	data := make([]byte, ifaceBaseSize+22)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	data[2] = byte(IfaceTypeVRF)
	copy(data[18:18+ifaceNameLen], "vrf0")
	binary.LittleEndian.PutUint32(data[ifaceBaseSize:ifaceBaseSize+4], 65536)

	iface, err := decodeIface(data)
	if err != nil {
		t.Fatalf("decodeIface: %v", err)
	}
	if iface.VRF == nil {
		t.Fatal("VRF info is nil")
	}
	if iface.VRF.IPv4FIB.MaxRoutes != 65536 {
		t.Errorf("IPv4 MaxRoutes = %d, want 65536", iface.VRF.IPv4FIB.MaxRoutes)
	}
}

func TestDecodeIface_VLAN(t *testing.T) {
	data := make([]byte, ifaceBaseSize+10)
	binary.LittleEndian.PutUint16(data[0:2], 2)
	data[2] = byte(IfaceTypeVLAN)
	copy(data[18:18+ifaceNameLen], "vlan100")
	binary.LittleEndian.PutUint16(data[ifaceBaseSize:ifaceBaseSize+2], 5)
	binary.LittleEndian.PutUint16(data[ifaceBaseSize+2:ifaceBaseSize+4], 100)

	iface, err := decodeIface(data)
	if err != nil {
		t.Fatalf("decodeIface: %v", err)
	}
	if iface.VLAN == nil {
		t.Fatal("VLAN info is nil")
	}
	if iface.VLAN.VLANID != 100 {
		t.Errorf("VLANID = %d, want 100", iface.VLAN.VLANID)
	}
}

func TestDecodeIface_IPIP(t *testing.T) {
	data := make([]byte, ifaceBaseSize+8)
	data[2] = byte(IfaceTypeIPIP)
	copy(data[18:18+ifaceNameLen], "ipip0")
	copy(data[ifaceBaseSize:ifaceBaseSize+4], []byte{10, 0, 0, 1})
	copy(data[ifaceBaseSize+4:ifaceBaseSize+8], []byte{10, 0, 0, 2})

	iface, err := decodeIface(data)
	if err != nil {
		t.Fatalf("decodeIface: %v", err)
	}
	if iface.IPIP == nil {
		t.Fatal("IPIP info is nil")
	}
	if iface.IPIP.Local.String() != "10.0.0.1" {
		t.Errorf("Local = %s, want 10.0.0.1", iface.IPIP.Local)
	}
}

func TestDecodeIface_VXLAN(t *testing.T) {
	data := make([]byte, ifaceBaseSize+8+l3AddrWireSize+6)
	data[2] = byte(IfaceTypeVXLAN)
	copy(data[18:18+ifaceNameLen], "vxlan0")
	binary.LittleEndian.PutUint32(data[ifaceBaseSize:ifaceBaseSize+4], 42)
	binary.LittleEndian.PutUint16(data[ifaceBaseSize+6:ifaceBaseSize+8], 4789)

	iface, err := decodeIface(data)
	if err != nil {
		t.Fatalf("decodeIface: %v", err)
	}
	if iface.VXLAN == nil {
		t.Fatal("VXLAN info is nil")
	}
	if iface.VXLAN.VNI != 42 {
		t.Errorf("VNI = %d, want 42", iface.VXLAN.VNI)
	}
	if iface.VXLAN.DstPort != 4789 {
		t.Errorf("DstPort = %d, want 4789", iface.VXLAN.DstPort)
	}
}

func TestInterfaceGetByName_NotFound(t *testing.T) {
	handler := func(conn net.Conn) {
		defer conn.Close()

		id, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, helloReqSize)
		resp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(resp[0:4], CurrentAPIVersion)
		copy(resp[4:], "grout-test")
		writeResponse(conn, id, 0, resp)

		id2, _, _, err := readRequestHeader(conn)
		if err != nil {
			return
		}
		readPayload(conn, 0)
		writeStreamEnd(conn, id2)
	}

	srv := newMockServer(t, handler)
	t.Cleanup(srv.close)
	client, err := Connect(WithSocketPath(srv.sockPath))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	_, err = client.InterfaceGetByName("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
