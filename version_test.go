package grout

import (
	"errors"
	"testing"
)

func TestCheckVersion_Supported(t *testing.T) {
	c := &Client{apiVersion: 3}
	if err := c.checkVersion(msgTypeIfaceList); err != nil {
		t.Errorf("checkVersion returned error for supported message: %v", err)
	}
}

func TestCheckVersion_Unsupported(t *testing.T) {
	c := &Client{apiVersion: 2}
	err := c.checkVersion(msgTypeIfaceList)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Errorf("error = %v, want ErrUnsupported", err)
	}
}

func TestCheckVersion_UnknownMessage(t *testing.T) {
	c := &Client{apiVersion: 1}
	if err := c.checkVersion(0xdeadbeef); err != nil {
		t.Errorf("checkVersion returned error for unknown message: %v", err)
	}
}
