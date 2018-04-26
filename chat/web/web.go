package web

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	"github.com/peonone/parrot"
	"github.com/peonone/parrot/chat/proto"
	"github.com/peonone/parrot/chat/srv"
)

// Name is the chat web service name
const Name = "go.micro.web.chat"

const (
	maxIdle           = time.Second * 5
	idleCheckInterval = time.Second * 1
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type baseCmdHandler struct {
	oum *onlineUsersManager
}

type commandHandler interface {
	canHandle(cmd string) bool
	validate(req map[string]interface{}) error
	handle(client *onlineUser, req map[string]interface{})

	canHandlePush(cmd string) bool
	handlePush(msg *proto.PushMsg)
}

type userClient interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

type onlineUser struct {
	mu              sync.Mutex
	userClient      userClient
	uid             string
	pushCh          chan interface{}
	lastRequestTime time.Time
}

type genericResp struct {
	Success bool   `json:"success"`
	ErrMsg  string `json:"errMsg"`
}

var (
	invalidCmdResp = &genericResp{
		Success: false,
		ErrMsg:  "invalid command",
	}
	idleToLongResp = &genericResp{
		Success: false,
		ErrMsg:  "idle too long",
	}
)

var ouManager *onlineUsersManager

// Init initializes required resources and register handlers
func Init(service web.Service) func() {
	ouManager := newOnlineUsersManager()
	rpcClient := client.NewClient(client.RequestTimeout(time.Second * 120))
	authHandler := newAuthHandler(rpcClient)
	stateClient := proto.StateServiceClient(srv.Name, rpcClient)
	bch := &baseCmdHandler{
		oum: ouManager,
	}
	cmdHandlers := []commandHandler{
		newPrivateMessageHandler(rpcClient, bch),
	}
	amqpClient, err := parrot.MakeAMQPClient()
	if err != nil {
		panic(err)
	}
	err = listenPushMsgs(amqpClient, cmdHandlers)
	if err != nil {
		panic(err)
	}

	// Serve static html/js
	service.Handle("/", http.FileServer(http.Dir("chat/web/static")))
	// Handle websocket connection
	service.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade request to websocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal("Upgrade: ", err)
			return
		}
		defer conn.Close()
		onlineUser := onlineUser{
			userClient:      conn,
			pushCh:          make(chan interface{}),
			lastRequestTime: time.Now(),
		}
		err = onlineUser.serve(authHandler, cmdHandlers, stateClient, ouManager)
		if err != nil {
			log.Printf("WS connection closed with error %s", err)
		} else {
			log.Println("WS connection closed")
		}
	})
	return func() {
		ouManager.clear()
	}
}

func (c *onlineUser) sendToClient() {
	for msg := range c.pushCh {
		c.userClient.WriteJSON(msg)
	}
}

func (c *onlineUser) checkIdle() {
	for {
		c.mu.Lock()
		lastReqTime := c.lastRequestTime
		c.mu.Unlock()
		if time.Since(lastReqTime) >= maxIdle {
			c.userClient.WriteJSON(idleToLongResp)
			c.userClient.Close()
			break
		}
		time.Sleep(idleCheckInterval)
	}
}
func (c *onlineUser) serve(
	authHandler *authHandler, cmdHandlers []commandHandler,
	stateService proto.StateService, ouManager *onlineUsersManager) error {
	defer func() {
		req := &proto.UserOfflineReq{
			Uid: c.uid,
		}
		stateService.Offline(context.Background(), req)
	}()
	defer ouManager.offline(c)
	go c.sendToClient()
	go c.checkIdle()
	err := authHandler.doAuth(c)

	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	req := &proto.UserOnlineReq{
		Uid:     c.uid,
		WebNode: web.DefaultId,
	}
	_, err = stateService.Online(context.Background(), req)
	if err != nil {
		return err
	}
	ouManager.online(c)

	for {
		err = processWsRequest(c, cmdHandlers)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func processWsRequest(c *onlineUser, cmdHandlers []commandHandler) error {
	// the WS client may send various kinds of request
	req := make(map[string]interface{})
	err := c.userClient.ReadJSON(&req)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.lastRequestTime = time.Now()
	c.mu.Unlock()
	cmd := req["command"]
	cmdStr, ok := cmd.(string)
	if cmd == nil || !ok {
		c.pushCh <- invalidCmdResp
		return nil
	}

	handled := false
	for _, h := range cmdHandlers {
		if h.canHandle(cmdStr) {
			if err = h.validate(req); err != nil {
				c.pushCh <- &genericResp{
					Success: false,
					ErrMsg:  err.Error(),
				}
			} else {
				h.handle(c, req)
			}
			handled = true
			break
		}
	}
	if !handled {
		c.pushCh <- invalidCmdResp
	}
	return nil
}
