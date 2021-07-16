package server

type server struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type serversRes struct {
	Servers []server `json:"servers"`
}

const (
	// WebsiteID is a fixed ID of a special pseudo-server Website.
	WebsiteID = -2

	// NetworkID is a fixed ID of a special pseudo-server Network which represents the whole Minecraft servers network.
	NetworkID = -1
)
