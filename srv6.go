package grout

// SRv6TunSrcSet sets the SRv6 tunnel source address.
func (c *Client) SRv6TunSrcSet(addr IP6Addr) error {
	if err := c.checkVersion(msgTypeSRv6TunSrcSet); err != nil {
		return err
	}
	_, err := c.request(msgTypeSRv6TunSrcSet, addr[:])
	return err
}

// SRv6TunSrcClear clears the SRv6 tunnel source address.
func (c *Client) SRv6TunSrcClear() error {
	if err := c.checkVersion(msgTypeSRv6TunSrcClear); err != nil {
		return err
	}
	_, err := c.requestNoPayload(msgTypeSRv6TunSrcClear)
	return err
}

// SRv6TunSrcShow returns the current SRv6 tunnel source address.
func (c *Client) SRv6TunSrcShow() (*IP6Addr, error) {
	if err := c.checkVersion(msgTypeSRv6TunSrcShow); err != nil {
		return nil, err
	}
	resp, err := c.requestNoPayload(msgTypeSRv6TunSrcShow)
	if err != nil {
		return nil, err
	}
	if len(resp) < 16 {
		return nil, ErrNotFound
	}
	var addr IP6Addr
	copy(addr[:], resp[0:16])
	return &addr, nil
}
