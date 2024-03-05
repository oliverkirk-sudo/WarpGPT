package tools

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strings"
)

const fieldSeparator = ":"

type SSEClient struct {
	EventSource io.ReadCloser
	logger      *log.Logger
}

type Event struct {
	ID    string
	Event string
	Data  string
	Retry string
}

func NewSSEClient(eventSource io.ReadCloser) *SSEClient {
	return &SSEClient{
		EventSource: eventSource,
		logger:      log.New(log.Writer(), "SSEClient: ", log.LstdFlags),
	}
}

func (c *SSEClient) Read() <-chan Event {
	events := make(chan Event)
	go func() {
		defer close(events)
		reader := bufio.NewReaderSize(c.EventSource, 128*1024)
		var data bytes.Buffer

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					c.logger.Printf("Message Done!") //当前会话结束时，提示结束而非报错。
					break
				}
				c.logger.Printf("Error reading from event source: %v", err)
				break
			}

			data.Write(line)

			if bytes.HasSuffix(data.Bytes(), []byte("\n\n")) || bytes.HasSuffix(data.Bytes(), []byte("\r\n\r\n")) {
				event := c.parseEvent(data.String())
				if event.Data != "" {
					events <- event
				}
				data.Reset()
			}
		}
	}()

	return events
}

func (c *SSEClient) parseEvent(data string) Event {
	event := Event{
		ID:    "",
		Event: "message",
		Data:  "",
		Retry: "",
	}
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, fieldSeparator) {
			continue
		}

		parts := strings.SplitN(line, fieldSeparator, 2)
		field := parts[0]
		var value string
		if len(parts) == 2 {
			value = strings.TrimPrefix(parts[1], " ")
		}

		switch field {
		case "id":
			event.ID = value
		case "event":
			event.Event = value
		case "data":
			event.Data += value + "\n"
		case "retry":
			event.Retry = value
		}
	}

	if strings.HasSuffix(event.Data, "\n") {
		event.Data = strings.TrimSuffix(event.Data, "\n")
	}

	return event
}

func (c *SSEClient) Close() error {
	err := c.EventSource.Close()
	if err != nil {
		return err
	}
	return nil
}
