package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/subipraNuvem/url-shortener/internal/api/handler"
	"github.com/subipraNuvem/url-shortener/internal/service"
)

func TestGetHTTPErrorWithFeedbackByError(t *testing.T) {
	t.Run("ErrNotFound maps to 404", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(service.ErrNotFound)
		require.Equal(t, http.StatusNotFound, httpErr.StatusCode)
	})

	t.Run("wrapped ErrNotFound maps to 404", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(fmt.Errorf("resolve: %w", service.ErrNotFound))
		require.Equal(t, http.StatusNotFound, httpErr.StatusCode)
	})

	t.Run("ErrInactive maps to 410 Gone", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(service.ErrInactive)
		require.Equal(t, http.StatusGone, httpErr.StatusCode)
	})

	t.Run("ErrAliasTaken maps to 409 Conflict", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(service.ErrAliasTaken)
		require.Equal(t, http.StatusConflict, httpErr.StatusCode)
	})

	t.Run("error containing 'required' maps to 400", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(errors.New("field long_url is required"))
		require.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("error containing 'url' maps to 400", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(errors.New("invalid url format"))
		require.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("error containing 'URL' maps to 400", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(errors.New("URL scheme missing"))
		require.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("unknown error maps to 500", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(errors.New("something unexpected happened"))
		require.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	})

	t.Run("response body has message and hint", func(t *testing.T) {
		httpErr := handler.GetHTTPErrorWithFeedbackByError(service.ErrNotFound)
		require.NotEmpty(t, httpErr.Message)
		require.NotEmpty(t, httpErr.Hint)
	})
}
