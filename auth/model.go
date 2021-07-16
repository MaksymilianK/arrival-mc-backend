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
	Perms []string `json:"perms"`
}

type rankFull struct {
	rankMin
	Perms map[int][]string `json:"perms"`
}

type rankCreation struct {
	Level       int
	Name        string
	DisplayName string
	ChatFormat  string
	Perms       map[int][]string
}

type rankModification struct {
	Level       int
	Name        string
	DisplayName string
	ChatFormat  string
	RemPerms    map[int][]string
	AddedPerms  map[int][]string
}

type Rank struct {
	id             int
	level          int
	allPerms       map[string]struct{}
	effectivePerms map[string]struct{}
}

type Player struct {
	id   int
	nick string
	rank *Rank
}

type playerAuthRes struct {
	ID int `json:"id"`
	Nick string `json:"nick"`
	Rank *rankWithPerms `json:"rank"`
}

type playerCredentials struct {
	id int
	passHash string
	rank int
}

type loginForm struct {
	Nick string
	Pass string
}

const RankLvlDef = 1000
const RankLvlOwner = 32767
const RankIDDef = -2
const RankIDOwner = -1

const PermRankView = "rank.view"
const PermRankModify = "rank.modifyRank"

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