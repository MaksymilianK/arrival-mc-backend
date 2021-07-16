package auth

import (
	"github.com/maksymiliank/arrival-mc-backend/db"
	"github.com/maksymiliank/arrival-mc-backend/server"
	v "github.com/maksymiliank/arrival-mc-backend/validator"
	"github.com/maksymiliank/arrival-mc-backend/web"
	"log"
	"regexp"
	"sort"
	"sync"
)

type Service interface {
	// RankExists checks if a rank with the given ID exists.
	RankExists(ID int) bool
	RequireAuth(SID string) (*Player, bool)
	RequirePerm(SID string, perm string) (*Player, error)
	HasPerm(p *Player, perm string) bool

	rankWithWebPerms(ID int) (*rankWithPerms, error)
	allRanks() minRanks
	oneRank(ID int) (rankFull, error)
	createRank(rank rankCreation) (int, error)
	removeRank(ID int) error
	modifyRank(ID int, rank rankModification) error

	current(SID string) (playerAuthRes, error)
	signIn(data loginForm) (playerAuthRes, string, error)
	signOut(SID string) bool
}

type serviceS struct{
	repo Repo
	sessions SessionManager
	crypto Crypto
	ranksWithWebPerms map[int]*rankWithPerms
	minRanks minRanks
	byLevel []*Rank
	byID map[int]*Rank
	lock sync.RWMutex
}

var permRegex = regexp.MustCompile(`^!?[A-Za-z0-9]+(.[A-Za-z0-9]+)*$`)
var nickRegex = regexp.MustCompile(`^\w{3,16}$`)

func NewService(repo Repo, sessions SessionManager, crypto Crypto) Service {
	service := &serviceS{
		repo: repo,
		sessions: sessions,
		crypto: crypto,
		ranksWithWebPerms: make(map[int]*rankWithPerms),
		byLevel: make([]*Rank, 0),
		byID: make(map[int]*Rank),
	}

	ranks, err := repo.getAllWebRanks()
	if err != nil {
		panic(err)
	}

	service.setRanks(ranks)
	return service
}

func (s *serviceS) RankExists(ID int) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.byID[ID]
	return ok
}

func (s *serviceS) RequireAuth(SID string) (*Player, bool) {
	return s.sessions.find(SID)
}

func (s *serviceS) RequirePerm(SID string, perm string) (*Player, error) {
	p, ok := s.sessions.find(SID)
	if !ok {
		return nil, web.ErrAuth
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.HasPerm(p, perm) {
		return nil, web.ErrPerm
	}

	return p, nil
}

func (s *serviceS) HasPerm(p *Player, perm string) bool {
	s.lock.RLock()
	s.lock.RUnlock()

	_, ok := p.rank.effectivePerms[perm]
	return ok
}

func (s *serviceS) rankWithWebPerms(ID int) (*rankWithPerms, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.RankExists(ID) {
		return nil, v.ErrNotFound
	}
	return s.ranksWithWebPerms[ID], nil
}

func (s *serviceS) allRanks() minRanks {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.minRanks
}

func (s *serviceS) oneRank(ID int) (rankFull, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.RankExists(ID) {
		return rankFull{}, v.ErrNotFound
	}

	perms, err := s.repo.getAllPerms(ID)
	if err != nil {
		log.Println(err)
		return rankFull{}, db.ErrPersistence
	}

	r := s.ranksWithWebPerms[ID]
	return rankFull{
		rankMin{r.ID, r.Level, r.Name, r.DisplayName, r.ChatFormat},
		perms,
	}, nil
}

func (s *serviceS) createRank(rank rankCreation) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := validateCreation(rank); err != nil {
		return 0, err
	}

	ID, err := s.repo.createRank(rank)
	if err != nil {
		return 0, db.ErrPersistence
	}

	perms := rank.Perms[server.WebsiteID]
	if perms == nil {
		perms = make([]string, 0)
	}

	s.ranksWithWebPerms[ID] = &rankWithPerms{
		rankMin{ID, rank.Level, rank.Name, rank.DisplayName, rank.ChatFormat},
		nil,
	}
	s.minRanks.Ranks = append(s.minRanks.Ranks, &rankMin{
		ID,
		rank.Level,
		rank.Name,
		rank.DisplayName,
		rank.ChatFormat,
	})

	r := &Rank{id: ID, level: rank.Level, allPerms: make(map[string]struct{})}
	s.byLevel = append(s.byLevel, r)
	s.byID[ID] = r
	s.orderRanks()

	return ID, nil
}

func (s *serviceS) removeRank(ID int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := s.repo.removeRank(ID); err != nil {
		return err
	}

	delete(s.ranksWithWebPerms, ID)

	minRanks := s.minRanks.Ranks
	s.minRanks.Ranks = make([]*rankMin, len(minRanks) - 1)
	for _, r := range minRanks {
		if r.ID != ID {
			s.minRanks.Ranks = append(s.minRanks.Ranks, r)
		}
	}

	delete(s.byID, ID)

	byLevel := s.byLevel
	s.byLevel = make([]*Rank, len(byLevel) - 1)
	for _, r := range byLevel {
		if r.id != ID {
			s.byLevel = append(s.byLevel, r)
		}
	}

	s.orderRanks()
	return nil
}

