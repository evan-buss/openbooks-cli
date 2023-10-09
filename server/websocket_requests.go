package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/evan-buss/openbooks/irc"
	"strings"
	"time"

	"github.com/evan-buss/openbooks/core"
	"github.com/evan-buss/openbooks/util"
)

// RequestHandler defines a generic handle() method that is called when a specific request type is made
type RequestHandler interface {
	handle(c *Client)
}

// messageRouter is used to parse the incoming request and respond appropriately
func (server *server) routeMessage(message Request, c *Client) {
	var obj interface{}

	switch message.MessageType {
	case CONNECT:
		obj = new(ConnectionRequest)
	case SEARCH:
		obj = new(SearchRequest)
	case DOWNLOAD:
		obj = new(DownloadRequest)
	}

	err := json.Unmarshal(message.Payload, &obj)
	if err != nil {
		server.log.Printf("Invalid request payload. %s.\n", err.Error())
		c.send <- StatusResponse{
			MessageType:      STATUS,
			NotificationType: DANGER,
			Title:            "Unknown request payload.",
		}
	}

	switch message.MessageType {
	case CONNECT:
		c.startIrcConnection(obj.(*ConnectionRequest), server)
	case SEARCH:
		c.sendSearchRequest(obj.(*SearchRequest), server)
	case DOWNLOAD:
		c.sendDownloadRequest(obj.(*DownloadRequest))
	default:
		server.log.Println("Unknown request type received.")
	}
}

// handle ConnectionRequests and either connect to the server or do nothing
func (c *Client) startIrcConnection(request *ConnectionRequest, server *server) {
	err := core.Join(c.irc, request.Address, request.Channel, request.EnableTLS)
	if err != nil {
		c.log.Println(err)
		if errors.Is(err, irc.ErrTLSHandshake) {
			c.send <- newErrorResponse(err.Error())
		} else {
			c.send <- newErrorResponse("Unable to connect to IRC server.")
		}
		return
	}

	address := strings.Split(request.Address, ":")[0]

	c.log.Printf("Connected to %s #%s as %s.\n", address, request.Channel, c.irc.Username)

	handler := server.NewIrcEventHandler(c)

	if server.config.Log {
		logger, _, err := util.CreateLogFile(c.irc.Username, server.config.DownloadDir)
		if err != nil {
			server.log.Println(err)
		}
		handler[core.Message] = func(text string) { logger.Println(text) }
	}

	go core.StartReader(c.ctx, c.irc, handler)

	c.send <- ConnectionResponse{
		StatusResponse: StatusResponse{
			MessageType:      CONNECT,
			NotificationType: SUCCESS,
			Title:            fmt.Sprintf("Connection established to %s #%s", address, request.Channel),
			Detail:           fmt.Sprintf("IRC username %s", c.irc.Username),
		},
		Name: c.irc.Username,
	}
}

// handle SearchRequests and send the query to the book server
func (c *Client) sendSearchRequest(s *SearchRequest, server *server) {
	server.lastSearchMutex.Lock()
	defer server.lastSearchMutex.Unlock()

	nextAvailableSearch := server.lastSearch.Add(server.config.SearchTimeout)

	if time.Now().Before(nextAvailableSearch) {
		remainingSeconds := time.Until(nextAvailableSearch).Seconds()
		c.send <- newRateLimitResponse(remainingSeconds)

		return
	}

	core.SearchBook(c.irc, server.config.SearchBot, s.Query)
	server.lastSearch = time.Now()

	c.send <- newStatusResponse(NOTIFY, "Search request sent.")
}

// handle DownloadRequests by sending the request to the book server
func (c *Client) sendDownloadRequest(d *DownloadRequest) {
	core.DownloadBook(c.irc, d.Book)
	c.send <- newStatusResponse(NOTIFY, "Download request received.")
}
