package grout_test_r

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	grout "github.com/DPDK/grout/go"
	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const groutImage = "quay.io/grout/grout:0.16.0"

// TAP interface type value (not yet exported by the library).
const ifaceTypeTAP = grout.IfaceType(8)

var sharedClient *grout.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	sockDir, err := os.MkdirTemp("", "grout-e2e-*")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(sockDir) }()

	req := testcontainers.ContainerRequest{
		Image: groutImage,
		Cmd:   []string{"/usr/bin/grout", "-t", "-s", "/run/grout/grout.sock", "-m", "0666"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = append(hc.Binds, sockDir+":/run/grout")
		},
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("start grout container: %v", err)
	}
	defer func() {
		if err := ctr.Terminate(ctx); err != nil {
			log.Printf("terminate grout container: %v", err)
		}
	}()

	sockPath := filepath.Join(sockDir, "grout.sock")
	if err := waitForSocket(sockPath, 60*time.Second); err != nil {
		logs, _ := ctr.Logs(ctx)
		if logs != nil {
			buf := make([]byte, 4096)
			n, _ := logs.Read(buf)
			log.Printf("grout container logs:\n%s", buf[:n])
		}
		log.Fatalf("grout socket not ready: %v", err)
	}

	sharedClient, err = grout.Connect(sockPath)
	if err != nil {
		log.Fatalf("connect to grout: %v", err)
	}
	defer func() { _ = sharedClient.Close() }()

	os.Exit(m.Run())
}

func waitForSocket(sockPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", sockPath)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("socket %s not ready after %s", sockPath, timeout)
}

func collectIfaces(t *testing.T, c *grout.Client) []*grout.Iface {
	t.Helper()
	var result []*grout.Iface
	for iface, err := range c.IfaceList(&grout.IfaceListReq{}) {
		require.NoError(t, err)
		result = append(result, iface)
	}
	return result
}

func findIface(ifaces []*grout.Iface, id uint16) *grout.Iface {
	for _, iface := range ifaces {
		if iface.ID == id {
			return iface
		}
	}
	return nil
}

func TestE2E_InterfaceLifecycle(t *testing.T) {
	resp, err := sharedClient.IfaceAdd(&grout.IfaceAddReq{
		Iface: grout.Iface{
			Name: "tap-e2e0",
			Type: ifaceTypeTAP,
		},
	})
	require.NoError(t, err)
	ifaceID := resp.IfaceID
	assert.NotZero(t, ifaceID)

	ifaces := collectIfaces(t, sharedClient)
	found := findIface(ifaces, ifaceID)
	require.NotNil(t, found, "interface %d should be present after add", ifaceID)
	assert.Equal(t, "tap-e2e0", found.Name)

	err = sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: ifaceID})
	require.NoError(t, err)

	ifaces = collectIfaces(t, sharedClient)
	assert.Nil(t, findIface(ifaces, ifaceID), "interface %d should be absent after del", ifaceID)
}

func TestE2E_AddressLifecycle(t *testing.T) {
	resp, err := sharedClient.IfaceAdd(&grout.IfaceAddReq{
		Iface: grout.Iface{
			Name: "tap-e2e1",
			Type: ifaceTypeTAP,
		},
	})
	require.NoError(t, err)
	ifaceID := resp.IfaceID
	t.Cleanup(func() {
		_ = sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: ifaceID})
	})

	addr := grout.IP4Ifaddr{
		IfaceID: ifaceID,
		Addr:    grout.IPv4NetFrom(netip.MustParsePrefix("10.200.0.2/24")),
	}

	err = sharedClient.IP4AddrAdd(&grout.IP4AddrAddReq{Addr: addr, ExistOk: 0})
	require.NoError(t, err)

	err = sharedClient.IP4AddrDel(&grout.IP4AddrDelReq{Addr: addr, MissingOk: 0})
	require.NoError(t, err)
}

