package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apiutil"
)

type Wrapper struct {
	api          *apigen.ClientWithResponses
	retryMax     int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
}

func NewWrapper(c *Client) *Wrapper {
	return &Wrapper{
		api:          c.API,
		retryMax:     3,
		retryWaitMin: 200 * time.Millisecond,
		retryWaitMax: 2 * time.Second,
	}
}

func NewWrapperFromAPI(api *apigen.ClientWithResponses) *Wrapper {
	return &Wrapper{
		api:          api,
		retryMax:     3,
		retryWaitMin: 200 * time.Millisecond,
		retryWaitMax: 2 * time.Second,
	}
}

func (w *Wrapper) GetItems(ctx context.Context, params *apigen.GetItemsParams) (*apigen.GetItemsResponse, error) {
	for attempt := 0; ; attempt++ {
		resp, err := w.api.GetItemsWithResponse(ctx, params)
		if err != nil {
			if w.shouldRetry(nil, err, "GET") && attempt < w.retryMax {
				w.sleep(nil, attempt)
				continue
			}
			return nil, err
		}
		if w.shouldRetry(resp.HTTPResponse, nil, "GET") && attempt < w.retryMax {
			w.sleep(resp.HTTPResponse, attempt)
			continue
		}
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			return resp, nil
		}
		return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
	}
}

func (w *Wrapper) GetItem(ctx context.Context, id openapi_types.UUID) (*apigen.GetItemResponse, error) {
	for attempt := 0; ; attempt++ {
		resp, err := w.api.GetItemWithResponse(ctx, id)
		if err != nil {
			if w.shouldRetry(nil, err, "GET") && attempt < w.retryMax {
				w.sleep(nil, attempt)
				continue
			}
			return nil, err
		}
		if w.shouldRetry(resp.HTTPResponse, nil, "GET") && attempt < w.retryMax {
			w.sleep(resp.HTTPResponse, attempt)
			continue
		}
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			return resp, nil
		}
		return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
	}
}

func (w *Wrapper) DownloadFile(ctx context.Context, fileID openapi_types.UUID) (*apigen.DownloadFileResponse, error) {
	for attempt := 0; ; attempt++ {
		resp, err := w.api.DownloadFileWithResponse(ctx, fileID)
		if err != nil {
			if w.shouldRetry(nil, err, "GET") && attempt < w.retryMax {
				w.sleep(nil, attempt)
				continue
			}
			return nil, err
		}
		if w.shouldRetry(resp.HTTPResponse, nil, "GET") && attempt < w.retryMax {
			w.sleep(resp.HTTPResponse, attempt)
			continue
		}
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			return resp, nil
		}
		return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
	}
}

func (w *Wrapper) AuthVerifyGet(ctx context.Context) (*apigen.AuthVerifyGetResponse, error) {
	for attempt := 0; ; attempt++ {
		resp, err := w.api.AuthVerifyGetWithResponse(ctx)
		if err != nil {
			if w.shouldRetry(nil, err, "GET") && attempt < w.retryMax {
				w.sleep(nil, attempt)
				continue
			}
			return nil, err
		}
		if w.shouldRetry(resp.HTTPResponse, nil, "GET") && attempt < w.retryMax {
			w.sleep(resp.HTTPResponse, attempt)
			continue
		}
		if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
			return resp, nil
		}
		return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
	}
}

func (w *Wrapper) CreateItem(ctx context.Context, body apigen.CreateItemJSONRequestBody) (*apigen.CreateItemResponse, error) {
	resp, err := w.api.CreateItemWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) UpdateItem(ctx context.Context, id openapi_types.UUID, body apigen.UpdateItemJSONRequestBody) (*apigen.UpdateItemResponse, error) {
	resp, err := w.api.UpdateItemWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) DeleteItem(ctx context.Context, id openapi_types.UUID) (*apigen.DeleteItemResponse, error) {
	resp, err := w.api.DeleteItemWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) PresignFile(ctx context.Context, body apigen.PresignFileJSONRequestBody) (*apigen.PresignFileResponse, error) {
	resp, err := w.api.PresignFileWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) UploadFileWithBody(ctx context.Context, params *apigen.UploadFileParams, contentType string, body io.Reader) (*apigen.UploadFileResponse, error) {
	resp, err := w.api.UploadFileWithBodyWithResponse(ctx, params, contentType, body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) AuthVerifyPost(ctx context.Context) (*apigen.AuthVerifyPostResponse, error) {
	resp, err := w.api.AuthVerifyPostWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) LogoutUser(ctx context.Context) (*apigen.LogoutUserResponse, error) {
	resp, err := w.api.LogoutUserWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return resp, nil
	}
	return nil, w.normalizeError(resp.HTTPResponse, resp.Body)
}

func (w *Wrapper) normalizeError(resp *http.Response, body []byte) error {
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	var serverErr apigen.Error
	var msg string
	if len(body) > 0 && strings.Contains(resp.Header.Get("Content-Type"), "json") {
		if json.Unmarshal(body, &serverErr) == nil {
			if serverErr.Message != nil && *serverErr.Message != "" {
				msg = *serverErr.Message
			}
		}
	}
	if msg == "" {
		if status > 0 {
			msg = http.StatusText(status)
		} else {
			msg = "request failed"
		}
	}
	return apiutil.Error{Status: status, Message: msg}
}

func (w *Wrapper) shouldRetry(resp *http.Response, err error, method string) bool {
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		return false
	}
	if err != nil {
		return true
	}
	if resp == nil {
		return true
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode >= 500 {
		return true
	}
	return false
}

func (w *Wrapper) sleep(resp *http.Response, attempt int) {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if d := parseRetryAfter(ra); d > 0 {
				time.Sleep(d)
				return
			}
		}
	}
	backoff := w.retryWaitMin
	for i := 0; i < attempt; i++ {
		backoff *= 2
		if backoff > w.retryWaitMax {
			backoff = w.retryWaitMax
			break
		}
	}
	time.Sleep(backoff)
}

func parseRetryAfter(v string) time.Duration {
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
