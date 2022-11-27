package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage_GetValueByKey(t *testing.T) {
	tests := []struct {
		name          string
		startStorage  Storage
		key           uint
		expectedValue string
	}{
		{
			"one_value",
			Storage{InternalStorage: map[uint]string{1: "aaa"}, NextIndex: 2},
			1,
			"aaa",
		},
		{
			"two_values",
			Storage{InternalStorage: map[uint]string{1: "aaa", 2: "bbb"}, NextIndex: 3},
			2,
			"bbb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultValue, err := tt.startStorage.GetValueByKey(tt.key)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedValue, resultValue)
		})
	}
}

func TestStorage_InsertValue(t *testing.T) {
	tests := []struct {
		name            string
		startStorage    Storage
		value           string
		expectedStorage Storage
	}{
		{
			"empty_storage",
			Storage{InternalStorage: map[uint]string{}, NextIndex: 1},
			"aaa",
			Storage{InternalStorage: map[uint]string{1: "aaa"}, NextIndex: 2},
		},
		{
			"one_value",
			Storage{InternalStorage: map[uint]string{1: "aaa"}, NextIndex: 2},
			"bbb",
			Storage{InternalStorage: map[uint]string{1: "aaa", 2: "bbb"}, NextIndex: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.startStorage.InsertValue(tt.value)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedStorage, tt.startStorage)
		})
	}
}
