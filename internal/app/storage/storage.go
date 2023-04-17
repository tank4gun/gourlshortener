package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// FullInfoURLResponse - response object for shortened URL with original one
type FullInfoURLResponse struct {
	ShortURL    string `json:"short_url"`    // ShortURL - result shorten URL
	OriginalURL string `json:"original_url"` // OriginalURL - original URL
}

// StatsResponse - response object for GetStatsHandler method
type StatsResponse struct {
	URLs  int `json:"urls"`  // URLs - total URLs amount in database
	Users int `json:"users"` // Users - total users amount in database
}

// AllPossibleChars - chars for shorten URL creation
var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// IRepository interface for usage as storage
type IRepository interface {
	InsertValue(value string, userID uint) error                                 // InsertValue - insert value for userID into IRepository
	GetValueByKeyAndUserID(key uint, userID uint) (string, int)                  // GetValueByKeyAndUserID - get value by key and userID from IRepository
	GetNextIndex() (uint, error)                                                 // GetNextIndex - get next index for insertion into IRepository
	GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int) // GetAllURLsByUserID - get all URLs by userID from IRepository
	InsertBatchValues(values []string, startIndex uint, userID uint) error       // InsertBatchValues - insert values batch for userID into IRepository
	MarkBatchAsDeleted(IDs []uint, userID uint) error                            // MarkBatchAsDeleted - set deleted=true for rows by its IDs and userID in IRepository
	GetStats() (response StatsResponse, errCode int)                             // GetStats - get stats from database
	Ping() error                                                                 // Ping - check that connection to IRepository is alive
	Shutdown() error                                                             // Shutdown - gracefully shotdown IRepository
}

// ExistError - error type for existing ID in Repository
type ExistError struct {
	ID  uint   // ID that already exists
	Err string // Err - string for error description
}

// Error - implementation Error method for ExisError struct
func (err *ExistError) Error() string {
	return fmt.Sprintf("%s, id = %v", err.Err, err.ID)
}

// URL - base struct with Value and deletion mark
type URL struct {
	Value   string // Value - URL value
	Deleted bool   // Deleted - true if URL is marked as deleted
}

// Storage - struct for file storage
type Storage struct {
	InternalStorage map[uint]URL    // InternalStorage - URLID map to URL struct
	UserIDToURLID   map[uint][]uint // UserIDToURLID - relationships between UserID and URLID
	NextIndex       uint            // NextIndex - next index to insert
	Encoder         *json.Encoder   // Encoder - object to encode URLs
	Decoder         *json.Decoder   // Decoder - object to decode encoded URLs
}

// DBStorage - struct for database storage
type DBStorage struct {
	db *sql.DB // db - sql.DB pointer
}

// Shutdown - close db connection for DBStorage
func (strg *DBStorage) Shutdown() error {
	return strg.db.Close()
}

// Shutdown - in case Storage do nothing
func (strg *Storage) Shutdown() error {
	return nil
}

// CreateShortURL - get short URL from its ID
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

// MapItem - struct for Storage getting-URLs usage
type MapItem struct {
	Key   uint   // Key - key for URL
	Value string // Value - value for URL
}

// Max - get max value from two uints
func Max(x, y uint) uint {
	if x < y {
		return y
	}
	return x
}

// NewStorage - create Storage instance with given parameters
func NewStorage(internalStorage map[uint]URL, nextInd uint, filename string, dbDSN string) (IRepository, error) {
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
		internalStorage := make(map[uint]URL)
		decoder := json.NewDecoder(file)
		encoder := json.NewEncoder(file)
		nextInd := uint(0)
		for {
			var mapItem MapItem
			err := decoder.Decode(&mapItem)
			if err != nil {
				return &Storage{internalStorage, make(map[uint][]uint), nextInd + 1, encoder, decoder}, nil
			}
			internalStorage[mapItem.Key] = URL{mapItem.Value, false}
			nextInd = Max(nextInd, mapItem.Key)
		}
	}
}

// Ping - check that connection to Storage is alive
func (strg *Storage) Ping() error {
	return nil
}

// GetAllURLsByUserID - get all URLs by userID from Storage
func (strg *Storage) GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int) {
	userURLs, ok := strg.UserIDToURLID[userID]
	if !ok {
		return nil, http.StatusNoContent
	}
	responseList := make([]FullInfoURLResponse, 0)
	// URLID - URL ID
	for _, URLID := range userURLs {
		shortURL := CreateShortURL(URLID)
		shortURL = baseURL + shortURL
		originalURL, ok := strg.InternalStorage[URLID]
		if !ok {
			return nil, http.StatusInternalServerError
		}
		responseList = append(responseList, FullInfoURLResponse{ShortURL: shortURL, OriginalURL: originalURL.Value})
	}
	return responseList, 200
}

// GetNextIndex - get next index for insertion into Storage
func (strg *Storage) GetNextIndex() (uint, error) {
	return strg.NextIndex, nil
}

