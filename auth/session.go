package auth

import (
	"encoding/base64"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"sync"
	"time"
)

type session struct {
	SID string
	player *Player
	expiration time.Time
}

type SessionManager interface {
	monitor()
	new(p *Player) (string, error)
	find(SID string) (*Player, bool)
	extendIfExists(SID string)
	remove(SID string) bool
}

type sessionManagerS struct{
	crypto Crypto
	bySID map[string]*session
	byExpiration *rbt.Tree
	lock sync.RWMutex
}

const sessionExpiration = 15 * time.Minute
const monitorDuration = 30 * time.Second

func NewSessionManager(crypto Crypto) SessionManager {
	return &sessionManagerS{
		crypto: crypto,
		bySID: make(map[string]*session),
		byExpiration: rbt.NewWith(compareSessions),
	}
}

func (s *sessionManagerS) monitor() {
	for {
		s.lock.Lock()
		s.walkAndRemove(s.byExpiration.Root, time.Now())
		s.lock.Unlock()

		time.Sleep(monitorDuration)
	}
}

func (s *sessionManagerS) new(p *Player) (string, error) {
	SID, err := s.generateSID()
	if err != nil {
		return "", err
	}
	session := &session{SID, p, time.Now().Add(sessionExpiration)}

	s.lock.Lock()
	s.bySID[SID] = session
	s.byExpiration.Put(session, struct{}{})
	s.lock.Unlock()

	return SID, nil
}

func (s *sessionManagerS) find(SID string) (*Player, bool) {
	s.lock.RLock()
	session, ok := s.bySID[SID]
	s.lock.RUnlock()

	if ok {
		return session.player, true
	} else {
		return nil, false
	}
}

func (s *sessionManagerS) extendIfExists(SID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	session, ok := s.bySID[SID]
	if ok {
		s.byExpiration.Remove(session)
		session.expiration = time.Now().Add(sessionExpiration)
		s.byExpiration.Put(session, struct{}{})
	}
}

func (s *sessionManagerS) remove(SID string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	session, ok := s.bySID[SID]
	if ok {
		delete(s.bySID, SID)
		s.byExpiration.Remove(session)
		return true
	} else {
		return false
	}
}

func (s *sessionManagerS) walkAndRemove(n *rbt.Node, time time.Time) {
	if n == nil {
		return
	}

	session := n.Key.(*session)

	if session.expiration.After(time) {
		s.walkAndRemove(n.Left, time)
	} else {
		s.walkAndRemove(n.Left, time)
		s.walkAndRemove(n.Right, time)
		s.remove(session.SID)
	}
}

func (s *sessionManagerS) generateSID() (string, error) {
	SID, err := s.crypto.Rand(32)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(SID), nil
}

func compareSessions(a interface{}, b interface{}) int {
	s1 := a.(*session)
	s2 := b.(*session)

	if s1.expiration.Before(s2.expiration) {
		return -1
	} else if s1.expiration.After(s2.expiration) {
		return 1
	} else if s1.SID == s2.SID {
		return 0
	} else {
		return utils.StringComparator(s1.SID, s2.SID)
	}
}
