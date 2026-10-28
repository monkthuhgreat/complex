package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// Переменные уровня пакета для хранения конфигурации и прокси-движка
var pythonURLStr = os.Getenv("SERVICE_URL")
var proxy *httputil.ReverseProxy

func main() {
	// Если переменная окружения не задана, шлюз экстренно завершает работу сразу при старте
	if pythonURLStr == "" {
		log.Fatal("SERVICE_URL environment variable is required")
	}

	// Валидация строки URL (проверка на корректность протокола, хоста и спецсимволов)
	targetURL, err := url.Parse(pythonURLStr)
	if err != nil {
		log.Fatalf("invalid SERVICE_URL: %v", err)
	}

	// Создание прокси, который автоматически будет менять хост в заголовках
	// и прозрачно пересылать HTTP-методы, Headers, Cookies и Request Body
	proxy = httputil.NewSingleHostReverseProxy(targetURL)

	// Если целевой Python-сервис начнет «виснуть» или долго отвечать, шлюз разорвет соединение через 15 секунд,
	// предотвращая исчерпание лимитов открытых дескрипторов ОС.
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 15 * time.Second,
	}

	// Точные системные эндпоинты шлюза. Будут обрабатываться внутри Go.
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/meta", metaHandler)

	// Корневой роут работает как catch-all (перехватывает абсолютно все остальные запросы)
	http.HandleFunc("/", proxyHandler)

	log.Println("go-gateway listening on :8080")
	// Прослушивание порта 8080. nil означает использование стандартного встроенного роутера (http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Внутренний хэндлер для проверки жизнеспособности самого шлюза (например, для Docker/Kubernetes healthcheck)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"go-gateway"}`))
}

// Внутренний хэндлер для отладки, возвращающий текущие параметры конфигурации шлюза
func metaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "gateway",
		"python":  pythonURLStr,
	})
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Если запрос пришел на внутренние пути, принудительно перенаправляем их в локальные хэндлеры.
	if r.URL.Path == "/health" {
		healthHandler(w, r)
		return
	}
	if r.URL.Path == "/meta" {
		metaHandler(w, r)
		return
	}

	// Валидация эндпоинта "/items". Разрешен только метод GET.
	if r.URL.Path == "/items" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Валидация динамических эндпоинтов вида "/items/123", "/items/create". Разрешен только POST.
	} else if strings.HasPrefix(r.URL.Path, "/items/") {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	} else {
		http.NotFound(w, r)
		return
	}

	proxy.ServeHTTP(w, r)
}