// InsertValue - insert value for userID into IRepository
func (strg *Storage) InsertValue(value string, userID uint) error {
	_, ok := strg.InternalStorage[strg.NextIndex]
	if ok {
		return errors.New("got same key already in storage")
	}
	for i := uint(0); i < strg.NextIndex; i++ {
		URLval, ok := strg.InternalStorage[strg.NextIndex]
		if ok && URLval.Value == value {
			log.Printf("Got same URL in storage %s", value)
			return &ExistError{ID: i, Err: "Got same URL in storage"}
		}
	}
	strg.InternalStorage[strg.NextIndex] = URL{value, false}
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

// GetValueByKeyAndUserID - get value by key and userID from IRepository
func (strg *Storage) GetValueByKeyAndUserID(key uint, userID uint) (string, int) {
	value, ok := strg.InternalStorage[key]
	if !ok {
		log.Printf("got key %d not presented in storage", key)
		return "", http.StatusBadRequest
	}
	if value.Deleted {
		return "", http.StatusGone
	}
	return value.Value, 0
}

// MarkBatchAsDeleted - set deleted=true for rows by its IDs and userID in IRepository
func (strg *Storage) MarkBatchAsDeleted(IDs []uint, userID uint) error {
	userURLs, ok := strg.UserIDToURLID[userID]
	if !ok {
		return errors.New("couldn't get userURLs")
	}
	for _, ID := range IDs {
		for _, userURLID := range userURLs {
			if ID == userURLID {
				value := strg.InternalStorage[ID]
				value.Deleted = true
				strg.InternalStorage[ID] = value
			}
		}
	}
	return nil
}

// InsertBatchValues - insert values batch for userID into IRepository
func (strg *Storage) InsertBatchValues(values []string, startIndex uint, userID uint) error {
	for index, value := range values {
		indexToInsert := startIndex + uint(index)
		_, ok := strg.InternalStorage[indexToInsert]
		if ok {
			return &ExistError{indexToInsert, "Got used index"}
		}
		strg.InternalStorage[indexToInsert] = URL{value, false}
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
		strg.NextIndex++
	}
	return nil
}

// GetStats - get stats from database
func (strg *Storage) GetStats() (response StatsResponse, errCode int) {
	URLsCount := int(strg.NextIndex) - 1
	UsersCount := len(strg.UserIDToURLID)
	return StatsResponse{URLs: URLsCount, Users: UsersCount}, 200
}

// GetNextIndex - get next index for insertion into DBStorage
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

// InsertValue - insert value for userID into DBStorage
func (strg *DBStorage) InsertValue(value string, userID uint) error {
	// URLID - URL ID
	var URLID uint
	row := strg.db.QueryRow("SELECT id from url where value = $1", value)
	err := row.Scan(&URLID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return err
	}
	if err == nil {
		return &ExistError{uint(URLID), "Got existing URL"}
	}
	log.Printf("Insert value %s into url table", value)
	row = strg.db.QueryRow("INSERT INTO url (value) values ($1) returning id", value)
	err = row.Scan(&URLID)
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

// GetValueByKeyAndUserID - get value by key and userID from DBStorage
func (strg *DBStorage) GetValueByKeyAndUserID(key uint, userID uint) (string, int) {
	row := strg.db.QueryRow("SELECT value, deleted from url where id = $1", key)
	var value string
	var deleted bool
	err := row.Scan(&value, &deleted)
	if err != nil {
		log.Printf("got key %d not presented in storage", key)
		return "", http.StatusBadRequest
	}
	if deleted {
		return "", http.StatusGone
	}
	return value, 0
}

// GetAllURLsByUserID - get all URLs by userID from DBStorage
func (strg *DBStorage) GetAllURLsByUserID(userID uint, baseURL string) ([]FullInfoURLResponse, int) {
	userURLs := make([]uint, 0)
	rows, err := strg.db.Query("SELECT url_id from user_url where user_id = $1", userID)
	if err != nil {
		return nil, http.StatusNoContent
	}
	defer rows.Close()
	for rows.Next() {
		// URLID - URL ID
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
	// URLID - URL ID
	for _, URLID := range userURLs {
		shortURL := CreateShortURL(URLID)
		shortURL = baseURL + shortURL
		originalURL, errCode := strg.GetValueByKeyAndUserID(URLID, userID)
		if errCode != 0 {
			return nil, http.StatusInternalServerError
		}
		responseList = append(responseList, FullInfoURLResponse{ShortURL: shortURL, OriginalURL: originalURL})
	}
	return responseList, 200
}

// Ping - check that connection to DBStorage is alive
func (strg *DBStorage) Ping() error {
	err := strg.db.Ping()
	return err
}

// InsertBatchValues - insert values batch for userID into DBStorage
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

// MarkBatchAsDeleted - set deleted=true for rows by its IDs and userID in DBStorage
func (strg *DBStorage) MarkBatchAsDeleted(IDs []uint, userID uint) error {
	tx, err := strg.db.Begin()
	if err != nil {
		return err
	}
	log.Printf("Delete urls %v for user_id %d", IDs, userID)
	updateStmt, err := tx.Prepare(
		"UPDATE url SET deleted = true WHERE id IN (SELECT url_id FROM user_url where user_id = ($1) AND url_id = ANY($2::integer[]))",
	)
	if err != nil {
		return err
	}
	defer updateStmt.Close()
	if _, err := updateStmt.Exec(userID, IDs); err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			log.Printf("Update stmt failed, %s", err1.Error())
			return err1
		}
		log.Printf("Error %s", err.Error())
		return err
	}
	if err := tx.Commit(); err != nil {
		log.Printf("Unable to commit: %s", err.Error())
		return err
	}
	return nil
}

// GetStats - get stats from database
func (strg *DBStorage) GetStats() (response StatsResponse, errCode int) {
	row := strg.db.QueryRow("SELECT count(*) from url where deleted = false")
	var URLsCount int
	var UsersCount int
	err := row.Scan(&URLsCount)
	if err != nil {
		log.Print("Couldn't get URLs count")
		return StatsResponse{}, http.StatusBadRequest
	}
	row = strg.db.QueryRow("SELECT count(distinct user_id) from user_url")
	err = row.Scan(&UsersCount)
	if err != nil {
		log.Print("Couldn't get Users count")
		return StatsResponse{}, http.StatusBadRequest
	}
	return StatsResponse{URLs: URLsCount, Users: UsersCount}, 200
}
