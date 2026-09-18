package http

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleResponse(ctx context.Context, w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func HandleResponseWithBody(ctx context.Context, w http.ResponseWriter, statusCode int, data any) {
	if data != nil {
		responseBody, err := json.Marshal(data)
		if err != nil {
			HandleError(ctx, w, err)
			return
		}

		w.WriteHeader(statusCode)
		// let's assume for now that it will always be JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseBody)
	} else {
		HandleResponse(ctx, w, statusCode)
	}
}

func HandlePostResponse(ctx context.Context, w http.ResponseWriter, statusCode int, data any, location string) {
	if data != nil {
		responseBody, err := json.Marshal(data)
		if err != nil {
			HandleError(ctx, w, err)
			return
		}

		w.WriteHeader(statusCode)
		w.Header().Set("Location", location)
		// let's assume for now that it will always be JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseBody)
	} else {
		HandleResponse(ctx, w, statusCode)
	}
}
