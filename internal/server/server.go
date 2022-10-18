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
	srv := &http.Server{
		Addr:         s.port,
		Handler:      s.router,
		WriteTimeout: 250 * time.Millisecond,
	}
	log.Println("Server started")
	return srv.ListenAndServe()
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/placements/request", s.handlePlacementsRequest()).Methods("POST")
}

func (s *server) handlePlacementsRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		placementRequest := &dto.PlacementRequest{}
		if err := json.Unmarshal(data, placementRequest); err != nil {
			log.Println(WrongSchema)
			s.errorRespond(w, http.StatusBadRequest)
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

		placementResponse := buildResponse(allImps, placementRequest)

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
	defer wg.Done()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println(err)
		return
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Millisecond * 200,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return
	}
	jsonData, _ := io.ReadAll(resp.Body)
	data := dto.AdvertisingResponse{}
	unmarshalErr := json.Unmarshal(jsonData, &data)
	if unmarshalErr != nil {
		return
	}
	ch <- data
}

func buildResponse(allImps []dto.AdResponseImp, placementRequest *dto.PlacementRequest) *dto.PlacementResponse {
	impIdToImp := make(map[uint]dto.AdResponseImp, len(allImps))

	for _, v := range allImps {
		if cur, ok := impIdToImp[v.Id]; !ok || ok && cur.Price < v.Price {
			impIdToImp[v.Id] = v
		}
	}
	placementImps := make([]dto.PlacementImp, 0, len(impIdToImp))
	for _, v := range placementRequest.Tiles {
		if _, ok := impIdToImp[*v.Id]; ok {
			placementImps = append(placementImps, dto.PlacementImp{
				Id:     impIdToImp[*v.Id].Id,
				Width:  impIdToImp[*v.Id].Width,
				Height: impIdToImp[*v.Id].Height,
				Title:  impIdToImp[*v.Id].Title,
				Url:    impIdToImp[*v.Id].Url,
			})
		}
	}
	return &dto.PlacementResponse{
		Id:  *placementRequest.Id,
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
