package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"testing"
)

func TestCreateServer(t *testing.T) {
	tests := []struct {
		name         string
		startStorage storage.Storage
	}{
		{
			"server_created",
			storage.Storage{InternalStorage: map[uint]string{}, NextIndex: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdServer := CreateServer(&tt.startStorage, nil)
			assert.NotNil(t, createdServer)
		})
	}
}
