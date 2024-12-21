package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type Config struct {
	MainBackendURL string
	ApiKey         string
	Endpoint       string
}

type ReminderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func loadConfig() Config {
	godotenv.Load(".env")
	// In production, use environment variables or AWS Parameter Store
	return Config{
		MainBackendURL: os.Getenv("LOCAL_BACKEND_URL"),
		ApiKey:         os.Getenv("API_KEY"),
		Endpoint:       "/api/v1/reminders/process",
	}
}

func triggerReminders(config Config) error {
	/*
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
	*/

	/*
		// Create request body if needed
		payload := map[string]interface{}{
			"trigger_time": time.Now().UTC().Format(time.RFC3339),
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("error marshaling payload: %v", err)
		}
	*/

	/*
		// Create request
		url := config.MainBackendURL + config.Endpoint
		req, err := http.NewRequest("GET", url, nil) // POST bytes.NewBuffer(jsonPayload)
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}

		// Add headers
		//req.Header.Set("Content-Type", "application/json")
		//req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.ApiKey))
		//req.Header.Set("X-Service-Name", "reminder-trigger")

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error executing request: %v", err)
		}
		defer resp.Body.Close()
	*/

	res, err := http.Get("http://127.0.0.1:5001/")
	fmt.Println(err)

	/*
		// Parse response
		var result ReminderResponse
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
	*/

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	fmt.Println("response:", res)

	message := string(body[:])

	result := ReminderResponse{Message: message}

	// Check response status
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, message: %s", res.StatusCode, result.Message)
	}

	log.Printf("Successfully triggered reminders: %s", result.Message)
	return nil
}

func setupHealthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Start HTTP server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	config := loadConfig()

	// Validate config
	if config.MainBackendURL == "" || config.ApiKey == "" {
		log.Fatal("Missing environment variables")
	}

	// Setup health check endpoint
	setupHealthCheck()

	/*
		// Create new cron scheduler
		c := cron.New(cron.WithLocation(time.UTC),
			cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))
	*/

	// Seconds field, optional
	c := cron.New(cron.WithParser(cron.NewParser(
		cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
	)), cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))

	// Schedule job to run at midnight UTC
	//_, err := c.AddFunc("0 0 * * *", func() {
	_, err := c.AddFunc("* * * * *", func() {
		log.Println("Starting reminder trigger job")
		if err := triggerReminders(config); err != nil {
			log.Printf("Error triggering reminders: %v", err)
			// In production, you might want to add alerts here
		}
	})

	if err != nil {
		log.Fatalf("Error scheduling job: %v", err)
	}

	// Start the scheduler
	c.Start()

	// Keep the service running
	select {}
}
