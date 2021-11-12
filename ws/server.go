package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"log"
	"net"
	"net/http"
	"sync"
)

type MsgHandler struct {
	NewDTO func() interface{}
	OnMsg  func(dto interface{}) error
}

type Server struct {
	lock sync.RWMutex

	mcAddr string
	mcConn net.Conn
	mcHandlers map[int][]MsgHandler
}

func SetUp(r *web.Router, cfg util.WSConfig) *Server {
	s := &Server{
		mcAddr: fmt.Sprintf("%s:%d", cfg.MCIP, cfg.MCPort),
		mcHandlers: make(map[int][]MsgHandler),
	}

	r.NewRoute(
		"websocket/mc",
		nil,
		map[string]web.Handler{
			http.MethodGet: s.handleMC,
		},
	)

	r.NewRoute(
		"websocket/player",
		nil,
		map[string]web.Handler{
			http.MethodGet: s.handlePlayer,
		},
	)

	return s
}

func (s * Server) AddMCHandler(msgType int, onMsg MsgHandler) {
	if _, ok := s.mcHandlers[msgType]; !ok {
		s.mcHandlers[msgType] = make([]MsgHandler, 0, 1)
	}
	s.mcHandlers[msgType] = append(s.mcHandlers[msgType], onMsg)
}

func (s *Server) SendToMC(msg interface{}) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return err
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.mcConn == nil {
		return errors.New("MC connection has been closed")
	}

	if err := wsutil.WriteClientText(s.mcConn, jsonMsg); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Server) handleMC(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	conn, err := s.upgradeConn(res, req)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Established WebSocket connection with %s\n", req.RemoteAddr)
	go s.readMCMsgs(conn)
}

func (s *Server) upgradeConn(res http.ResponseWriter, req *http.Request) (net.Conn, error) {
	conn, _, _, err := ws.UpgradeHTTP(req, res)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if s.mcConn != nil {
		closeWS(conn, 4000, "Cannot open more than 1 connection")
		return nil, errors.New("cannot open more than 1 connection")
	}

	if conn.RemoteAddr().String() != s.mcAddr {
		closeWS(conn, 4001, "Connection from this address is not allowed")
		return nil, errors.New(fmt.Sprintf("connection from this address is not allowed '%s'", conn.RemoteAddr().String()))
	}

	s.mcConn = conn
	return conn, nil
}

func (s *Server) readMCMsgs(conn net.Conn) {
	defer func() {
		conn.Close()

		s.lock.Lock()
		s.mcConn = nil
		s.lock.Unlock()
	}()

	for {
		msg, err := wsutil.ReadClientText(conn)
		if err != nil {
			closeWS(conn, 4500, err.Error())
			return
		}

		var msgType msgType
		if err := json.Unmarshal(msg, &msgType); err != nil {
			closeWS(conn, 4002, err.Error())
			return
		}

		handlers, ok := s.mcHandlers[msgType.T]
		if !ok {
			closeWS(conn, 4004, fmt.Sprintf("There is no handler for the message type %d", msgType.T))
			return
		}

		for _, h := range handlers {
			go s.onMCMsg(msg, h)
		}
	}
}

func (s *Server) onMCMsg(msg []byte, h MsgHandler) {
	DTO := h.NewDTO()

	if DTO != nil {
		if err := json.Unmarshal(msg, DTO); err != nil {
			log.Println(err)
			return
		}
	}

	if err := h.OnMsg(DTO); err != nil {
		log.Println(err)
		return
	}
}

func (s *Server) handlePlayer(res http.ResponseWriter, req *http.Request, _ web.PathVars) {

}

func (s *Server) sendToPlayer(msg interface{}) error {
	return nil
}

func closeWS(conn net.Conn, code ws.StatusCode, reason string) {
	if err := ws.WriteFrame(conn, ws.NewCloseFrame(ws.NewCloseFrameBody(code, reason))); err != nil {
		log.Println(err)
	}
	conn.Close()
}
