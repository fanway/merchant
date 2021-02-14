package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/jackc/pgx/v4"
)

var DATABASE_URL = fmt.Sprintf("postgresql://%s:%s@database:5432/avito", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))

// Очень смутила фраза "Ни один параметр не является обязательным",
// поэтому использую маску и свитч на 8 комбинаций параметров.
// Можно сделать через ссылки и проверять на nil и/или через CASE в
// sql query, но мне хотелось передавать разное количество параметров.
func getOffers(mask uint8, params ...interface{}) (names []string, err error) {
	var conn *pgx.Conn
	conn, err = pgx.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var rows pgx.Rows
	switch mask {
	case 7:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE id=$1 AND offer_id=$2 AND name LIKE $3", params[0].(int), params[1].(int), params[2].(string)+"%")
	case 6:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE id=$1 AND offer_id=$2", params[0].(int), params[1].(int))
	case 5:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE id=$1 AND name LIKE $2", params[0].(int), params[1].(string)+"%")
	case 4:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE id=$1", params[0].(int))
	case 3:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE id=$1 AND name LIKE $2", params[0].(int), params[1].(string)+"%")
	case 2:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE offer_id=$1", params[0].(int))
	case 1:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers WHERE name LIKE $1", params[0].(string)+"%")
	case 0:
		rows, err = conn.Query(context.Background(), "SELECT name FROM Offers")
	}
	if err != nil {
		log.Println(err)
		return
	}

	for rows.Next() {
		var str string
		err = rows.Scan(&str)
		if err != nil {
			log.Println(err)
			return
		}
		names = append(names, str)
	}
	return names, nil
}

func processXlsx(link string, id int) (countAdded, countUpdated, countDeleted, countErr int) {
	f, err := excelize.OpenFile(link)
	if err != nil {
		log.Println(err)
	}

	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	rows, err := f.GetRows("Offers")
	for i, row := range rows {
		if i == 0 {
			continue
		}
		var s [5]string
		for j, colCell := range row {
			s[j] = colCell
		}

		if s[4] == "false" {
			ct, err := conn.Exec(context.Background(), "DELETE FROM Offers WHERE id=$1 AND offer_id=$2", id, s[0])
			if err != nil || ct.RowsAffected() == 0 {
				log.Println(err)
				countErr++
				continue
			}
			countDeleted++
			continue
		}

		update := true
		_id := 0
		err := conn.QueryRow(context.Background(), "SELECT id FROM Offers WHERE id=$1 AND offer_id=$2", id, s[0]).Scan(&_id)

		if err != nil {
			update = false
		}

		var errs error
		offer_id, err := strconv.Atoi(s[0])
		errs = multierror.Append(errs, err)
		name := s[1]
		price, err := strconv.Atoi(s[2])
		errs = multierror.Append(errs, err)
		quantity, err := strconv.Atoi(s[3])

		if err != nil {
			countErr++
			log.Println(errs.Error())
			continue
		}

		err = conn.QueryRow(context.Background(), "INSERT INTO Offers (id, offer_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id, offer_id) DO UPDATE SET name = EXCLUDED.name, price = EXCLUDED.price, quantity = EXCLUDED.quantity RETURNING id", id, offer_id, name, price, quantity).Scan(&id)
		if err != nil {
			log.Println(err)
			countErr++
			continue
		}
		if update {
			countUpdated++
		} else {
			countAdded++
		}
	}
	return
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	_, err = conn.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS Offers (id int NOT NULL, offer_id int NOT NULL, name text NOT NULL, price int NOT NULL CHECK (price >= 0), quantity int NOT NULL, PRIMARY KEY (id, offer_id))")
	if err != nil {
		log.Println(err)
		return
	}
	added, updated, deleted, errors := processXlsx("./test.xlsx", 0)
	fmt.Printf("added: %d, updated: %d, deleted: %d, errors: %d\n", added, updated, deleted, errors)
	fmt.Println(getOffers(7, 0, 1, "t"))
	fmt.Println(getOffers(5, 0, "test"))
	fmt.Println(getOffers(1, "te"))
	fmt.Println(getOffers(0))
}
