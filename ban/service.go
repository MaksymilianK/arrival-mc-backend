package ban

import (
	"github.com/maksymiliank/arrival-mc-backend/auth"
	"github.com/maksymiliank/arrival-mc-backend/server"
	v "github.com/maksymiliank/arrival-mc-backend/util/validator"
	"time"
)

type Service interface {
	all(SID string, req banReq) ([]*banMin, error)
	createOne(SID string, ban banCreationReq) (int, error)
	modifyOne(SID string, ID int, ban banModificationReq) (int, error)
	deleteOne(SID string, ID int, removalReason string) error
}

type serviceS struct {
	repo Repo
	serverService server.Service
	authService auth.Service
}

func NewService(repo Repo, serverService server.Service, authService auth.Service) Service {
	return serviceS{repo, serverService, authService}
}

func (s serviceS) all(SID string, req banReq) ([]*banMin, error) {
	if _, err := s.authService.RequirePerm(SID, "ban.view"); err != nil {
		return nil, err
	}

	if err := v.Validate(
		req.server == 0 || s.serverService.Exists(req.server),
		req.recipient == "" || s.authService.NickValid(req.recipient),
		req.startFrom.IsZero() || req.startTo.IsZero() || !req.startFrom.After(req.startTo),
		req.expirationFrom.IsZero() || req.expirationTo.IsZero() || !req.expirationFrom.After(req.expirationTo),
		req.startFrom.IsZero() || req.expirationFrom.IsZero() || !req.startFrom.After(req.expirationFrom),
		req.startTo.IsZero() || req.expirationTo.IsZero() || !req.startTo.After(req.expirationTo),
	); err != nil {
		return nil, err
	}

	return s.repo.getAll(req)
}

func (s serviceS) createOne(SID string, ban banCreationReq) (int, error) {
	p, err := s.authService.RequirePerm(SID, "ban.give")
	if err != nil {
		return 0, err
	}

	if err := v.Validate(
		s.serverService.Exists(ban.Server),
		s.authService.NickValid(ban.Recipient),
		ban.Duration >= 24 * time.Hour && ban.Duration <= 365 * 24 * time.Hour,
		ban.Reason != "" && len(ban.Reason) <= 255,
	); err != nil {
		return 0, err
	}

	return s.repo.createOne(banCreation{
		ban.Server,
		ban.Recipient,
		p.ID(),
		ban.Duration,
		ban.Reason,
	})
}

func (s serviceS) modifyOne(SID string, ID int, ban banModificationReq) (int, error) {
	p, err := s.authService.RequirePerm(SID, "ban.modify")
	if err != nil {
		return 0, err
	}

	if err := v.Validate(
		ban.Server == 0 || s.serverService.Exists(ban.Server),
		ban.Recipient == "" || s.authService.NickValid(ban.Recipient),
		ban.Duration == 0 || ban.Duration >= 24 * time.Hour && ban.Duration <= 365 * 24 * time.Hour,
		ban.Reason == "" || ban.Reason != "" && len(ban.Reason) <= 255,
		len(ban.ModificationReason) > 0 && len(ban.ModificationReason) < 256,
	); err != nil {
		return 0, err
	}

	return s.repo.modifyOne(ID, p.ID(), ban)
}

func (s serviceS) deleteOne(SID string, ID int, removalReason string) error {
	p, err := s.authService.RequirePerm(SID, "ban.modify")
	if err != nil {
		return err
	}

	if err := v.Validate(
		len(removalReason) > 0 && len(removalReason) < 256,
	); err != nil {
		return err
	}

	return s.repo.deleteOne(ID, p.ID(), removalReason)
}
