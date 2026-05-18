package domain

// TrendPoint — daily trend cell.
type TrendPoint struct {
	Date     string `json:"date"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}

// HeatmapCell — hour × dow bucket.
type HeatmapCell struct {
	DayOfWeek int    `json:"dow"`
	Hour      int    `json:"hour"`
	Category  string `json:"category"`
	MS        uint64 `json:"ms"`
}

// LangCell — language breakdown cell.
type LangCell struct {
	Lang     string `json:"lang"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}
