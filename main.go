package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

/*
 Задача: принять HTTP запрос с Auth Token, отдать ошибку авторизации или пермишены пользователя
  1. [x] Поднять HTTP сервер
  2. Написать обработчик: считывает из запроса UserID, возвращает список строк в виде JSON ответа
  3. Подключить БД и написать запрос для получения permissions
  4. Написать middleware для преобразования токена в user_id
*/

func permissionHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.Write([]byte("Hello"))
	w.Header().Add("Content-Type", "application/json")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/user/permissions", permissionHandler)

	err := http.ListenAndServe(":8085", router)

	if err != nil {
		fmt.Print(err)
	}
}
