package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"Wrk_Api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestUploadDocument(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "doc_owner@example.com", "Owner")
	projectId := createProjectForSprintTest(r, token)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.Nil(t, err)
	io.WriteString(part, "Hello World Content")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/documents/", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var doc models.Document
	err = json.Unmarshal(w.Body.Bytes(), &doc)
	assert.Nil(t, err)
	assert.Equal(t, "test.txt", doc.Name)
	assert.Equal(t, 1, doc.Version)

	// Clean up
	os.Remove(doc.URL)
}

func TestDocumentVersioning(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "ver_owner@example.com", "Ver Owner")
	projectId := createProjectForSprintTest(r, token)

	// 1. Upload Version 1
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "spec.pdf")
	io.WriteString(part, "V1")
	writer.Close()
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/documents/", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var docV1 models.Document
	json.Unmarshal(w.Body.Bytes(), &docV1)
	defer os.Remove(docV1.URL)

	assert.Equal(t, 1, docV1.Version)

	// 2. Upload Version 2 (linked to V1)
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	part2, _ := writer2.CreateFormFile("file", "spec_v2.pdf")
	io.WriteString(part2, "V2")
	writer2.Close()
	// Add parentId query param
	req2, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/documents/?parentId="+docV1.ID, body2)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)
	var docV2 models.Document
	json.Unmarshal(w2.Body.Bytes(), &docV2)
	defer os.Remove(docV2.URL)

	assert.Equal(t, 2, docV2.Version)
	assert.Equal(t, docV1.ID, *docV2.ParentID)
}

func TestGetDocuments(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "reader@example.com", "Reader")
	projectId := createProjectForSprintTest(r, token)

	// Upload a doc first
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "read.txt")
	io.WriteString(part, "Data")
	writer.Close()
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/documents/", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)
	defer os.Remove(doc.URL)

	// Get List
	req, _ = http.NewRequest("GET", "/api/projects/"+projectId+"/documents/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var docs []models.Document
	json.Unmarshal(w.Body.Bytes(), &docs)
	assert.Len(t, docs, 1)
	assert.Equal(t, "read.txt", docs[0].Name)
}

func TestDeleteDocument(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "del@example.com", "Del")
	projectId := createProjectForSprintTest(r, token)

	// Upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "del.txt")
	io.WriteString(part, "Delete Me")
	writer.Close()
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/documents/", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var doc models.Document
	json.Unmarshal(w.Body.Bytes(), &doc)

	// Verify exists
	_, err := os.Stat(doc.URL)
	assert.Nil(t, err)

	// Delete
	req, _ = http.NewRequest("DELETE", "/api/projects/"+projectId+"/documents/"+doc.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify gone from disk (or at least try)
	_, err = os.Stat(doc.URL)
	assert.True(t, os.IsNotExist(err))
}
