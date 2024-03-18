package main

import (
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/sirupsen/logrus"
)

func init() {
    // Устанавливаем формат логов в JSON
    logrus.SetFormatter(&logrus.JSONFormatter{})

    // Загрузка переменных среды из файла .env
    if err := godotenv.Load(); err != nil {
        logrus.Warn("No .env file found")
    }

    // Устанавливаем уровень логирования
    logLevelStr := os.Getenv("LOG_LEVEL")
    logLevel, err := logrus.ParseLevel(logLevelStr)
    if err != nil {
        logrus.Warnf("Invalid LOG_LEVEL value '%s', using default level 'info'", logLevelStr)
        logLevel = logrus.InfoLevel
    }
    logrus.SetLevel(logLevel)

    // Устанавливаем выходной поток логов в стандартный вывод
    logrus.SetOutput(os.Stdout)
}

func main() {
    // Создание нового коллектора Twilio
    twilioCollector := NewTwilioCollector()

    // Регистрация коллектора Twilio в реестре метрик Prometheus
    prometheus.MustRegister(twilioCollector)

    // Запуск
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<html>
             <head><title>Twilio statistics exporter</title></head>
             <body>
             <h1>Twilio statistics exporter</h1>
             <p><a href='metrics'>Metrics</a></p>
             </body>
             </html>`))
    })
    http.Handle("/metrics", promhttp.Handler())
    logrus.Fatal(http.ListenAndServe(":8080", nil))
}
