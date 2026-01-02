package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/cache"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
)

type ItemsService struct {
	w   *api.Wrapper
	c   *cache.Manager
	cfg config.Config
}

func NewItemsService(w *api.Wrapper, c *cache.Manager, cfg config.Config) *ItemsService {
	return &ItemsService{w: w, c: c, cfg: cfg}
}

func (s *ItemsService) List(ctx context.Context, params *apigen.GetItemsParams) (*apigen.GetItemsResponse, error) {
	key := s.keyForList(params)
	resp, err := s.w.GetItems(ctx, params)
	if err == nil {
		if s.cfg.Cache.Enabled {
			pj := resp.Body
			_ = s.c.Put(key, pj, nil, "")
		}
		return resp, nil
	}
	if !s.cfg.Cache.Enabled {
		return nil, err
	}
	if !isNetworkError(err) {
		return nil, err
	}
	pj, _, ts, _, gerr := s.c.Get(key)
	if gerr != nil || !s.c.IsFresh(ts) {
		return nil, err
	}
	var parsed apigen.GetItemsResponse
	parsed.Body = pj
	parsed.HTTPResponse = nil
	return &parsed, nil
}

func (s *ItemsService) Get(ctx context.Context, id openapi_types.UUID) (*apigen.GetItemResponse, error) {
	key := s.keyForGet(id)
	resp, err := s.w.GetItem(ctx, id)
	if err == nil {
		if s.cfg.Cache.Enabled {
			pj := resp.Body
			_ = s.c.Put(key, pj, nil, "")
		}
		return resp, nil
	}
	if !s.cfg.Cache.Enabled {
		return nil, err
	}
	if !isNetworkError(err) {
		return nil, err
	}
	pj, _, ts, _, gerr := s.c.Get(key)
	if gerr != nil || !s.c.IsFresh(ts) {
		return nil, err
	}
	var parsed apigen.GetItemResponse
	parsed.Body = pj
	parsed.HTTPResponse = nil
	return &parsed, nil
}

func (s *ItemsService) keyForList(params *apigen.GetItemsParams) string {
	t := time.Now().UnixNano()
	_ = t
	if params == nil {
		return "items:list:"
	}
	var ty, se string
	if params.Type != nil {
		ty = string(*params.Type)
	}
	if params.S != nil {
		se = *params.S
	}
	lim, off := 0, 0
	if params.Limit != nil {
		lim = *params.Limit
	}
	if params.Offset != nil {
		off = *params.Offset
	}
	return fmt.Sprintf("items:list:type=%s;s=%s;limit=%d;offset=%d", ty, se, lim, off)
}

func (s *ItemsService) keyForGet(id openapi_types.UUID) string {
	return fmt.Sprintf("items:get:%s", id.String())
}

func isNetworkError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func (s *ItemsService) Create(ctx context.Context, body apigen.ItemCreate) (*apigen.CreateItemResponse, error) {
	resp, err := s.w.CreateItem(ctx, body)
	if err != nil {
		return nil, err
	}
	_ = s.c.DeletePrefix("items:list:")
	return resp, nil
}

func (s *ItemsService) Update(ctx context.Context, id openapi_types.UUID, body apigen.UpdateItemJSONRequestBody) (*apigen.UpdateItemResponse, error) {
	resp, err := s.w.UpdateItem(ctx, id, body)
	if err != nil {
		return nil, err
	}
	_ = s.c.DeletePrefix("items:list:")
	_ = s.c.Delete(s.keyForGet(id))
	return resp, nil
}

func (s *ItemsService) Delete(ctx context.Context, id openapi_types.UUID) (*apigen.DeleteItemResponse, error) {
	resp, err := s.w.DeleteItem(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.c.DeletePrefix("items:list:")
	_ = s.c.Delete(s.keyForGet(id))
	return resp, nil
}
