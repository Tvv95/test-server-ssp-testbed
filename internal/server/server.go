package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"test_task/internal/dto"
	"time"
)

var wg sync.WaitGroup

type server struct {
	port      string
	adsIpPort []string
	router    *mux.Router
}

func NewServer(port int, adsIpPort []string) *server {
	s := &server{
		port:      fmt.Sprintf(":%d", port),
		adsIpPort: adsIpPort,
		router:    mux.NewRouter(),
	}
	s.configureRouter()
	return s
}

func (s *server) Start() error {
	log.Println("Server started")
	return http.ListenAndServe(s.port, s.router)
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/placements/request", s.handlePlacementsRequest()).Methods("POST")
}

func (s *server) handlePlacementsRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		placementRequest := &dto.PlacementRequest{}
		if err := json.Unmarshal(data, placementRequest); err != nil {
			log.Println(err)
			return
		}
		if err := validateRequest(placementRequest); err != nil {
			log.Println(err)
			s.errorRespond(w, http.StatusBadRequest)
			return
		}

		advertisingRequest := buildRequestToAdServices(placementRequest)

		ch := make(chan dto.AdvertisingResponse)

		s.postToAdServices(ch, advertisingRequest)

		allImps := make([]dto.AdResponseImp, 0)
		for el := range ch {
			allImps = append(allImps, el.Imp...)
		}
		if len(allImps) == 0 {
			s.errorRespond(w, http.StatusNoContent)
			return
		}

		placementResponse := buildResponse(allImps, *placementRequest.Id)

		s.respond(w, http.StatusCreated, placementResponse)
	}
}

func buildRequestToAdServices(placementRequest *dto.PlacementRequest) *dto.AdvertisingRequest {
	advertisingRequest := &dto.AdvertisingRequest{}
	advertisingRequest.Id = *placementRequest.Id
	imps := make([]dto.AdvertisingImp, 0, len(placementRequest.Tiles))
	for _, tile := range placementRequest.Tiles {
		imps = append(imps, dto.AdvertisingImp{
			Id:        *tile.Id,
			MinWidth:  *tile.Width,
			MinHeight: uint(math.Floor(float64(*tile.Width) * *tile.Ratio)),
		})
	}
	advertisingRequest.Imp = imps
	advertisingRequest.Context = dto.AdvertisingContext{
		Ip:        *placementRequest.Context.Ip,
		UserAgent: *placementRequest.Context.UserAgent,
	}
	return advertisingRequest
}

func (s *server) postToAdServices(ch chan dto.AdvertisingResponse, advertisingRequest *dto.AdvertisingRequest) {
	for _, url := range s.adsIpPort {
		wg.Add(1)
		url = fmt.Sprintf("http://%s/bid_request", url)
		go postAdRequest(url, ch, advertisingRequest)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
}

func postAdRequest(url string, ch chan<- dto.AdvertisingResponse, body *dto.AdvertisingRequest) {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Millisecond * 200,
	}
	resp, _ := client.Do(req)
	defer wg.Done()
	defer resp.Body.Close()
	jsonData, _ := io.ReadAll(resp.Body)
	data := dto.AdvertisingResponse{}
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Println(err)
	}
	ch <- data
}

func buildResponse(allImps []dto.AdResponseImp, placementRequestId string) *dto.PlacementResponse {
	impIdToImp := make(map[uint]dto.AdResponseImp, len(allImps))

	for _, v := range allImps {
		val, ok := impIdToImp[v.Id]
		if !ok || ok && val.Price < v.Price {
			impIdToImp[v.Id] = v
		}
	}
	placementImps := make([]dto.PlacementImp, 0, len(impIdToImp))
	for _, v := range impIdToImp {
		placementImps = append(placementImps, dto.PlacementImp{
			Id:     v.Id,
			Width:  v.Width,
			Height: v.Height,
			Title:  v.Title,
			Url:    v.Url,
		})
	}
	return &dto.PlacementResponse{
		Id:  placementRequestId,
		Imp: placementImps,
	}
}

func (s *server) respond(w http.ResponseWriter, code int, placementResponse *dto.PlacementResponse) {
	w.WriteHeader(code)
	jsonData, err := json.Marshal(placementResponse)
	if err != nil {
		log.Println(err)
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Println(err)
	}
}

func (s *server) errorRespond(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
