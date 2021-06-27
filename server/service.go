package server

type ServiceI interface {
	// All returns all the servers.
	All() []Server

	// Exists checks a server with the given ID exists.
	Exists(ID int) bool
}

type serviceS struct{}

var service ServiceI = serviceS{}

func (serviceS) All() []Server {
	return repo.getAll()
}

func (serviceS) Exists(ID int) bool {
	return repo.existsByID(ID)
}
