package domain

type Message struct {
	RequestID string `json:"-"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Text      string `json:"text"`
}

type MetaIP struct {
	IP      string `json:"ip"`
	ISP     string `json:"isp"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type MetaReq struct {
	RequestID string
	IP        string
	Path      string
	Useragent string
}

type Query struct {
	Pass string `json:"pass"`
}

type Log struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	ISP       string `json:"isp"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Date      string `json:"date"`
	Path      string `json:"path"`
	Useragent string `json:"useragent"`
}
