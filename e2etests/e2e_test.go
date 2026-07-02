//go:build e2e

package grout_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	grout "github.com/zeeke/grout-go"
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

	sharedClient, err = grout.Connect(grout.WithSocketPath(sockPath))
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

func findIface(ifaces []grout.Iface, id uint16) *grout.Iface {
	for i := range ifaces {
		if ifaces[i].ID == id {
			return &ifaces[i]
		}
	}
	return nil
}

func TestE2E_InterfaceLifecycle(t *testing.T) {
	ifaceID, err := sharedClient.InterfaceAdd(grout.InterfaceAddRequest{
		Name: "tap-e2e0",
		Type: ifaceTypeTAP,
	})
	require.NoError(t, err)
	assert.NotZero(t, ifaceID)

	ifaces, err := sharedClient.InterfaceList()
	require.NoError(t, err)
	found := findIface(ifaces, ifaceID)
	require.NotNil(t, found, "interface %d should be present after add", ifaceID)
	assert.Equal(t, "tap-e2e0", found.Name)

	err = sharedClient.InterfaceDel(ifaceID)
	require.NoError(t, err)

	ifaces, err = sharedClient.InterfaceList()
	require.NoError(t, err)
	assert.Nil(t, findIface(ifaces, ifaceID), "interface %d should be absent after del", ifaceID)
}

func TestE2E_AddressLifecycle(t *testing.T) {
	ifaceID, err := sharedClient.InterfaceAdd(grout.InterfaceAddRequest{
		Name: "tap-e2e1",
		Type: ifaceTypeTAP,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sharedClient.InterfaceDel(ifaceID)
	})

	addr := grout.IP4IfAddr{
		IfaceID: ifaceID,
		Addr:    grout.MustParseIP4Net("10.200.0.2/24"),
	}

	err = sharedClient.IP4AddrAdd(addr, false)
	require.NoError(t, err)

	err = sharedClient.IP4AddrDel(addr, false)
	require.NoError(t, err)
}

func TestE2E_RouteLifecycle(t *testing.T) {
	ifaceID, err := sharedClient.InterfaceAdd(grout.InterfaceAddRequest{
		Name: "tap-e2e2",
		Type: ifaceTypeTAP,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sharedClient.InterfaceDel(ifaceID)
	})

	ifaceAddr := grout.IP4IfAddr{
		IfaceID: ifaceID,
		Addr:    grout.MustParseIP4Net("10.201.0.1/24"),
	}
	require.NoError(t, sharedClient.IP4AddrAdd(ifaceAddr, false))
	t.Cleanup(func() {
		_ = sharedClient.IP4AddrDel(ifaceAddr, true)
	})

	dest := grout.MustParseIP4Net("10.202.0.0/24")
	nh := grout.MustParseIP4("10.201.0.254")

	require.NoError(t, sharedClient.IP4RouteAdd(grout.IP4RouteAddRequest{
		Dest: dest,
		NH:   nh,
	}))

	require.NoError(t, sharedClient.IP4RouteDel(0, dest, false))
}

func TestE2E_ErrorCases(t *testing.T) {
	t.Run("delete non-existent interface", func(t *testing.T) {
		err := sharedClient.InterfaceDel(9999)
		assert.Error(t, err)
	})

	t.Run("delete non-existent address with missingOK", func(t *testing.T) {
		ifaceID, err := sharedClient.InterfaceAdd(grout.InterfaceAddRequest{
			Name: "tap-e2e3",
			Type: ifaceTypeTAP,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = sharedClient.InterfaceDel(ifaceID)
		})

		err = sharedClient.IP4AddrDel(grout.IP4IfAddr{
			IfaceID: ifaceID,
			Addr:    grout.MustParseIP4Net("10.99.99.99/24"),
		}, true)
		assert.NoError(t, err)
	})

	t.Run("delete non-existent route with missingOK", func(t *testing.T) {
		err := sharedClient.IP4RouteDel(0, grout.MustParseIP4Net("10.88.88.0/24"), true)
		assert.NoError(t, err)
	})
}

func TestE2E_FullCRUDSequence(t *testing.T) {
	ifaceID, err := sharedClient.InterfaceAdd(grout.InterfaceAddRequest{
		Name: "tap-e2e4",
		Type: ifaceTypeTAP,
	})
	require.NoError(t, err)

	addr := grout.IP4IfAddr{
		IfaceID: ifaceID,
		Addr:    grout.MustParseIP4Net("10.203.0.1/24"),
	}
	require.NoError(t, sharedClient.IP4AddrAdd(addr, false))

	dest := grout.MustParseIP4Net("10.204.0.0/24")
	require.NoError(t, sharedClient.IP4RouteAdd(grout.IP4RouteAddRequest{
		Dest: dest,
		NH:   grout.MustParseIP4("10.203.0.254"),
	}))

	ifaces, err := sharedClient.InterfaceList()
	require.NoError(t, err)
	assert.NotNil(t, findIface(ifaces, ifaceID))

	require.NoError(t, sharedClient.IP4RouteDel(0, dest, false))
	require.NoError(t, sharedClient.IP4AddrDel(addr, false))
	require.NoError(t, sharedClient.InterfaceDel(ifaceID))

	ifaces, err = sharedClient.InterfaceList()
	require.NoError(t, err)
	assert.Nil(t, findIface(ifaces, ifaceID))
}