func TestE2E_RouteLifecycle(t *testing.T) {
	resp, err := sharedClient.IfaceAdd(&grout.IfaceAddReq{
		Iface: grout.Iface{
			Name: "tap-e2e2",
			Type: ifaceTypeTAP,
		},
	})
	require.NoError(t, err)
	ifaceID := resp.IfaceID
	t.Cleanup(func() {
		_ = sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: ifaceID})
	})

	ifaceAddr := grout.IP4Ifaddr{
		IfaceID: ifaceID,
		Addr:    grout.IPv4NetFrom(netip.MustParsePrefix("10.201.0.1/24")),
	}
	require.NoError(t, sharedClient.IP4AddrAdd(&grout.IP4AddrAddReq{Addr: ifaceAddr, ExistOk: 0}))
	t.Cleanup(func() {
		_ = sharedClient.IP4AddrDel(&grout.IP4AddrDelReq{Addr: ifaceAddr, MissingOk: 1})
	})

	dest := grout.IPv4NetFrom(netip.MustParsePrefix("10.202.0.0/24"))
	nh := netip.MustParseAddr("10.201.0.254")

	require.NoError(t, sharedClient.IP4RouteAdd(&grout.IP4RouteAddReq{
		Dest: dest,
		NH:   nh,
	}))

	require.NoError(t, sharedClient.IP4RouteDel(&grout.IP4RouteDelReq{
		VRFID:     0,
		Dest:      dest,
		MissingOk: 0,
	}))
}

func TestE2E_ErrorCases(t *testing.T) {
	t.Run("delete non-existent interface", func(t *testing.T) {
		err := sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: 9999})
		assert.Error(t, err)
	})

	t.Run("delete non-existent address with missingOK", func(t *testing.T) {
		resp, err := sharedClient.IfaceAdd(&grout.IfaceAddReq{
			Iface: grout.Iface{
				Name: "tap-e2e3",
				Type: ifaceTypeTAP,
			},
		})
		require.NoError(t, err)
		ifaceID := resp.IfaceID
		t.Cleanup(func() {
			_ = sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: ifaceID})
		})

		err = sharedClient.IP4AddrDel(&grout.IP4AddrDelReq{
			Addr: grout.IP4Ifaddr{
				IfaceID: ifaceID,
				Addr:    grout.IPv4NetFrom(netip.MustParsePrefix("10.99.99.99/24")),
			},
			MissingOk: 1,
		})
		assert.NoError(t, err)
	})

	t.Run("delete non-existent route with missingOK", func(t *testing.T) {
		err := sharedClient.IP4RouteDel(&grout.IP4RouteDelReq{
			VRFID:     0,
			Dest:      grout.IPv4NetFrom(netip.MustParsePrefix("10.88.88.0/24")),
			MissingOk: 1,
		})
		assert.NoError(t, err)
	})
}

func TestE2E_FullCRUDSequence(t *testing.T) {
	resp, err := sharedClient.IfaceAdd(&grout.IfaceAddReq{
		Iface: grout.Iface{
			Name: "tap-e2e4",
			Type: ifaceTypeTAP,
		},
	})
	require.NoError(t, err)
	ifaceID := resp.IfaceID

	addr := grout.IP4Ifaddr{
		IfaceID: ifaceID,
		Addr:    grout.IPv4NetFrom(netip.MustParsePrefix("10.203.0.1/24")),
	}
	require.NoError(t, sharedClient.IP4AddrAdd(&grout.IP4AddrAddReq{Addr: addr, ExistOk: 0}))

	dest := grout.IPv4NetFrom(netip.MustParsePrefix("10.204.0.0/24"))
	require.NoError(t, sharedClient.IP4RouteAdd(&grout.IP4RouteAddReq{
		Dest: dest,
		NH:   netip.MustParseAddr("10.203.0.254"),
	}))

	ifaces := collectIfaces(t, sharedClient)
	assert.NotNil(t, findIface(ifaces, ifaceID))

	require.NoError(t, sharedClient.IP4RouteDel(&grout.IP4RouteDelReq{VRFID: 0, Dest: dest, MissingOk: 0}))
	require.NoError(t, sharedClient.IP4AddrDel(&grout.IP4AddrDelReq{Addr: addr, MissingOk: 0}))
	require.NoError(t, sharedClient.IfaceDel(&grout.IfaceDelReq{IfaceID: ifaceID}))

	ifaces = collectIfaces(t, sharedClient)
	assert.Nil(t, findIface(ifaces, ifaceID))
}
