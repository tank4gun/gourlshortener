package server

import (
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tank4gun/gourlshortener/internal/app/storage"
)

func TestCreateServer(t *testing.T) {
	tests := []struct {
		name         string
		startStorage storage.Storage
	}{
		{
			"server_created",
			storage.Storage{InternalStorage: map[uint]storage.URL{}, NextIndex: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdServer := CreateServer(&tt.startStorage, make(chan handlers.RequestToDelete, 10))
			assert.NotNil(t, createdServer)
		})
	}
}

func TestGenerateNewID(t *testing.T) {
	t.Run("same_size", func(t *testing.T) {
		assert.Equal(t, len(GenerateNewID()), 4)
	})
}
