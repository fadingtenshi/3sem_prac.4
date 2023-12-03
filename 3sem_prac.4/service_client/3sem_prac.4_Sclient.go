package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Структура для представления данных запроса
type StatsRequest struct {
	IPAddress   string    `json:"ip_address"`
	RedirectURL string    `json:"redirect_url"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	// Замените этот URL на ваш сервер
	serverURL := "http://localhost:6380/"

	// Создание экземпляра структуры StatsRequest
	statsData := StatsRequest{
		IPAddress:   "192.168.1.1",
		RedirectURL: "http://example.com",
		Timestamp:   time.Now(),
	}

	// Преобразование данных в формат JSON
	jsonData, err := json.Marshal(statsData)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Отправка POST-запроса на сервер
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Печать статуса ответа
	fmt.Println("Response Status:", resp.Status)
}
