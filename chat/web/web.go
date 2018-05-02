package web

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-web"
	"github.com/peonone/parrot"
	"github.com/peonone/parrot/chat"
	"github.com/peonone/parrot/chat/proto"
)

var maxIdle = time.Second * 20

const (
	idleCheckInterval = time.Second / 2
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type baseCmdHandler struct {
	oum  *onlineUsersManager
	stat *stat
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
	closed          bool
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

func (u *onlineUser) close() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.closed {
		u.closed = true
		close(u.pushCh)
	}
}

// Init initializes required resources and register handlers
func Init(service web.Service) func() {
	ouManager := newOnlineUsersManager()
	rpcClient := client.NewClient(client.RequestTimeout(time.Second * 120))
	authHandler := newAuthHandler(rpcClient)

	var stat *stat
	statLogF, err := os.OpenFile("logs/chat_stat.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("failed to open stat log file: %s", err)
	} else {
		stat = newStat(statLogF)
		go stat.runStat()
	}

	stateClient := proto.StateServiceClient(chat.SrvServiceName, rpcClient)
	bch := &baseCmdHandler{
		oum:  ouManager,
		stat: stat,
	}
	cmdHandlers := []commandHandler{
		newPrivateMessageHandler(rpcClient, bch),
		newWorldShoutHandler(rpcClient, bch),
	}
	amqpConn, amqpChan, err := parrot.MakeAMQPClient()
	if err != nil {
		panic(err)
	}
	err = listenPushMsgs(amqpChan, cmdHandlers)
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
		if amqpConn != nil {
			amqpConn.Close()
		}
		if amqpChan != nil {
			amqpChan.Close()
		}

		wg := new(sync.WaitGroup)
		for uid := range ouManager.getAllOnlineUsers() {
			wg.Add(1)
			go func(uid string) {
				req := &proto.UserOfflineReq{
					Uid: uid,
				}
				stateClient.Offline(context.Background(), req)
				wg.Done()
			}(uid)
		}
		wg.Wait()
	}
}

func (u *onlineUser) sendToClient() {
	for msg := range u.pushCh {
		u.userClient.WriteJSON(msg)
	}
}

func (u *onlineUser) checkIdle() {
	for !u.closed {
		u.mu.Lock()
		lastReqTime := u.lastRequestTime
		u.mu.Unlock()
		if time.Since(lastReqTime) >= maxIdle {
			u.pushCh <- idleToLongResp
			u.userClient.Close()
			u.close()
			break
		}
		time.Sleep(idleCheckInterval)
	}
}

func (u *onlineUser) serve(
	authHandler *authHandler, cmdHandlers []commandHandler,
	stateService proto.StateService, ouManager *onlineUsersManager) error {
	defer func() {
		req := &proto.UserOfflineReq{
			Uid: u.uid,
		}
		stateService.Offline(context.Background(), req)
		ouManager.offline(u)
	}()
	go u.sendToClient()
	go u.checkIdle()
	err := authHandler.doAuth(u)

	if err == io.EOF {
		return nil
	} else if u.closed {
		return nil
	} else if err != nil {
		return err
	}
	req := &proto.UserOnlineReq{
		Uid:     u.uid,
		WebNode: web.DefaultId,
	}
	_, err = stateService.Online(context.Background(), req)
	if err != nil {
		return err
	}
	ouManager.online(u)

	for {
		err = processWsRequest(u, cmdHandlers)
		if err == io.EOF {
			return nil
		} else if u.closed {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func processWsRequest(u *onlineUser, cmdHandlers []commandHandler) error {
	// the WS client may send various kinds of request
	var req map[string]interface{}
	err := u.userClient.ReadJSON(&req)
	if err != nil {
		return err
	}
	u.mu.Lock()
	u.lastRequestTime = time.Now()
	u.mu.Unlock()
	cmd := req["command"]
	cmdStr, ok := cmd.(string)
	if cmd == nil || !ok {
		u.pushCh <- invalidCmdResp
		return nil
	}

	handled := false
	for _, h := range cmdHandlers {
		if h.canHandle(cmdStr) {
			if err = h.validate(req); err != nil {
				u.pushCh <- &genericResp{
					Success: false,
					ErrMsg:  err.Error(),
				}
			} else {
				h.handle(u, req)
			}
			handled = true
			break
		}
	}
	if !handled {
		u.pushCh <- invalidCmdResp
	}
	return nil
}
