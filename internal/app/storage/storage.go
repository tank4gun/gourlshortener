package storage

import (
	"encoding/json"
	"errors"
	"os"
)

type Repository interface {
	InsertValue(value string) error
	GetValueByKey(key uint) (string, error)
	GetNextIndex() (uint, error)
}

type Storage struct {
	InternalStorage map[uint]string
	NextIndex       uint
	Encoder         *json.Encoder
	Decoder         *json.Decoder
}

type MapItem struct {
	Key   uint
	Value string
}

func Max(x, y uint) uint {
	if x < y {
		return y
	}
	return x
}

func NewStorage(internalStorage map[uint]string, nextInd uint, filename string) (*Storage, error) {
	if filename == "" {
		return &Storage{internalStorage, nextInd, nil, nil}, nil
	} else {
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return nil, err
		}
		internalStorage := make(map[uint]string)
		decoder := json.NewDecoder(file)
		encoder := json.NewEncoder(file)
		nextInd := uint(0)
		for {
			var mapItem MapItem
			err := decoder.Decode(&mapItem)
			if err != nil {
				return &Storage{internalStorage, nextInd, encoder, decoder}, nil
			}
			internalStorage[mapItem.Key] = mapItem.Value
			nextInd = Max(nextInd, mapItem.Key)
		}
	}
}

func (strg *Storage) GetNextIndex() (uint, error) {
	return strg.NextIndex, nil
}

func (strg *Storage) InsertValue(value string) error {
	_, ok := strg.InternalStorage[strg.NextIndex]
	if ok {
		return errors.New("got same key already in storage")
	}
	strg.InternalStorage[strg.NextIndex] = value
	mapItem := MapItem{Key: strg.NextIndex, Value: value}
	if err := strg.Encoder.Encode(mapItem); err != nil {
		return err
	}
	strg.NextIndex++
	return nil
}

func (strg *Storage) GetValueByKey(key uint) (string, error) {
	_, ok := strg.InternalStorage[key]
	if !ok {
		return "", errors.New("got key not presented in storage")
	}
	return strg.InternalStorage[key], nil
}
