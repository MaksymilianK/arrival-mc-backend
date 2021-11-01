package player

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/util/web"
)

type Service interface {
	all(SID string, page web.PageReq, nick string) (web.PageRes, error)
	one(SID string, nick string) (Res, error)
}

type serviceS struct {
	repo        Repo
	authService auth.Service
}

func NewService(repo Repo, authService auth.Service) Service {
	return serviceS{repo, authService}
}

func (s serviceS) all(SID string, page web.PageReq, nick string) (web.PageRes, error) {
	if _, err := s.authService.RequirePerm(SID, "player.view"); err != nil {
		return web.PageRes{}, err
	}
	return s.repo.getAll(page, nick)
}

func (s serviceS) one(SID string, nick string) (Res, error) {
	p, ok := s.authService.RequireAuth(SID)
	if !ok {
		return Res{}, web.ErrAuth
	}

	if !s.authService.HasPerm(p, "player.view") && p.Nick() != nick {
		return Res{}, web.ErrPerm
	}

	return s.repo.getOne(nick)
}
