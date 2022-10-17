package dto

type AdvertisingResponse struct {
	Id  string          `json:"id"`
	Imp []AdResponseImp `json:"imp"`
}

type AdResponseImp struct {
	Id     uint    `json:"id"`
	Width  uint    `json:"width"`
	Height uint    `json:"height"`
	Title  string  `json:"title"`
	Url    string  `json:"url"`
	Price  float64 `json:"price"`
}
