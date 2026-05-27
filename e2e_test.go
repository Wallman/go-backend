package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"go-backend/db"
	"go-backend/mistral"
	"go-backend/otel"
	"go-backend/transcribe"
)

var (
	DB      *db.Queries
	baseURL string
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	shutdown, err := otel.Setup()
	if err != nil {
		log.Fatal("setup otel: ", err)
	}
	defer shutdown()

	postgres, err := testcontainers.Run(
		ctx, "postgres:17",
		testcontainers.WithEnv(map[string]string{"POSTGRES_PASSWORD": "postgres"}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	if err != nil {
		log.Fatal("start postgres: ", err)
	}

	host, _ := postgres.Host(ctx)
	port, _ := postgres.MappedPort(ctx, "5432/tcp")
	os.Setenv("DATABASE_URL", fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port()))

	database, err := db.Connect(ctx)
	if err != nil {
		log.Fatal("connect db: ", err)
	}
	DB = db.New(database)

	mistralStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mistral.TranscribeResponse{
			Text:  "Hello from stub",
			Usage: mistral.Usage{TotalTokens: 42},
		})
	}))

	controller := transcribe.NewController(DB, mistral.NewMistral(mistralStub.URL))
	appServer := httptest.NewServer(NewServer(controller).Handler())
	baseURL = appServer.URL

	code := m.Run()

	appServer.Close()
	mistralStub.Close()
	database.Close()
	postgres.Terminate(ctx)
	os.Exit(code)
}

func createUser(t *testing.T) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	_, err := DB.CreateUser(t.Context(), userID)
	require.NoError(t, err)
	return userID
}

func tokenUsage(t *testing.T, userID uuid.UUID) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, baseURL+"/tokenUsage", nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var count int
	fmt.Sscanf(string(body), "%d", &count)
	return count
}

func TestTranscribe(t *testing.T) {
	userID := createUser(t)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/transcribe", strings.NewReader(`{"uri":"https://example.com/audio.mp3"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello from stub", string(body))
}

func TestTokenUsage(t *testing.T) {
	userID := createUser(t)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/transcribe", strings.NewReader(`{"uri":"https://example.com/audio.mp3"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, 42, tokenUsage(t, userID))
}

func TestMetrics(t *testing.T) {
	userID := createUser(t)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/transcribe", strings.NewReader(`{"uri":"https://example.com/audio.mp3"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	metricsResp, err := http.Get(baseURL + "/metrics")
	require.NoError(t, err)
	defer metricsResp.Body.Close()
	require.Equal(t, http.StatusOK, metricsResp.StatusCode)

	body, err := io.ReadAll(metricsResp.Body)
	require.NoError(t, err)
	metrics := string(body)

	assert.Contains(t, metrics, "http_server_request_duration_seconds", "HTTP server duration metric missing")
	assert.Contains(t, metrics, "http_client_request_duration_seconds", "HTTP client (Mistral) duration metric missing")
	assert.Contains(t, metrics, "db_sql_latency_milliseconds", "DB query duration metric missing")
	assert.Contains(t, metrics, "db_sql_connection_open", "DB connection pool metric missing")
}
