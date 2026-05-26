package main

import (
	"context"
	"database/sql"
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
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"go-backend/db"
	"go-backend/mistral"
	"go-backend/transcribe"
	"go-backend/user"
)

var (
	DB      *sql.DB
	baseURL string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

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

	DB, err = db.Connect(ctx)
	if err != nil {
		log.Fatal("connect db: ", err)
	}

	mistralStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mistral.TranscribeResponse{
			Text:  "Hello from stub",
			Usage: mistral.Usage{TotalTokens: 42},
		})
	}))

	controller := transcribe.NewController(user.NewUserRepository(DB), mistral.NewMistral(mistralStub.URL))
	appServer := httptest.NewServer(NewServer(controller).Handler())
	baseURL = appServer.URL

	code := m.Run()

	appServer.Close()
	mistralStub.Close()
	DB.Close()
	postgres.Terminate(ctx)
	os.Exit(code)
}

func createUser(t *testing.T) string {
	t.Helper()
	userID := uuid.New().String()
	_, err := DB.ExecContext(context.Background(), "INSERT INTO users (id, tokens_used) VALUES ($1, 0)", userID)
	require.NoError(t, err)
	return userID
}

func tokenUsage(t *testing.T, userID string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, baseURL+"/tokenUsage", nil)
	req.Header.Set("X-User-ID", userID)
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
	req.Header.Set("X-User-ID", userID)

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
	req.Header.Set("X-User-ID", userID)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, 42, tokenUsage(t, userID))
}
