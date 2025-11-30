package DCDUtils

import (
	"os"

	"github.com/ricochet2200/go-disk-usage/du"
)

// Get file size and exclude unreadable files
func SizeStore(path string, size int64) int64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return -1
	}

	return size + int64(len(data))
}

// Amount of free space on the partition the program is running off of (in bytes)
func FreeStorage(path string) int64 {
	usage := du.NewDiskUsage(path)
	free := usage.Free()

	var spareStorage int64 = int64(free)
	return spareStorage
}
