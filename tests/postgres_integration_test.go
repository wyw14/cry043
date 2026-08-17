package tests

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
	"os"
	"testing"
	"time"
)

func TestSpecificationRoundTrip(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	id := "integration-" + time.Now().Format("150405.000000")
	want := domain.Specification{ID: id, DocumentID: id, Version: 1, Status: domain.SpecDraft, Revision: 1}
	got, err := repo.SaveSpecification(ctx, want, 0, id)
	if err != nil || got.ID != id {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}
