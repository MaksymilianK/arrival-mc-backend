package server

// Server represents a Minecraft server in the network. There are two special pseudo-servers: Website and Network.
type Server struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

type serversRes struct {
	servers []Server
}

const (
	// WebsiteID is a fixed ID of a special pseudo-server Website.
	WebsiteID = -2

	// NetworkID is a fixed ID of a special pseudo-server Network which represents the whole Minecraft servers network.
	NetworkID = -1
)
