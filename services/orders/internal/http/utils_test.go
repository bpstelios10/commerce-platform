package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleResponse_Succeeds(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	HandleResponse(ctx, w, 204)

	assert.NotNil(t, w.Code)
	assert.Equal(t, 204, w.Code)
	assert.Empty(t, w.Header())
	assert.NotNil(t, w.Body)
	assert.Empty(t, w.Body.Bytes())
}

func TestHandleResponseWithBody_WhenBodyExists_ReturnsJson(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	body := make(map[string]int)
	body["key1"] = 1
	body["key2"] = 2
	HandleResponseWithBody(ctx, w, 200, body)

	assert.NotNil(t, w.Code)
	assert.Equal(t, 200, w.Code)
	assert.Len(t, w.Header(), 1)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))
	assert.NotNil(t, w.Body)
	assert.NotEmpty(t, w.Body.Bytes())
	assert.JSONEq(t, `{"key1":1,"key2":2}`, w.Body.String())
}

func TestHandleResponseWithBody_WhenBodyNotExists_ReturnsCode(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	var body any = nil
	HandleResponseWithBody(ctx, w, 200, body)

	assert.NotNil(t, w.Code)
	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header())
	assert.NotNil(t, w.Body)
	assert.Empty(t, w.Body.Bytes())
}

func TestHandleResponseWithBody_WhenBodyNotJson_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	body := make(chan int) // channels cannot be JSON marshaled
	HandleResponseWithBody(ctx, w, 200, body)

	assert.NotNil(t, w.Code)
	assert.Equal(t, 500, w.Code)
	assert.Len(t, w.Header(), 1)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))
	assert.NotNil(t, w.Body)
	assert.NotEmpty(t, w.Body.Bytes())
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR", "message":"internal server error"}`, w.Body.String())
}

func TestHandlePostResponse_WhenBodyExists_ReturnsJson(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	body := make(map[string]int)
	body["key1"] = 1
	body["key2"] = 2
	HandlePostResponse(ctx, w, 200, body, "some/location")

	assert.NotNil(t, w.Code)
	assert.Equal(t, 200, w.Code)
	assert.Len(t, w.Header(), 2)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))
	assert.Equal(t, "some/location", w.Header().Get("location"))
	assert.NotNil(t, w.Body)
	assert.NotEmpty(t, w.Body.Bytes())
	assert.JSONEq(t, `{"key1":1,"key2":2}`, w.Body.String())
}

func TestHandlePostResponse_WhenBodyNotExists_ReturnsCode(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	var body any = nil
	HandlePostResponse(ctx, w, 200, body, "some/location")

	assert.NotNil(t, w.Code)
	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header())
	assert.NotNil(t, w.Body)
	assert.Empty(t, w.Body.Bytes())
}

func TestHandlePostResponse_WhenBodyNotJson_ReturnsError(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()

	body := make(chan int) // channels cannot be JSON marshaled
	HandlePostResponse(ctx, w, 200, body, "some/location")

	assert.NotNil(t, w.Code)
	assert.Equal(t, 500, w.Code)
	assert.Len(t, w.Header(), 1)
	assert.Equal(t, "application/json", w.Header().Get("content-type"))
	assert.NotNil(t, w.Body)
	assert.NotEmpty(t, w.Body.Bytes())
	assert.JSONEq(t, `{"code":"INTERNAL_SERVER_ERROR", "message":"internal server error"}`, w.Body.String())
}