func (s *serviceS) modifyRank(ID int, rank rankModification) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if err := validateModification(ID, rank); err != nil {
		return err
	}

	err := s.repo.modifyRank(ID, rank)
	if err != nil {
		return db.ErrPersistence
	}

	if rank.Level != 0 {
		s.ranksWithWebPerms[ID].Level = rank.Level
		s.byID[ID].level = rank.Level
	}

	if rank.Name != "" {
		s.ranksWithWebPerms[ID].Name = rank.Name
	}

	if rank.DisplayName != "" {
		s.ranksWithWebPerms[ID].DisplayName = rank.DisplayName
	}

	if rank.ChatFormat != "" {
		s.ranksWithWebPerms[ID].ChatFormat = rank.ChatFormat
	}

	s.orderPerms(0)
	return nil
}

func (s *serviceS) current(SID string) (playerAuthRes, error) {
	p, ok := s.sessions.find(SID)
	if !ok {
		return playerAuthRes{}, web.ErrNotFound
	}
	return playerAuthRes{p.id, p.nick, s.ranksWithWebPerms[p.rank.id]}, nil
}

func (s *serviceS) signIn(data loginForm) (playerAuthRes, string, error) {
	if err := v.Validate(
		nickRegex.MatchString(data.Nick),
		len(data.Pass) > 5 && len(data.Pass) <= 50,
	); err != nil {
		return playerAuthRes{}, "", err
	}

	p, err := s.repo.getPlayerCredentials(data.Nick)
	if err != nil {
		return playerAuthRes{}, "", err
	}

	if err := s.crypto.verifyPass(data.Pass, p.passHash); err != nil {
		if err == ErrWrongPass {
			return playerAuthRes{}, "", web.ErrAuth
		} else {
			return playerAuthRes{}, "", err
		}
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	SID, err := s.sessions.new(&Player{p.id, data.Nick, s.byID[p.rank]})
	if err != nil {
		return playerAuthRes{}, "", err
	}

	return playerAuthRes{p.id, data.Nick, s.ranksWithWebPerms[p.rank]}, SID, nil
}

func (s *serviceS) signOut(SID string) bool {
	return s.sessions.remove(SID)
}

func (s *serviceS) setRanks(ranksWeb []rankWithPerms) {
	ranks := make([]*rankMin, 0)
	for _, r := range ranksWeb {
		s.ranksWithWebPerms[r.ID] = &r

		ranks = append(ranks, &rankMin{
			r.ID,
			r.Level,
			r.Name,
			r.DisplayName,
			r.ChatFormat,
		})

		rank := &Rank{id: r.ID, level: r.Level, allPerms: make(map[string]struct{})}
		for _, p := range r.Perms {
			rank.allPerms[p] = struct{}{}
		}

		s.byLevel = append(s.byLevel, rank)
		s.byID[rank.id] = rank
	}
	s.minRanks = minRanks{ranks}
	s.orderPerms(0)
}

func (s *serviceS) orderRanks() {
	sort.Slice(s.byLevel, func(i, j int) bool {
		return s.byLevel[i].level < s.byLevel[j].level
	})
	sort.Slice(s.minRanks.Ranks, func(i, j int) bool {
		return s.minRanks.Ranks[i].Level < s.minRanks.Ranks[j].Level
	})
	s.orderPerms(0)
}

func (s *serviceS) orderPerms(startIndex int) {
	inherited := make(map[string]struct{})
	for _, r := range s.byLevel[startIndex:] {
		for p := range r.allPerms {
			if p[0] == '!' {
				delete(inherited, p[1:])
			} else {
				inherited[p] = struct{}{}
			}
		}

		s.ranksWithWebPerms[r.id].Perms = make([]string, 0)
		r.effectivePerms = make(map[string]struct{})
		for p := range inherited {
			r.effectivePerms[p] = struct{}{}
			s.ranksWithWebPerms[r.id].Perms = append(s.ranksWithWebPerms[r.id].Perms, p)
		}
	}
}

func validateCreation(rank rankCreation) error {
	if err := v.Validate(
		rank.Level > 0 && rank.Level < RankLvlOwner,
		!v.InSlice(rank.Level, RankLvlDef, RankLvlOwner),
		len(rank.Name) > 0 && len(rank.Name) <= 30,
		len(rank.DisplayName) > 0 && len(rank.DisplayName) <= 75,
		len(rank.ChatFormat) > 0 && len(rank.ChatFormat) <= 200,
		permsValid(rank.Perms),
	); err != nil {
		return err
	}
	return nil
}

func validateModification(ID int, rank rankModification) error {
	if err := v.Validate(
		(ID > 0 && ID < 32768) || v.InSlice(ID, RankIDDef, RankIDOwner),
		rank.Level >= 0 && rank.Level < RankLvlOwner,
		rank.Level == 0 || !v.InSlice(ID, RankIDDef, RankIDOwner),
		!v.InSlice(rank.Level, RankLvlDef, RankLvlOwner),
		len(rank.Name) >= 0 && len(rank.Name) <= 30,
		len(rank.DisplayName) >= 0 && len(rank.DisplayName) <= 75,
		len(rank.ChatFormat) >= 0 && len(rank.ChatFormat) <= 200,
		rank.AddedPerms == nil || permsValid(rank.AddedPerms),
		rank.AddedPerms == nil || permsValid(rank.RemPerms),
	); err != nil {
		return err
	}

	if rank.RemPerms != nil && rank.RemPerms[server.WebsiteID] != nil {
		for _, p := range rank.RemPerms[server.WebsiteID] {
			if v.InSlice(p, PermRankView, PermRankModify) {
				return v.ErrValidation
			}
		}
	}
	return nil
}

func permsValid(perms map[int][]string) bool {
	for _, sp := range perms {
		for _, p := range sp {
			if !permRegex.MatchString(p) {
				return false
			}
		}
	}
	return true
}
