package dto

type Document struct {
	ID     int64        `json:"id"`
	Title  string       `json:"title"`
	Level1 []Level1Item `json:"level1"`
}

type Level1Item struct {
	Sort   int          `json:"sort"`
	Name   string       `json:"name"`
	Level2 []Level2Item `json:"level2"`
}

type Level2Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
