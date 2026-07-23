package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/subipraNuvem/url-shortener/internal/service"
)

type HTTPErrorWithFeedback struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Hint       string `json:"hint"`
}

type errorMapping struct {
	match    func(error) bool
	feedback HTTPErrorWithFeedback
}

var errorMappings = []errorMapping{
	{
		match: func(err error) bool { return errors.Is(err, service.ErrNotFound) },
		feedback: HTTPErrorWithFeedback{
			StatusCode: http.StatusNotFound,
			Message:    "URL not found",
			Hint:       "Check if the code is correct or if the URL has been deactivated",
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, service.ErrInactive) },
		feedback: HTTPErrorWithFeedback{
			StatusCode: http.StatusGone,
			Message:    "URL is no longer active",
			Hint:       "This URL has been deactivated and can no longer be used",
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, service.ErrAliasTaken) },
		feedback: HTTPErrorWithFeedback{
			StatusCode: http.StatusConflict,
			Message:    "Alias already taken",
			Hint:       "Choose a different alias or omit it to have one generated automatically",
		},
	},
	{
		match: func(err error) bool { return strings.Contains(err.Error(), "required") },
		feedback: HTTPErrorWithFeedback{
			StatusCode: http.StatusBadRequest,
			Message:    "Missing required field",
			Hint:       "Check the request body and ensure all required fields are present",
		},
	},
	{
		match: func(err error) bool {
			msg := err.Error()
			return strings.Contains(msg, "url") || strings.Contains(msg, "URL")
		},
		feedback: HTTPErrorWithFeedback{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid URL format",
			Hint:       "Provide a valid URL including the scheme (e.g. https://example.com)",
		},
	},
}

var defaultFeedback = HTTPErrorWithFeedback{
	StatusCode: http.StatusInternalServerError,
	Message:    "An unexpected error occurred",
	Hint:       "Please try again later",
}

func GetHTTPErrorWithFeedbackByError(err error) HTTPErrorWithFeedback {
	for _, mapping := range errorMappings {
		if mapping.match(err) {
			return mapping.feedback
		}
	}
	return defaultFeedback
}
