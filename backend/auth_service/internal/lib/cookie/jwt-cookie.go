package cookie

import "os"

var (
	accessSecret = []byte(os.Getenv("accessSecret"))
	refreshSecret = []byte(os.Getenv("refreshSecret"))
)

type Claims struct {
	
}