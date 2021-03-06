package main

import (
	"flag"
	// "github.com/pkg/profile"
	"log"
	"os/exec"
	"runtime"
	"runtime/debug"
	"time"
)

func main() {
	flag.Parse()
	loadInitialData()
	go heatServer()
	runtime.GC()
	debug.SetGCPercent(-1)
	startWebServer()
}

// func main() {
// 	flag.Parse()
// 	// defer profile.Start(profile.CPUProfile).Stop()
// 	// defer profile.Start(profile.MutexProfile).Stop()
// 	// defer profile.Start(profile.BlockProfile).Stop()
// 	// defer profile.Start(profile.MemProfile).Stop()
// 	loadInitialData()
// 	time.Sleep(3 * time.Second)
// 	defer profile.Start(profile.MemProfile).Stop()
// 	startWebServer()
// }

var (
	ADDR         = flag.String("addr", ":80", "TCP address to listen to")
	DATAZIP_PATH = flag.String("zip", "/tmp/data/data.zip", "Zipfile path")
	DATA_DIR     = flag.String("data", "/", "Directory with extacted jsons")
	OPTIONS_PATH = flag.String("options", "/tmp/data/options.txt", "options file path")

	Users           = makeUsersRepo()
	Locations       = makeLocationsRepo()
	Visits          = makeVisitsRepo()
	UsersVisits     = makeEntityVisitsRepo(USERS_REPO_COLLECTION_SIZE)
	LocationsVisits = makeEntityVisitsRepo(LOCATIONS_REPO_COLLECTION_SIZE)

	InitialTime time.Time
)

func entity_repo(entity_kind_len int) EntityRepo {
	switch entity_kind_len {
	case 5: //"users":
		return &Users
	case 9: //"locations":
		return &Locations
	case 6: //"visits":
		return &Visits
	}
	return nil
}

func heatServer() {
	time.Sleep(5 * time.Second)
	log.Println("Starting heater")
	cmd := exec.Command("./heater", "--addr", *ADDR)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second)
}
