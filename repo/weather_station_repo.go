package repo

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type WeatherStation struct {
	City        string
	Temperature float32
}

type WeatherStationRepository struct {
	db *sql.DB
}

func NewWeatherStationRepository(db *sql.DB) *WeatherStationRepository {
	return &WeatherStationRepository{db: db}
}

func (w *WeatherStationRepository) BulkInsert(values []WeatherStation) error {
	insert := "INSERT INTO weather_stations(city, temperature) VALUES "

	placeholders := make([]string, 0, len(values))
	vals := make([]any, 0, len(values)*2)

	for i, k := range values {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		vals = append(vals, k.City, k.Temperature)
	}

	// Join placeholders and finalize the query
	insert += strings.Join(placeholders, ", ")

	fmt.Println("Final Query: ", insert)

	// Prepare and execute
	stmt, err := w.db.Prepare(insert)
	if err != nil {
		log.Fatal("Prepare error: ", err)
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(vals...); err != nil {
		log.Fatal("Exec error: ", err)
		return err
	}

	return nil
}
