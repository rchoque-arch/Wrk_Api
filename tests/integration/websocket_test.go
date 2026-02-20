package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/realtime"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketTaskUpdates(t *testing.T) {
	// Start Hub logic manually for tests or ensure it's running.
	// Since GlobalHub is a global var created via NewHub(), channels are init.
	// But h.Run() needs to be running.
	go realtime.GlobalHub.Run()

	SetupTestDB()
	r := SetupRouter()

	// Use httptest Server because standard ws client needs a TCP listener (ws://)
	// simple httptest.NewRecorder() doesn't handle hijacking for WS well.
	s := httptest.NewServer(r)
	defer s.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/api/ws"

	token, _ := GetAuthToken(r, "ws_user@example.com", "WS User")
	projectId := createProjectForSprintTest(r, token)

	// Connect WS Client
	// Header: Authorization: Bearer <token>
	// Query: projectId
	wsURL = wsURL + "?projectId=" + projectId
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token)

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	assert.Nil(t, err, "Should connect to websocket")
	defer ws.Close()

	// Perform REST Action (Create Task)
	taskReq := handlers.CreateTaskRequest{
		Title: "Realtime Task",
	}
	jsonValue, _ := json.Marshal(taskReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	// We need to execute this against the SERVER handler, or just use the router manually?
	// If we use s.URL, we are making real network call.
	// Let's use the router directly for the REST part to keep it fast, but s for WS.
	// Wait, REST handler uses GlobalHub. GlobalHub is shared. So it should work.

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Read Message from WS
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := ws.ReadMessage()
	assert.Nil(t, err, "Should receive message")

	var msg realtime.Message
	err = json.Unmarshal(message, &msg)
	assert.Nil(t, err)

	assert.Equal(t, "TASK_CREATED", msg.Type)
	assert.Equal(t, projectId, msg.ProjectID)

	payloadMap := msg.Payload.(map[string]interface{})
	assert.Equal(t, "Realtime Task", payloadMap["title"])
}
