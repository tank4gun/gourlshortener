package db

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestCreateDb(t *testing.T) {
	type args struct {
		dbDSN string
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDb(tt.args.dbDSN)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDb() got = %v, want %v", got, tt.want)
			}
		})
	}
}
