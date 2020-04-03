package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

/*
 Задача: принять HTTP запрос с Auth Token, отдать ошибку авторизации или пермишены пользователя
  1. [x] Поднять HTTP сервер
  2. [x] Написать обработчик: считывает из запроса UserID, возвращает список строк в виде JSON ответа
  3. Подключить БД и написать запрос для получения permissions
  4. Написать middleware для преобразования токена в user_id
*/

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
			// response = u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("require authorization"))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //Токен обычно поставляется в формате `Bearer {token-body}`, мы проверяем, соответствует ли полученный токен этому требованию
		if len(splitted) != 2 {
			// response = u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte("invalid access token"))
			return
		}

		tokenPart := splitted[1] //Получаем вторую часть токена

		fmt.Sprintf("Token %s", tokenPart)

		// Всё прошло хорошо, продолжаем выполнение запроса

		// userId := r.URL.Query().Get("user_id")
		userId := tokenPart

		fmt.Sprintf("User %s", userId) //Полезно для мониторинга

		ctx := context.WithValue(r.Context(), "user_id", userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) //передать управление следующему обработчику!
	})
}

func permissionHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.Header().Add("Content-Type", "application/json")

	// userId := "uuid"
	permissions := make([]string, 0)
	fmt.Println(r.Context().Value("user_id"))

	if r.Context().Value("user_id") == "sergey" {
		sergeyPermissions := []string{"read", "dance", "play"}
		permissions = append(permissions, sergeyPermissions...)
	}

	json, err := json.Marshal(permissions)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(json))
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
