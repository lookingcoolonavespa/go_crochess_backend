package repository_gameseeks

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/lookingcoolonavespa/go_crochess_backend/src/domain"
	"github.com/stretchr/testify/assert"
)

func initMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestGameseeksRepository_List(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "color", "time", "increment", "seeker"}).
		AddRow(0, "black", 3000, 0, "abcd").
		AddRow(1, "white", 5000, 5, "efgh")

	query := fmt.Sprintf(
		`SELECT * FROM gameseeks`,
	)

	mock.ExpectQuery(query).WillReturnRows(rows)

	r := NewGameseeksRepo(db)

	gameseeks, err := r.List(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, gameseeks)
	assert.Len(t, gameseeks, 2)
}

func TestGameseeksRepository_Insert(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	query := fmt.Sprintf(`
    INSERT INTO gameseeks (
        color,
        time,
        increment,
        seeker,
    ) VALUES (
        $1, $2, $3, $4
    )`,
	)

	color := "black"
	time := 30000
	increment := 5
	seeker := "fdafea"

	mock.ExpectExec(query).WithArgs(color, time, increment, seeker).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameseeksRepo(db)

	err := r.Insert(
		context.Background(),
		&domain.Gameseek{
			Color:     color,
			Time:      time,
			Increment: increment,
			Seeker:    seeker,
		})

	assert.NoError(t, err)
}

func TestGameseeksRepository_Delete(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	seeker1 := "fdafda"
	seeker2 := "faaaa"
	query := fmt.Sprintf(`
    DELETE FROM 
        gameseeks
    WHERE 
        seeker 
    IN (
        $1, $2
    )`,
	)

	mock.ExpectBegin()
	mock.ExpectExec(query).
		WithArgs(seeker1, seeker2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameseeksRepo(db)

	tx, err := r.db.Begin()
	err = r.Delete(context.Background(), tx, seeker1, seeker2)

	assert.NoError(t, err)
}
