package grout

import "encoding/binary"

// Event represents a received event from the grout daemon.
type Event struct {
	EvType     uint32
	PayloadLen uint64
	Payload    []byte
}

// EventSubscribe subscribes to events of the given type.
// Use EventAll to subscribe to all event types.
// If suppressSelf is true, events triggered by this client are not delivered.
func (c *Client) EventSubscribe(evType uint32, suppressSelf bool) error {
	if err := c.checkVersion(msgTypeEventSubscribe); err != nil {
		return err
	}
	payload := make([]byte, 8) // ev_type(4) + suppress_self(4)
	binary.LittleEndian.PutUint32(payload[0:4], evType)
	if suppressSelf {
		binary.LittleEndian.PutUint32(payload[4:8], 1)
	}
	_, err := c.request(msgTypeEventSubscribe, payload)
	return err
}

// EventUnsubscribe unsubscribes from all events.
func (c *Client) EventUnsubscribe() error {
	if err := c.checkVersion(msgTypeEventUnsubscribe); err != nil {
		return err
	}
	_, err := c.requestNoPayload(msgTypeEventUnsubscribe)
	return err
}

// EventRecv waits for and returns the next event.
// The connection must have an active event subscription.
func (c *Client) EventRecv() (*Event, error) {
	hdr, payload, err := c.readEvent()
	if err != nil {
		return nil, err
	}
	return &Event{
		EvType:     hdr.EvType,
		PayloadLen: hdr.PayloadLen,
		Payload:    payload,
	}, nil
}
