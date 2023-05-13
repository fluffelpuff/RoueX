package utils

import "fmt"

func FormatSize(size uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	var i int
	for i = 0; size >= 1024 && i < len(units)-1; i++ {
		size /= 1024
	}
	return fmt.Sprintf("%d %s", size, units[i])
}
