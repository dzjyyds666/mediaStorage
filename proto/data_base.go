package proto

var DatabaseName = struct {
	FileDataBaseName  string
	BoxDataBaseName   string
	DepotDataBaseName string
}{
	FileDataBaseName:  "files_db",
	BoxDataBaseName:   "boxes_db",
	DepotDataBaseName: "depots_db",
}
