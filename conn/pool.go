package conn

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/maksymiliank/arrival-mc-backend/util"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type GameServerConn struct {
	conn net.Conn
	out  chan []byte
	lock sync.Mutex
}

type Pool interface {
	Broker() Broker
	SendGameMsg(msg []byte)
}

type PoolS struct {
	gameAllowedAddr string
	game            *GameServerConn
	broker          Broker
}

func SetUp(r *web.Router, cfg util.WSConfig) Pool {
	broker := NewBroker()
	gameAllowedAddr := cfg.IP + ":" + strconv.Itoa(cfg.Port)
	pool := &PoolS{
		gameAllowedAddr,
		&GameServerConn{nil, make(chan []byte, 100), sync.Mutex{}},
		broker}

	r.NewRoute(
		"/conn",
		nil,
		map[string]web.Handler{
			http.MethodGet: pool.onConn,
		},
	)

	return pool
}

func (p *PoolS) Broker() Broker {
	return p.broker
}

func (p *PoolS) SendGameMsg(msg []byte) {
	p.game.out <- msg
}

func (p *PoolS) onConn(res http.ResponseWriter, req *http.Request, _ web.PathVars) {
	if req.RemoteAddr != p.gameAllowedAddr {
		web.Forbidden(res)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(req, res)
	if err != nil {
		log.Println("upgrading HTTP failed: ", err)
		return
	}

	if p.game != nil {
		p.closeGame()
		p.game.conn = conn
	}

	go func() {
		for {
			msg, err := wsutil.ReadClientText(conn)
			if err != nil {
				log.Println("error while receiving message: ", err)
				p.closeGame()
				return
			}
			p.broker.Transfer(msg)
		}
	}()

	go func() {
		for msg := range p.game.out {
			if err := wsutil.WriteClientText(conn, msg); err != nil {
				log.Println("error while sending message: ", err)
				p.closeGame()
			}
		}
	}()
}

func (p *PoolS) closeGame() {
	p.game.lock.Lock()
	defer p.game.lock.Unlock()

	if p.game.conn != nil {
		if err := p.game.conn.Close(); err != nil {
			log.Println("error while closing connection: ", err)
		}
		p.game.conn = nil
	}
}
