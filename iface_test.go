package grout

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestIfaceType_String(t *testing.T) {
	tests := []struct {
		t    IfaceType
		want string
	}{
		{IfaceTypePort, "port"},
		{IfaceTypeVRF, "vrf"},
		{IfaceTypeVLAN, "vlan"},
		{IfaceTypeIPIP, "ipip"},
		{IfaceTypeBond, "bond"},
		{IfaceTypeBridge, "bridge"},
		{IfaceTypeVXLAN, "vxlan"},
		{IfaceTypeUndef, "undef"},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("IfaceType(%d).String() = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestDecodeIface_Port(t *testing.T) {
	buf := make([]byte, ifaceBaseSize+14)
	binary.LittleEndian.PutUint16(buf[0:2], 42)   // ID
	buf[2] = byte(IfaceTypePort)                    // Type
	buf[3] = byte(IfaceModeVRF)                     // Mode
	binary.LittleEndian.PutUint16(buf[4:6], uint16(IfaceFlagUp)) // Flags
	binary.LittleEndian.PutUint16(buf[8:10], 1500)  // MTU
	binary.LittleEndian.PutUint16(buf[10:12], VRFDefaultID)
	copy(buf[18:18+ifaceNameLen], "p0")

	// Port info: n_rxq(2) + n_txq(2) + rxq_size(2) + txq_size(2) + mac(6)
	info := buf[ifaceBaseSize:]
	binary.LittleEndian.PutUint16(info[0:2], 4)   // n_rxq
	binary.LittleEndian.PutUint16(info[2:4], 4)   // n_txq
	binary.LittleEndian.PutUint16(info[4:6], 2048) // rxq_size
	binary.LittleEndian.PutUint16(info[6:8], 2048) // txq_size
	info[8] = 0xaa
	info[9] = 0xbb
	info[10] = 0xcc
	info[11] = 0xdd
	info[12] = 0xee
	info[13] = 0xff

	iface, err := decodeIface(buf)
	if err != nil {
		t.Fatalf("decodeIface error: %v", err)
	}

	if iface.ID != 42 {
		t.Errorf("ID = %d, want 42", iface.ID)
	}
	if iface.Type != IfaceTypePort {
		t.Errorf("Type = %d, want Port", iface.Type)
	}
	if iface.Name != "p0" {
		t.Errorf("Name = %q, want %q", iface.Name, "p0")
	}
	if iface.MTU != 1500 {
		t.Errorf("MTU = %d, want 1500", iface.MTU)
	}
	if iface.Port == nil {
		t.Fatal("Port info is nil")
	}
	if iface.Port.NRxQ != 4 {
		t.Errorf("NRxQ = %d, want 4", iface.Port.NRxQ)
	}
	if iface.Port.MAC.String() != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("MAC = %s, want aa:bb:cc:dd:ee:ff", iface.Port.MAC)
	}
}

func TestDecodeIface_TooShort(t *testing.T) {
	_, err := decodeIface(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestDecodeIfaceStats(t *testing.T) {
	buf := make([]byte, ifaceStatsSize)
	binary.LittleEndian.PutUint16(buf[0:2], 5)       // iface_id
	binary.LittleEndian.PutUint64(buf[8:16], 1000)    // rx_packets
	binary.LittleEndian.PutUint64(buf[16:24], 50000)  // rx_bytes
	binary.LittleEndian.PutUint64(buf[32:40], 800)    // tx_packets

	stats, err := decodeIfaceStats(buf)
	if err != nil {
		t.Fatalf("decodeIfaceStats error: %v", err)
	}
	if stats.IfaceID != 5 {
		t.Errorf("IfaceID = %d, want 5", stats.IfaceID)
	}
	if stats.RxPackets != 1000 {
		t.Errorf("RxPackets = %d, want 1000", stats.RxPackets)
	}
	if stats.RxBytes != 50000 {
		t.Errorf("RxBytes = %d, want 50000", stats.RxBytes)
	}
	if stats.TxPackets != 800 {
		t.Errorf("TxPackets = %d, want 800", stats.TxPackets)
	}
}

func TestDecodeIfaceMAC(t *testing.T) {
	buf := make([]byte, ifaceMACSize)
	binary.LittleEndian.PutUint16(buf[0:2], 3) // iface_id
	binary.LittleEndian.PutUint16(buf[2:4], 1) // ref_count
	buf[4] = 1                                   // primary
	copy(buf[6:12], []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66})

	m, err := decodeIfaceMAC(buf)
	if err != nil {
		t.Fatalf("decodeIfaceMAC error: %v", err)
	}
	if m.IfaceID != 3 {
		t.Errorf("IfaceID = %d, want 3", m.IfaceID)
	}
	if !m.Primary {
		t.Error("expected Primary = true")
	}
	if m.MAC.String() != "11:22:33:44:55:66" {
		t.Errorf("MAC = %s, want 11:22:33:44:55:66", m.MAC)
	}
}

func TestInterfaceAdd_Integration(t *testing.T) {
	s := newMockServer(t, func(conn net.Conn) {
		defer conn.Close()

		// Handle hello.
		id, _, _, _ := readRequestHeader(conn)
		readPayload(conn, helloReqSize)
		helloResp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(helloResp[0:4], CurrentAPIVersion)
		copy(helloResp[4:], "mock")
		writeResponse(conn, id, 0, helloResp)

		// Handle InterfaceAdd.
		id, _, pLen, _ := readRequestHeader(conn)
		readPayload(conn, pLen)

		// Build iface response.
		resp := make([]byte, ifaceBaseSize)
		binary.LittleEndian.PutUint16(resp[0:2], 7)  // assigned ID
		resp[2] = byte(IfaceTypePort)
		binary.LittleEndian.PutUint16(resp[8:10], 1500) // MTU
		copy(resp[18:18+ifaceNameLen], "p0")
		writeResponse(conn, id, 0, resp)
	})
	defer s.close()

	client, err := Connect(WithSocketPath(s.sockPath))
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	ifaceID, err := client.InterfaceAdd(InterfaceAddRequest{
		Type: IfaceTypePort,
		Name: "p0",
		MTU:  1500,
	})
	if err != nil {
		t.Fatalf("InterfaceAdd error: %v", err)
	}
	if ifaceID != 7 {
		t.Errorf("ifaceID = %d, want 7", ifaceID)
	}
}

func TestInterfaceList_WithFilter(t *testing.T) {
	s := newMockServer(t, func(conn net.Conn) {
		defer conn.Close()

		// Handle hello.
		id, _, _, _ := readRequestHeader(conn)
		readPayload(conn, helloReqSize)
		helloResp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(helloResp[0:4], CurrentAPIVersion)
		copy(helloResp[4:], "mock")
		writeResponse(conn, id, 0, helloResp)

		// Handle InterfaceList.
		id, _, pLen, _ := readRequestHeader(conn)
		readPayload(conn, pLen)

		// Return 2 interfaces.
		for i := 0; i < 2; i++ {
			resp := make([]byte, ifaceBaseSize)
			binary.LittleEndian.PutUint16(resp[0:2], uint16(i+1))
			resp[2] = byte(IfaceTypePort)
			binary.LittleEndian.PutUint16(resp[8:10], 1500)
			copy(resp[18:18+ifaceNameLen], "p"+string(rune('0'+i)))
			writeResponse(conn, id, 0, resp)
		}
		writeStreamEnd(conn, id)
	})
	defer s.close()

	client, err := Connect(WithSocketPath(s.sockPath))
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	ifaces, err := client.InterfaceList(WithIfaceType(IfaceTypePort))
	if err != nil {
		t.Fatalf("InterfaceList error: %v", err)
	}
	if len(ifaces) != 2 {
		t.Errorf("got %d interfaces, want 2", len(ifaces))
	}
}

func TestInterfaceDel_Integration(t *testing.T) {
	s := newMockServer(t, func(conn net.Conn) {
		defer conn.Close()

		// Handle hello.
		id, _, _, _ := readRequestHeader(conn)
		readPayload(conn, helloReqSize)
		helloResp := make([]byte, helloReqSize)
		binary.LittleEndian.PutUint32(helloResp[0:4], CurrentAPIVersion)
		copy(helloResp[4:], "mock")
		writeResponse(conn, id, 0, helloResp)

		// Handle InterfaceDel.
		id, _, pLen, _ := readRequestHeader(conn)
		readPayload(conn, pLen)
		writeResponse(conn, id, 0, nil)
	})
	defer s.close()

	client, err := Connect(WithSocketPath(s.sockPath))
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	if err := client.InterfaceDel(42); err != nil {
		t.Fatalf("InterfaceDel error: %v", err)
	}
}
