package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage_GetValueByKeyAndUserID(t *testing.T) {
	tests := []struct {
		name          string
		startStorage  Storage
		key           uint
		expectedValue string
	}{
		{
			"one_value",
			Storage{
				InternalStorage: map[uint]URL{1: URL{"aaa", false}}, UserIDToURLID: map[uint][]uint{1: {1}}, NextIndex: 2, Encoder: nil, Decoder: nil,
			},
			1,
			"aaa",
		},
		{
			"two_values",
			Storage{
				InternalStorage: map[uint]URL{1: URL{"aaa", false}, 2: URL{"bbb", false}}, UserIDToURLID: map[uint][]uint{1: {2}}, NextIndex: 3, Encoder: nil, Decoder: nil,
			},
			2,
			"bbb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultValue, err := tt.startStorage.GetValueByKeyAndUserID(tt.key, 1)
			assert.Equal(t, err, 0)
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
			Storage{
				InternalStorage: map[uint]URL{}, UserIDToURLID: make(map[uint][]uint), NextIndex: 1, Encoder: nil, Decoder: nil,
			},
			"aaa",
			Storage{
				InternalStorage: map[uint]URL{1: URL{"aaa", false}}, UserIDToURLID: map[uint][]uint{1: {1}}, NextIndex: 2, Encoder: nil, Decoder: nil,
			},
		},
		{
			"one_value",
			Storage{
				InternalStorage: map[uint]URL{1: URL{"aaa", false}}, UserIDToURLID: map[uint][]uint{1: {1}}, NextIndex: 2, Encoder: nil, Decoder: nil,
			},
			"bbb",
			Storage{
				InternalStorage: map[uint]URL{1: URL{"aaa", false}, 2: URL{"bbb", false}}, UserIDToURLID: map[uint][]uint{1: {1, 2}}, NextIndex: 3, Encoder: nil, Decoder: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.startStorage.InsertValue(tt.value, 1)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedStorage, tt.startStorage)
		})
	}
}

func TestStorage_GetNextIndex(t *testing.T) {
	tests := []struct {
		name            string
		storage         Storage
		expectedNextInd uint
	}{
		{
			"init_next_index",
			Storage{
				InternalStorage: map[uint]URL{}, UserIDToURLID: make(map[uint][]uint), NextIndex: 1, Encoder: nil, Decoder: nil,
			},
			1,
		},
		{
			"10th_next_index",
			Storage{
				InternalStorage: map[uint]URL{
					1: URL{"a", false}, 2: URL{"b", false}, 3: URL{"c", false}, 4: URL{"aa", false}, 5: URL{"r", false}, 6: URL{"1", false}, 7: URL{"qwe", false}, 8: URL{"d", false}, 9: URL{"tt", false},
				},
				UserIDToURLID: make(map[uint][]uint),
				NextIndex:     10, Encoder: nil, Decoder: nil,
			},
			10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultNextIndex, err := tt.storage.GetNextIndex()
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedNextInd, resultNextIndex)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name   string
		x      uint
		y      uint
		result uint
	}{
		{
			"left_bigger",
			10,
			5,
			10,
		},
		{
			"right_bigger",
			10,
			15,
			15,
		},
		{
			"equal",
			10,
			10,
			10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxResult := Max(tt.x, tt.y)
			assert.Equal(t, tt.result, maxResult)
		})
	}
}
