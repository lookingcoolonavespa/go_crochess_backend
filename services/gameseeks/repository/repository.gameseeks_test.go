package repository_gameseeks

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lookingcoolonavespa/go_crochess_backend/domain"
	"github.com/spf13/viper"
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

	query := fmt.Sprintf(`
    SELECT * FROM %s.gameseeks
    `, viper.GetString("database.schema"),
	)

	prep := mock.ExpectPrepare(query)
	prep.ExpectQuery().WillReturnRows(rows)

	r := NewGameseeksRepo(db)

	gameseeks, err := r.List()

	assert.NoError(t, err)
	assert.NotNil(t, gameseeks)
	assert.Len(t, gameseeks, 2)
}

func TestGameseeksRepository_Insert(t *testing.T) {
	db, mock := initMock()

	defer db.Close()

	query := fmt.Sprintf(`
    INSERT INTO %s.gameseeks (
        color,
        time,
        increment,
        seeker,
    ) VALUES (
        $1, $2, $3, $4
    )`,
		viper.GetString("database.schema"),
	)
	prep := mock.ExpectPrepare(query)

	color := "black"
	time := 30000
	increment := 5
	seeker := "fdafea"
	prep.ExpectExec().WithArgs(color, time, increment, seeker).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameseeksRepo(db)

	err := r.Insert(&domain.Gameseek{
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
    DELETE FROM %s.gameseeks
    WHERE 
        seeker IN ('%s', '%s')`,
		viper.GetString("database.schema"),
		seeker1,
		seeker2,
	)

	prep := mock.ExpectPrepare(query)

	prep.ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := NewGameseeksRepo(db)

	err := r.Delete(seeker1, seeker2)

	assert.NoError(t, err)
}
