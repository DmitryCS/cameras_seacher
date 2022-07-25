package main

import (
	"cameras_seacher/config"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Ullaakut/cameradar/v5"
)

func getStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /status request\n")
	io.WriteString(w, "Status check passed!\n")
}

type CameradarJSON struct {
	Targets []string `json:"targets"`
	Ports   []string `json:"ports"`
}

type CameradarResponseJSON struct {
	Ips       []string `json:"ips"`
	Ports     []string `json:"ports"`
	Routes    []string `json:"routes"`
	Usernames []string `json:"usernames"`
	Passwords []string `json:"passwords"`
}
type CameraR struct {
	Ip       string   `json:"ip"`
	Port     string   `json:"port"`
	Routes   []string `json:"routes"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}
type CameraResponseJSON struct {
	Cameras []CameraR `json:"cameras"`
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func postCameradar(w http.ResponseWriter, r *http.Request) {
	time_start_request := time.Now().UnixNano()
	decoder := json.NewDecoder(r.Body)
	var data_json CameradarJSON
	err := decoder.Decode(&data_json)
	if err != nil {
		panic(err)
	}

	c, err := cameradar.New(
		cameradar.WithTargets(data_json.Targets),
		cameradar.WithPorts(data_json.Ports),
		cameradar.WithDebug(false),
		cameradar.WithVerbose(false),
		cameradar.WithCustomCredentials("./dictionaries/credentials.json"),
		cameradar.WithCustomRoutes("./dictionaries/routes"),
		cameradar.WithScanSpeed(4),
		cameradar.WithAttackInterval(0),
		cameradar.WithTimeout(2000*time.Millisecond),
	)
	if err != nil {
		fmt.Println(err)
	}

	scanResult, err := c.Scan()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(scanResult)

	streams, err := c.Attack(scanResult)
	if err != nil {
		fmt.Println(err)
	}
	var p CameraResponseJSON
	for _, stream := range streams {
		cam_t := CameraR{Ip: stream.Address,
			Port:     strconv.Itoa(int(stream.Port)),
			Routes:   stream.Routes,
			Username: stream.Username,
			Password: stream.Password}
		p.Cameras = append(p.Cameras, cam_t)
		// fmt.Println("\nCamera ip address: ", stream.Address, "\nCamera Port: ", stream.Port, "\nCamera routes: ", stream.Routes, "\nCamera username: ",
		// 	stream.Username, "\nCamera password: ", stream.Password)
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(p)
	log.Println("Full request time: ", float64(time.Now().UnixNano()-time_start_request)/float64(1e9))
}

func main() {
	fmt.Println(config.HttpServerConfig.URL)
	http.HandleFunc("/status", getStatus)
	http.HandleFunc("/cameradar", postCameradar)
	err := http.ListenAndServe(":"+strconv.Itoa(config.HttpServerConfig.PORT), nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
