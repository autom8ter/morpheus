package imdb

import (
	"context"
	"github.com/autom8ter/morpheus/pkg/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func Test(t *testing.T) {
	db, err := sqlx.Open("mysql", "guest:relational@tcp(relational.fit.cvut.cz)/imdb_small")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	//
	cli := client.NewClient("morpheus", "morpheus", "http://localhost:8080/query", 5*time.Minute)
	if err := ImportIMDB(db)(context.Background(), cli); err != nil {
		t.Fatal(err)
	}
}
