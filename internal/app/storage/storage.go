package storage

import "errors"

type Repository interface {
	InsertValue(value string) error
	GetValueByKey(key uint) (string, error)
}

type Storage struct {
	InternalStorage map[uint]string
	NextIndex       uint
}

func (strg *Storage) InsertValue(value string) error {
	_, ok := strg.InternalStorage[strg.NextIndex]
	if ok {
		return errors.New("got same key already in storage")
	}
	strg.InternalStorage[strg.NextIndex] = value
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
