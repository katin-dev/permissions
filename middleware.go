package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

var AuthorizeRequest = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Требуем присылать access_token в формате Authorization: Bearer {auth_token}
		accessToken := getAccessTokenFromHeader(r.Header)

		if accessToken == "" {
			http.Error(w, "Forbidden", 401)
		}

		userId, err := h.getUserIdByAccessToken(accessToken)
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

// Вытаскиваем access_token из заголовков запроса
func getAccessTokenFromHeader(h http.Header) string {
	tokenHeader := h.Get("Authorization")

	if tokenHeader == "" {
		return ""
	}

	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		return ""
	}

	return splitted[1]
}
