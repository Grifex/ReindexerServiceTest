package repo

import (
	"context"
	"fmt"
	"reindexer-service/internal/model"

	"github.com/restream/reindexer"
)

type DocumentRepo struct {
	db        *reindexer.Reindexer
	namespace string
}

type ListResult struct {
	Items  []model.Document
	Total  int
	Limit  int
	Offset int
}

func NewDocumentRepo(db *reindexer.Reindexer, namespace string) *DocumentRepo {
	return &DocumentRepo{db: db, namespace: namespace}
}

func (r *DocumentRepo) Create(ctx context.Context, doc *model.Document) (*model.Document, error) {
	_, err := r.db.WithContext(ctx).Insert(r.namespace, doc, "id=serial()")
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocumentRepo) Update(ctx context.Context, doc *model.Document) error {
	_, err := r.db.WithContext(ctx).Update(r.namespace, doc)
	return err
}

func (r *DocumentRepo) Delete(ctx context.Context, id int64) error {
	tmp := &model.Document{ID: id}
	err := r.db.WithContext(ctx).Delete(r.namespace, tmp)
	return err
}

func (r *DocumentRepo) Get(ctx context.Context, id int64) (*model.Document, error) {
	it := r.db.WithContext(ctx).
		Query(r.namespace).
		WhereInt64("id", reindexer.EQ, id).
		Limit(1).
		Exec()

	item, err := it.FetchOne()
	if err != nil {
		return nil, err
	}

	switch v := item.(type) {
	case *model.Document:
		return v, nil
	case model.Document:
		return &v, nil
	default:
		return nil, fmt.Errorf("unexpected type from reindexer: %T", item)
	}
}

func (r *DocumentRepo) List(ctx context.Context, limit, offset int) (ListResult, error) {
	q := r.db.WithContext(ctx).
		Query(r.namespace).
		Limit(limit).
		Offset(offset).
		ReqTotal()

	it := q.Exec()
	total := it.TotalCount()

	raw, err := it.FetchAll()
	if err != nil {
		return ListResult{}, err
	}

	items := make([]model.Document, 0, len(raw))
	for _, x := range raw {
		switch v := x.(type) {
		case *model.Document:
			items = append(items, *v)
		case model.Document:
			items = append(items, v)
		default:
			return ListResult{}, fmt.Errorf("unexpected type from reindexer: %T", x)
		}
	}

	return ListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
