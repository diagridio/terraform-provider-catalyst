package region

import "os"

type Region string

const (
	DefaultRegion = Region("onebox")
)

func (r Region) ID() string {
	return string(r)
}

func (r Region) Name() string {
	switch r {
	case "onebox":
		return "Onebox Host #1"
	}

	return "Unknown"
}

func GetEnvOrDefault(key string, def Region) Region {
	r, ok := os.LookupEnv(key)
	if ok {
		return Region(r)
	}

	return def
}
