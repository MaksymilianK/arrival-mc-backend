package server

type Service interface {
	// Exists checks if a server with the given ID exists.
	Exists(ID int) bool

	all() serversRes
}

type serviceS struct {
	repo Repo
	servers serversRes
	byID map[int]server
}

func NewService(repo Repo) Service {
	service := &serviceS{repo: repo, byID: make(map[int]server)}

	servers, err := repo.getAll()
	if err != nil {
		panic(err)
	}

	service.servers = serversRes{servers}
	for _, s := range servers {
		service.byID[s.ID] = s
	}

	return service
}

func (s *serviceS) Exists(ID int) bool {
	_, ok := s.byID[ID]
	return ok
}

func (s *serviceS) all() serversRes {
	return s.servers
}
