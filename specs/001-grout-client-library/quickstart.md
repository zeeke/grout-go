# Quickstart: grout-go

## Installation

```bash
go get github.com/zeeke/grout-go
```

## Connect to a Grout Instance

```go
package main

import (
    "fmt"
    "log"

    grout "github.com/zeeke/grout-go"
)

func main() {
    client, err := grout.ConnectDefault()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    fmt.Printf("Connected to grout %s (API v%d)\n",
        client.ServerVersion(), client.APIVersion())
}
```

## Create a Port Interface

```go
id, err := client.InterfaceAdd(grout.InterfaceAddRequest{
    Type:    grout.IfaceTypePort,
    Name:    "p0",
    DevArgs: "0000:8a:00.0",
    NRxQ:    1,
    RxQSize: 2048,
})
if err != nil {
    log.Fatal(err)
}

// Bring the interface up
err = client.InterfaceSet(grout.InterfaceSetRequest{
    IfaceID: id,
    Flags:   grout.IfaceFlagUp,
    SetFlags: grout.IfaceSetFlags,
})
```

## Add an IP Address

```go
err = client.IP4AddrAdd(grout.IP4IfAddr{
    IfaceID: id,
    Addr:    grout.MustParseIP4Net("172.16.0.2/24"),
}, false)
```

## Add a Route

```go
err = client.IP4RouteAdd(grout.IP4RouteAddRequest{
    VRFID:  grout.VRFDefaultID,
    Dest:   grout.MustParseIP4Net("0.0.0.0/0"),
    NH:     grout.MustParseIP4("172.16.1.183"),
    Origin: grout.NHOriginStatic,
})
```

## List Interfaces

```go
ifaces, err := client.InterfaceList(grout.IfaceTypePort)
if err != nil {
    log.Fatal(err)
}
for _, iface := range ifaces {
    fmt.Printf("%-6s id=%d mtu=%d flags=%v\n",
        iface.Name, iface.ID, iface.MTU, iface.Flags)
}
```

## Ping

```go
resp, err := client.IP4Ping(grout.IP4PingRequest{
    Addr:   grout.MustParseIP4("172.16.0.1"),
    VRFID:  grout.VRFDefaultID,
    SeqNum: 1,
    TTL:    64,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Reply from %s: ttl=%d time=%dns\n",
    resp.SrcAddr, resp.TTL, resp.ResponseTime)
```

## Subscribe to Events

```go
err = client.EventSubscribe(grout.EventAll, false)
if err != nil {
    log.Fatal(err)
}

for {
    event, err := client.EventRecv()
    if err != nil {
        break
    }
    fmt.Printf("Event: type=%d payload_len=%d\n",
        event.EvType, event.PayloadLen)
}
```

## NAT Configuration

```go
// Add static DNAT44 rule
err = client.DNAT44Add(grout.DNAT44Policy{
    IfaceID: id,
    Match:   grout.MustParseIP4("10.0.0.1"),
    Replace: grout.MustParseIP4("192.168.1.100"),
}, false)

// List conntrack entries
entries, err := client.ConntrackList()
for _, entry := range entries {
    fmt.Printf("conntrack: %s -> %s state=%v\n",
        entry.FwdFlow.Src, entry.FwdFlow.Dst, entry.State)
}
```

## Error Handling

```go
client, err := grout.Connect("/tmp/grout.sock")
if err != nil {
    // Check specific error types
    var grErr *grout.GrError
    if errors.As(err, &grErr) {
        fmt.Printf("Grout error: errno=%d\n", grErr.Errno)
    }
    if errors.Is(err, grout.ErrVersionMismatch) {
        fmt.Println("Incompatible API version")
    }
    if errors.Is(err, grout.ErrConnection) {
        fmt.Println("Cannot connect to daemon")
    }
}

// Version-gated features
err = client.SRv6TunSrcSet(addr)
if errors.Is(err, grout.ErrUnsupported) {
    fmt.Println("SRv6 not supported in this grout version")
}
```

## Running Tests Against a Container

```bash
# Start grout in test mode (no DPDK hardware needed)
podman run -d --name grout-test --privileged \
    -v /tmp/grout-test:/run \
    quay.io/grout/grout:0.16.0 \
    grout -t -s /run/grout.sock

# Run integration tests
go test -tags integration -race ./... \
    -grout-sock=/tmp/grout-test/grout.sock

# Clean up
podman rm -f grout-test
```
