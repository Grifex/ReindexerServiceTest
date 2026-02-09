package service

import (
	"context"
	"sort"

	"reindexer-service/internal/cache"
	"reindexer-service/internal/dto"
	"reindexer-service/internal/model"
	"reindexer-service/internal/repo"
)

type DocumentService struct {
	repo  *repo.DocumentRepo
	cache cache.DocCache
}

func NewDocumentService(r *repo.DocumentRepo, c cache.DocCache) *DocumentService {
	return &DocumentService{repo: r, cache: c}
}

func (s *DocumentService) Create(ctx context.Context, doc *model.Document) (*model.Document, error) {
	sortLevel1Desc(doc)

	created, err := s.repo.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *DocumentService) GetDTO(ctx context.Context, id int64) (dto.Document, bool, error) {
	if cached, ok, err := s.cache.Get(ctx, id); err == nil && ok {
		return cached, true, nil
	}

	doc, err := s.repo.Get(ctx, id)
	if err != nil {
		return dto.Document{}, false, err
	}

	dtoDoc := ToDocumentDTO(*doc)

	_ = s.cache.Set(ctx, id, dtoDoc)

	return dtoDoc, false, nil
}

func (s *DocumentService) Update(ctx context.Context, doc *model.Document) error {
	sortLevel1Desc(doc)

	if err := s.repo.Update(ctx, doc); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, doc.ID)
	return nil
}

func (s *DocumentService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, id)
	return nil
}

func (s *DocumentService) List(ctx context.Context, limit, offset int) (repo.ListResult, error) {
	result, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return repo.ListResult{}, err
	}

	for i := range result.Items {
		sortLevel1Desc(&result.Items[i])

	}
	return result, nil
}

func sortLevel1Desc(doc *model.Document) {
	if doc == nil || len(doc.Level1) <= 1 {
		return
	}

	sort.SliceStable(doc.Level1, func(i, j int) bool {
		return doc.Level1[i].Sort > doc.Level1[j].Sort
	})
}
