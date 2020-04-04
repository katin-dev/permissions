package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
}

func main() {
	router := mux.NewRouter()
	router.Use(CheckAuth)
	router.HandleFunc("/api/v1/user/permissions", permissionHandler)

	err := http.ListenAndServe(":8085", router)

	if err != nil {
		fmt.Print(err)
	}
}

var CheckAuth = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		notAuth := []string{"/api/user/new", "/api/user/login"} //Список эндпоинтов, для которых не требуется авторизация
		requestPath := r.URL.Path                               //текущий путь запроса

		//проверяем, не требует ли запрос аутентификации, обслуживаем запрос, если он не нужен
		for _, value := range notAuth {

			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("Authorization") //Получение токена

		if tokenHeader == "" { //Токен отсутствует, возвращаем  403 http-код Unauthorized
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("require authorization"))
			return
		}

		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("invalid access token"))
			return
		}

		accessToken := splitted[1]

		fmt.Println(accessToken)

		userId, err := getUserIdByAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Println(userId)

		ctx := context.WithValue(r.Context(), "user_id", userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) //передать управление следующему обработчику!
	})
}

func getUserIdByAccessToken(accessToken string) (string, error) {
	hydraUrl := "http://127.0.0.1:4445/oauth2/introspect"
	resp, err := http.PostForm(
		hydraUrl,
		url.Values{"token": {accessToken}, "scope": {""}},
	)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.New("Invalid token")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read error: %s", err.Error())
		return "", err
	}

	token := make(map[string]interface{})
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Printf("Json error: %s", err.Error())
		return "", err
	}

	if token["active"] == false {
		return "", errors.New("Invalid token (is not active)")
	}

	return token["sub"].(string), nil
}

func getUserPermissions(userId string) ([]string, error) {
	rows, err := db.Query(`
		SELECT p.name 
		FROM permission p
		JOIN role_permission rp ON rp.permission_id = p.id
		JOIN user_role ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1 	
	`, userId)

	if err != nil {
		log.Println("DB error: " + err.Error())
	}

	defer rows.Close()

	permissions := make([]string, 0)

	for rows.Next() {
		permission := ""
		rows.Scan(&permission)
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, error(err)
	}

	return permissions, nil
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
