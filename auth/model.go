package auth

type rankMin struct {
	ID          int    `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ChatFormat  string `json:"chatFormat"`
}

type minRanks struct {
	Ranks []*rankMin `json:"ranks"`
}

type rankWithPerms struct {
	rankMin
	Perms []string `json:"permissions"`
	NegatedPerms []string `json:"negatedPermissions"`
}

type ranksWithPerms struct {
	Ranks []*rankWithPerms `json:"ranks"`
}

type rankFull struct {
	rankMin
	Perms map[int][]string `json:"permissions"`
	NegatedPerms map[int][]string `json:"negatedPermissions"`
}

type rankCreation struct {
	Level       int
	Name        string
	DisplayName string
	ChatFormat  string
	Perms       map[int][]string `json:"permissions"`
	NegatedPerms map[int][]string `json:"negatedPermissions"`
}

type rankModification struct {
	Level       int
	Name        string
	DisplayName string
	ChatFormat  string
	RemovedPerms    map[int][]string `json:"removedPermissions"`
	AddedPerms  map[int][]string
	RemovedNegatedPerms map[int][]string `json:"removedNegatedPermissions"`
	AddedNegatedPerms map[int][]string `json:"addedNegatedPermissions"`
}

type Rank struct {
	id             int
	level          int
	allPerms       map[string]struct{}
	allNegatedPerms map[string]struct{}
	effectivePerms map[string]struct{}
}

type Player struct {
	id   int
	nick string
	rank *Rank
}

type playerMin struct {
	Nick string         `json:"nick"`
	Rank *rankWithPerms `json:"rank"`
}

type playerCredentials struct {
	id       int
	passHash string
	rank     int
}

type loginForm struct {
	Nick     string
	Password string
}

type rankList struct {
	Server int
}

const (
	rankLvlDef   = 1000
	rankLvlOwner = 32767
	rankIDDef   = -2
	rankIDOwner = -1

	permRankView   = "rank.view"
	permRankModify = "rank.modifyRank"
)

func (r *Rank) ID() int {
	return r.id
}

func (r *Rank) Level() int {
	return r.level
}

func (p *Player) ID() int {
	return p.id
}

func (p *Player) Nick() string {
	return p.nick
}

func (p *Player) Rank() *Rank {
	return p.rank
}
