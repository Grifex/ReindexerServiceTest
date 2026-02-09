package service

import (
	"reindexer-service/internal/dto"
	"reindexer-service/internal/model"
)

func ToDocumentDTO(doc model.Document) dto.Document {

	out := dto.Document{
		ID:     doc.ID,
		Title:  doc.Title,
		Level1: make([]dto.Level1Item, len(doc.Level1)),
	}

	for i, l1 := range doc.Level1 {
		l2 := make([]dto.Level2Item, len(l1.Level2))
		for j, x := range l1.Level2 {

			l2[j] = dto.Level2Item{
				Key:   x.Key,
				Value: x.Value,
			}
		}

		out.Level1[i] = dto.Level1Item{
			Sort:   l1.Sort,
			Name:   l1.Name,
			Level2: l2,
		}
	}

	return out
}
