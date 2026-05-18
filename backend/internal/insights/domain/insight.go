package domain

type Insight struct {
	Key  string         `json:"key"`
	Vars map[string]any `json:"vars,omitempty"`
}

type LangCell struct {
	Lang     string `json:"lang"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}

type TrendPoint struct {
	Date     string `json:"date"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}

type Inputs struct {
	Last7d, Prev7d map[string]uint64
	Languages      []LangCell
	Trend          []TrendPoint
}
