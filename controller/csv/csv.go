package csv

import (
	"encoding/csv"
	"fmt"
	"go-form/core/database"
	"go-form/core/session"
	"go-form/repo"
	"io"
	"log"
	"net/http"
	"strconv"
)

func Csv(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post(w, r)
	case http.MethodGet:
		get(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	// セッションマネージャの初期化
	manager, err := session.NewManager()
	if err != nil {
		log.Printf("Session Manager Initialization Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// セッション開始
	s, err := manager.SessionStart(w, r)
	if err != nil {
		log.Printf("Session Start Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := s.Values["user"]
	if user == nil {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	// ファイルを取得
	file, _, err := r.FormFile("csvfile")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uploaded file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// CSVリーダーを作成
	reader := csv.NewReader(file)
	isHeader := true // ヘッダー行をスキップするためのフラグ

	db := database.DB()
	defer db.Close()
	wsRepo := repo.NewWeatherStationRepository(db)

	var weatherStations []repo.WeatherStation
	// 1行ずつ読み込んで処理
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break // ファイルの終わりに到達
			}
			http.Error(w, fmt.Sprintf("Failed to read CSV: %v", err), http.StatusInternalServerError)
			return
		}

		// ヘッダー行をスキップ
		if isHeader {
			isHeader = false
			continue
		}

		// データを検証
		if len(record) < 3 {
			http.Error(w, "Invalid CSV format: insufficient columns", http.StatusBadRequest)
			return
		}
		var city string
		var temperature float32
		city = record[0]
		v, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			log.Fatalf("Failed to convert temperature: %v", err)
		}
		temperature = float32(v)

		weatherStations = append(weatherStations, repo.WeatherStation{City: city, Temperature: temperature})

		if len(weatherStations) >= 1000 {
			// データベースに挿入
			err = wsRepo.BulkInsert(weatherStations)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to insert record: %v", err), http.StatusInternalServerError)
				return
			}
			weatherStations = nil
		}
	}

	if weatherStations != nil {
		// データベースに挿入
		err = wsRepo.BulkInsert(weatherStations)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert record: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "File uploaded and data saved successfully!")
}

func get(w http.ResponseWriter, r *http.Request) {
	// セッションマネージャの初期化
	manager, err := session.NewManager()
	if err != nil {
		log.Printf("Session Manager Initialization Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// セッション開始
	s, err := manager.SessionStart(w, r)
	if err != nil {
		log.Printf("Session Start Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := s.Values["user"]
	if user == nil {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	// DB接続とリソースの解放を確実に行う
	db := database.DB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Database Close Error: %v", err)
		}
	}()

	// ユーザーリポジトリからデータ取得
	userRepo := repo.NewUserRepository(db)
	rows, err := userRepo.FindAll()
	if err != nil {
		log.Printf("Data Fetch Error: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// レスポンス用ヘッダー設定
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=sample.csv")

	// CSVライターの初期化
	writer := csv.NewWriter(w)

	// ヘッダー行を書き込み
	headers := []string{"id", "name", "created_at"}
	if err := writer.Write(headers); err != nil {
		log.Printf("CSV Header Write Error: %v", err)
		http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
		return
	}

	// データ行を書き込み
	for rows.Next() {
		var id, name, createdAt string
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			log.Printf("Row Scan Error: %v", err)
			http.Error(w, "Failed to scan data", http.StatusInternalServerError)
			return
		}

		record := []string{id, name, createdAt}
		if err := writer.Write(record); err != nil {
			log.Printf("CSV Record Write Error: %v", err)
			http.Error(w, "Failed to write CSV record", http.StatusInternalServerError)
			return
		}
	}

	// データ取得中のエラー確認
	if err := rows.Err(); err != nil {
		log.Printf("Row Iteration Error: %v", err)
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		return
	}

	// 最後にバッファをフラッシュ
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("CSV Writer Flush Error: %v", err)
		http.Error(w, "Error finalizing CSV", http.StatusInternalServerError)
	}
}
