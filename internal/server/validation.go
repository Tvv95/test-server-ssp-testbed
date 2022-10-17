package server

import (
	"errors"
	"test_task/internal/dto"
)

const (
	WrongSchema = "WRONG_SCHEMA"
	EmptyTiles  = "EMPTY_TILES"
	EmptyField  = "EMPTY_FIELD"
)

func validateRequest(placementRequest *dto.PlacementRequest) error {
	if placementRequest.Id == nil || placementRequest.Context.Ip == nil || placementRequest.Context.UserAgent == nil {
		return errors.New(WrongSchema)
	}
	if len(placementRequest.Tiles) == 0 {
		return errors.New(EmptyTiles)
	}
	for _, tile := range placementRequest.Tiles {
		if tile.Id == nil || tile.Width == nil || tile.Ratio == nil {
			return errors.New(WrongSchema)
		}
		if *tile.Id == 0 || *tile.Width == 0 || *tile.Ratio == 0 {
			return errors.New(EmptyField)
		}
	}
	if *placementRequest.Id == "" || *placementRequest.Context.Ip == "" || *placementRequest.Context.UserAgent == "" {
		return errors.New(EmptyField)
	}
	return nil
}
