package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAddMember(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	tokenOwner, _ := GetAuthToken(r, "owner_add@example.com", "Owner")
	projectId := createProjectForSprintTest(r, tokenOwner)

	// Create another user to add
	GetAuthToken(r, "new_member@example.com", "New User")

	memberReq := handlers.AddMemberRequest{
		Email: "new_member@example.com",
		Role:  "DEVELOPER",
	}
	jsonValue, _ := json.Marshal(memberReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/members/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+tokenOwner)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var member models.ProjectMember
	err := json.Unmarshal(w.Body.Bytes(), &member)
	assert.Nil(t, err)
	assert.Equal(t, "DEVELOPER", member.Role)
	assert.Equal(t, projectId, member.ProjectID)
	assert.NotEmpty(t, member.UserID)
}

func TestAddMemberAccessControl(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	tokenOwner, _ := GetAuthToken(r, "owner_ac@example.com", "Owner")
	tokenOther, _ := GetAuthToken(r, "other_ac@example.com", "Other")
	projectId := createProjectForSprintTest(r, tokenOwner)

	// Create user to add
	GetAuthToken(r, "target@example.com", "Target")

	memberReq := handlers.AddMemberRequest{
		Email: "target@example.com",
	}
	jsonValue, _ := json.Marshal(memberReq)

	// Other user tries to add member
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/members/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+tokenOther)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetMembers(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	tokenOwner, _ := GetAuthToken(r, "owner_get@example.com", "Owner")
	projectId := createProjectForSprintTest(r, tokenOwner)

	req, _ := http.NewRequest("GET", "/api/projects/"+projectId+"/members/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenOwner)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var members []models.ProjectMember
	json.Unmarshal(w.Body.Bytes(), &members)
	assert.Len(t, members, 1) // Only owner initially
	assert.Equal(t, "OWNER", members[0].Role)
}

func TestRemoveMember(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	tokenOwner, _ := GetAuthToken(r, "owner_rem@example.com", "Owner")
	projectId := createProjectForSprintTest(r, tokenOwner)

	// Add member first
	GetAuthToken(r, "removable@example.com", "Removable")
	memberReq := handlers.AddMemberRequest{Email: "removable@example.com"}
	jsonValue, _ := json.Marshal(memberReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/members/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+tokenOwner)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var member models.ProjectMember
	json.Unmarshal(w.Body.Bytes(), &member)
	memberId := member.ID

	// Remove Member
	req, _ = http.NewRequest("DELETE", "/api/projects/"+projectId+"/members/"+memberId, nil)
	req.Header.Set("Authorization", "Bearer "+tokenOwner)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify removal
	req, _ = http.NewRequest("GET", "/api/projects/"+projectId+"/members/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenOwner)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var members []models.ProjectMember
	json.Unmarshal(w.Body.Bytes(), &members)
	assert.Len(t, members, 1) // Back to just owner
}
