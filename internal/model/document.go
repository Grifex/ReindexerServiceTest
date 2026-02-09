package model

type Document struct {
	ID    int64  `json:"id" reindex:"id,hash,pk"`
	Title string `json:"title" reindex:"title,text"`

	RemovableField string `json:"remove-field"`

	Level1 []Level1 `json:"level1"`
}

type Level1 struct {
	Sort int    `json:"sort"`
	Name string `json:"name"`

	RemovableField string `json:"remove-field"`

	Level2 []Level2 `json:"level2"`
}

type Level2 struct {
	Key   string `json:"key"`
	Value string `json:"value"`

	RemovableField string `json:"remove-field"`
}
