package dto

type AdvertisingRequest struct {
	Id      string             `json:"id"`
	Imp     []AdvertisingImp   `json:"imp"`
	Context AdvertisingContext `json:"context"`
}

type AdvertisingImp struct {
	Id        uint `json:"id"`
	MinWidth  uint `json:"minwidth"`
	MinHeight uint `json:"minheight"`
}

type AdvertisingContext struct {
	Ip        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}
