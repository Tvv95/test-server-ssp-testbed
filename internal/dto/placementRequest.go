package dto

type PlacementRequest struct {
	Id      *string          `json:"id"`
	Tiles   []PlacementTile  `json:"tiles"`
	Context PlacementContext `json:"context"`
}

type PlacementTile struct {
	Id    *uint    `json:"id"`
	Width *uint    `json:"width"`
	Ratio *float64 `json:"ratio"`
}

type PlacementContext struct {
	Ip        *string `validate:"ipv4" json:"ip"`
	UserAgent *string `json:"user_agent"`
}
