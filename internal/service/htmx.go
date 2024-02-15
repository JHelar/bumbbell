package service

import "net/http"

type HtmxService struct {
}

func NewHtmxService() *HtmxService {
	return &HtmxService{}
}

func (s *HtmxService) IsHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (s *HtmxService) IsHtmxHistoryRequest(r *http.Request) bool {
	return r.Header.Get("HX-History-Request") == "true"
}
