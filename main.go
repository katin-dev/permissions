package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
 Задача: принять HTTP запрос с Auth Token, отдать ошибку авторизации или пермишены пользователя
  1. [x] Поднять HTTP сервер
  2. [x] Написать обработчик: считывает из запроса UserID, возвращает список строк в виде JSON ответа
  3. [x] Подключить БД и написать запрос для получения permissions
	* Подключить библиотеку
	* написать SQL, чтоб получить пермишены пользователя
  4. [x] Написать middleware для преобразования токена в user_id
*/

var db *sql.DB
var h *hydra

func init() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbType := os.Getenv("DB_TYPE")
	dbUrl := os.Getenv("DB_URL")

	db, err = sql.Open(dbType, dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	hydraUrl := os.Getenv("HYDRA_URL")

	h = NewHydra(hydraUrl)
}

func main() {
	router := mux.NewRouter()

	router.Use(AuthorizeRequest)

	router.HandleFunc("/api/v1/user/permissions", permissionHandler)

	err := http.ListenAndServe(":8085", router)

	if err != nil {
		log.Fatal(err)
	}
}

func permissionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	userId := r.Context().Value("user_id").(string)

	permissions, err := getUserPermissions(userId)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	json, err := json.Marshal(permissions)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(json))
}
