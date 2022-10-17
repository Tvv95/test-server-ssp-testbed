package dto

type PlacementResponse struct {
	Id  string         `json:"id"`
	Imp []PlacementImp `json:"imp"`
}

type PlacementImp struct {
	Id     uint   `json:"id"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Title  string `json:"title"`
	Url    string `json:"url"`
}
