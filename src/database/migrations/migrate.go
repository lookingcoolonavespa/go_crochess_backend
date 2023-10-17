package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/spf13/viper"
)

func Up(db *sql.DB) error {
	fmt.Println(filepath.Abs("./src/database/migrations/*.sql"))
	files, err := filepath.Glob("./src/database/migrations/*.sql")
	if err != nil {
		return err

	}

	sort.Strings(files)

	err = db.Ping()
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file)

		log.Printf("running migration %s", filename)

		args :=
			[]string{"-h",
				viper.GetString("database.host"),
				"-p",
				strconv.Itoa(viper.GetInt("database.port")),
				"-U",
				viper.GetString("database.username"),
				"-d",
				viper.GetString("database.name"),
				"-f",
				file,
			}

		cmd := exec.Command("psql", args...)
		log.Println(cmd)

		err = cmd.Run()
		if err != nil {
			log.Printf("failed executing the sql file: %s\n", filename)
			log.Printf("all further migrations cancelled\n\n")
			return err
		}

		log.Printf("completed migration %s\n", filename)
	}

	log.Printf("completed all migrations")
	return nil
}
