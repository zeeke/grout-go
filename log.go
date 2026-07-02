package grout

import (
	"encoding/binary"
	"fmt"
)

// LogEntry represents a log module and its level.
type LogEntry struct {
	Name  string
	Level uint32
}

const (
	logNameLen   = 64
	logEntrySize = logNameLen + 4 // name(64) + level(4)
)

// LogLevelList returns configured log levels.
func (c *Client) LogLevelList(showAll bool) ([]LogEntry, error) {
	if err := c.checkVersion(msgTypeLogLevelList); err != nil {
		return nil, err
	}
	payload := make([]byte, 1)
	if showAll {
		payload[0] = 1
	}

	results, err := c.requestStream(msgTypeLogLevelList, payload)
	if err != nil {
		return nil, err
	}

	entries := make([]LogEntry, 0, len(results))
	for _, data := range results {
		e, err := decodeLogEntry(data)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// LogLevelSet sets the log level for modules matching a pattern.
func (c *Client) LogLevelSet(pattern string, level uint32) error {
	if err := c.checkVersion(msgTypeLogLevelSet); err != nil {
		return err
	}
	buf := make([]byte, logNameLen+4)
	copy(buf[0:logNameLen], pattern)
	binary.LittleEndian.PutUint32(buf[logNameLen:logNameLen+4], level)
	_, err := c.request(msgTypeLogLevelSet, buf)
	return err
}

// LogPacketsSet enables or disables packet logging globally.
func (c *Client) LogPacketsSet(enabled bool) error {
	if err := c.checkVersion(msgTypeLogPacketsSet); err != nil {
		return err
	}
	payload := make([]byte, 1)
	if enabled {
		payload[0] = 1
	}
	_, err := c.request(msgTypeLogPacketsSet, payload)
	return err
}

func decodeLogEntry(data []byte) (LogEntry, error) {
	if len(data) < logEntrySize {
		return LogEntry{}, fmt.Errorf("log entry data too short: %d bytes", len(data))
	}
	return LogEntry{
		Name:  cstring(data[0:logNameLen]),
		Level: binary.LittleEndian.Uint32(data[logNameLen : logNameLen+4]),
	}, nil
}
