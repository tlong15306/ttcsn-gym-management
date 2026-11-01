package catalog

type Filter struct {
	AvailableAt string
	Query       string
	Status      string
	Date        string
	RoomID      int64
	Page        int
	PageSize    int
}
