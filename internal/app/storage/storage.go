package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type FullInfoURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Repository interface {
	InsertValue(value string, userID uint) error
	GetValueByKeyAndUserID(key uint, userID uint) (string, error)
	GetNextIndex() (uint, error)
	GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int)
	InsertBatchValues(values []string, startIndex uint, userID uint) error
	Ping() error
}

type Storage struct {
	InternalStorage map[uint]string
	UserIDToURLID   map[uint][]uint
	NextIndex       uint
	Encoder         *json.Encoder
	Decoder         *json.Decoder
}

type DBStorage struct {
	db *sql.DB
}

func CreateShortURL(currInd uint) string {
	var sb strings.Builder
	for {
		if currInd == 0 {
			break
		}
		sb.WriteByte(AllPossibleChars[currInd%62])
		currInd = currInd / 62
	}
	return sb.String()
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

func NewStorage(internalStorage map[uint]string, nextInd uint, filename string, dbDSN string) (Repository, error) {
	if dbDSN != "" {
		database, err := sql.Open("pgx", dbDSN)
		if err != nil {
			return nil, err
		}
		return &DBStorage{database}, nil
	}
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

func (strg *Storage) Ping() error {
	return nil
}

func (strg *Storage) GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int) {
	userURLs, ok := strg.UserIDToURLID[userID]
	if !ok {
		return nil, http.StatusNoContent
	}
	responseList := make([]FullInfoURLResponse, 0)
	for _, URLID := range userURLs {
		shortURL := CreateShortURL(URLID)
		shortURL = baseURL + shortURL
		originalURL, ok := strg.InternalStorage[URLID]
		if !ok {
			return nil, http.StatusInternalServerError
		}
		responseList = append(responseList, FullInfoURLResponse{ShortURL: shortURL, OriginalURL: originalURL})
	}
	return responseList, 200
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
	_, ok := strg.InternalStorage[key]
	if !ok {
		return "", errors.New("got key not presented in storage")
	}
	return strg.InternalStorage[key], nil
}

func (strg *Storage) InsertBatchValues(values []string, startIndex uint, userID uint) error {
	for index, value := range values {
		indexToInsert := startIndex + uint(index)
		_, ok := strg.InternalStorage[indexToInsert]
		if ok {
			return errors.New("got same key already in storage")
		}
		strg.InternalStorage[indexToInsert] = value
		_, ok = strg.UserIDToURLID[userID]
		if !ok {
			strg.UserIDToURLID[userID] = make([]uint, 0)
		}
		strg.UserIDToURLID[userID] = append(strg.UserIDToURLID[userID][:], indexToInsert)
		if strg.Encoder != nil {
			mapItem := MapItem{Key: indexToInsert, Value: value}
			if err := strg.Encoder.Encode(mapItem); err != nil {
				return err
			}
		}
	}
	return nil
}

func (strg *DBStorage) GetNextIndex() (uint, error) {
	row := strg.db.QueryRow("Select last_value from url_id_seq")
	var currInd sql.NullInt64
	err := row.Scan(&currInd)
	if err != nil {
		return 0, err
	}
	if currInd.Valid {
		val := currInd.Int64
		return uint(val) + 1, nil
	} else {
		return 1, nil
	}
}

func (strg *DBStorage) InsertValue(value string, userID uint) error {
	var URLID uint
	row := strg.db.QueryRow("INSERT INTO url (value) values ($1) returning id", value)
	err := row.Scan(&URLID)
	if err != nil {
		log.Fatal(err)
		return err
	}
	row = strg.db.QueryRow("INSERT INTO user_url (user_id, url_id) values ($1, $2)", userID, URLID)
	err = row.Scan()
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
		return err
	}
	return nil
}

func (strg *DBStorage) GetValueByKeyAndUserID(key uint, userID uint) (string, error) {
	row := strg.db.QueryRow("SELECT value from url where id = $1", key)
	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", errors.New("got key not presented in storage")
	}
	return value, nil
}

func (strg *DBStorage) GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int) {
	userURLs := make([]uint, 0)
	rows, err := strg.db.Query("SELECT url_id from user_url where user_id = $1", userID)
	if err != nil {
		return nil, http.StatusNoContent
	}
	defer rows.Close()
	for rows.Next() {
		var URLID uint
		err = rows.Scan(&URLID)
		if err != nil {
			return nil, http.StatusNoContent
		}
		userURLs = append(userURLs, URLID)
	}
	err = rows.Err()
	if err != nil {
		return nil, http.StatusNoContent
	}

	responseList := make([]FullInfoURLResponse, 0)
	for _, URLID := range userURLs {
		shortURL := CreateShortURL(URLID)
		shortURL = baseURL + shortURL
		originalURL, err := strg.GetValueByKeyAndUserID(URLID, userID)
		if err != nil {
			return nil, http.StatusInternalServerError
		}
		responseList = append(responseList, FullInfoURLResponse{ShortURL: shortURL, OriginalURL: originalURL})
	}
	return responseList, 200
}

func (strg *DBStorage) Ping() error {
	err := strg.db.Ping()
	return err
}

func (strg *DBStorage) InsertBatchValues(values []string, startIndex uint, userID uint) error {
	tx, err := strg.db.Begin()
	if err != nil {
		return err
	}
	URLstmt, err := tx.Prepare("INSERT INTO url (value) VALUES ($1)")
	if err != nil {
		return err
	}
	UserURLstmt, err := tx.Prepare("INSERT INTO user_url (user_id, url_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer URLstmt.Close()
	defer UserURLstmt.Close()
	for index, value := range values {
		if _, err := URLstmt.Exec(value); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("Insert to url, need rollback, %v", err)
				return err
			}
			return err
		}
		if _, err = UserURLstmt.Exec(userID, startIndex+uint(index)); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("Insert to user_url, need rollback, %v", err)
				return err
			}
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("Unable to commit: %v", err)
		return err
	}
	return nil
}
