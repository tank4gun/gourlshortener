package storage

type Storage struct {
	InternalStorage map[uint]string
	NextIndex       uint
}
