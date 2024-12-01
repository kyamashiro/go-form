package repo

import (
	"database/sql"
	"fmt"
	"log"
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

	vals := make([]any, 0, len(values))
	for i, k := range values {
		insert += fmt.Sprintf(`(%v, ?),`, i)
		vals = append(vals, k)
	}
	insert = insert[:len(insert)-1]

	fmt.Println(insert)

	// prepareして実行
	stmt, err := w.db.Prepare(insert)
	if err != nil {
		log.Fatal("Prepare error: ", err)
		return err
	}

	if _, err := stmt.Exec(vals...); err != nil {
		log.Fatal("Exec error: ", err)
		return err
	}

	return nil
}
