package storage

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type Repository interface {
	InsertValue(value string, userID uint) error
	GetValueByKeyAndUserID(key uint, userID uint) (string, error)
	GetNextIndex() (uint, error)
}

type Storage struct {
	InternalStorage map[uint]string
	UserIDToURLID   map[uint][]uint
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
		return &Storage{internalStorage, make(map[uint][]uint), nextInd, nil, nil}, nil
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
				return &Storage{internalStorage, make(map[uint][]uint), nextInd + 1, encoder, decoder}, nil
			}
			internalStorage[mapItem.Key] = mapItem.Value
			nextInd = Max(nextInd, mapItem.Key)
		}
	}
}

func (strg *Storage) GetNextIndex() (uint, error) {
	return strg.NextIndex, nil
}

func (strg *Storage) InsertValue(value string, userID uint) error {
	_, ok := strg.InternalStorage[strg.NextIndex]
	if ok {
		return errors.New("got same key already in storage")
	}
	strg.InternalStorage[strg.NextIndex] = value
	_, ok = strg.UserIDToURLID[userID]
	if !ok {
		strg.UserIDToURLID[userID] = make([]uint, 0)
	}
	strg.UserIDToURLID[userID] = append(strg.UserIDToURLID[userID][:], strg.NextIndex)
	if strg.Encoder != nil {
		mapItem := MapItem{Key: strg.NextIndex, Value: value}
		if err := strg.Encoder.Encode(mapItem); err != nil {
			return err
		}
	}
	strg.NextIndex++
	return nil
}

func Contains(list []uint, value uint) error {
	for _, v := range list {
		if v == value {
			return nil
		}
	}
	return errors.New("No value " + strconv.Itoa(int(value)) + " in list")
}

func (strg *Storage) GetValueByKeyAndUserID(key uint, userID uint) (string, error) {
	//userURLs, ok := strg.UserIDToURLID[userID]
	//if !ok {
	//	return "", errors.New("got userID not presented in storage")
	//}
	//if err := Contains(userURLs, key); err != nil {
	//	return "", err
	//}
	_, ok := strg.InternalStorage[key]
	if !ok {
		return "", errors.New("got key not presented in storage")
	}
	return strg.InternalStorage[key], nil
}
