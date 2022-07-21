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

// Stream represents a camera's RTSP stream
type Stream struct {
	Device   string   `json:"device"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Routes   []string `json:"route"`
	Address  string   `json:"address" validate:"required"`
	Port     uint16   `json:"port" validate:"required"`

	CredentialsFound bool `json:"credentials_found"`
	RouteFound       bool `json:"route_found"`
	Available        bool `json:"available"`

	AuthenticationType int `json:"authentication_type"`
}
type CameradarJSON struct {
	Targets []string `json:"targets"`
	Ports   []string `json:"ports"`
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func postCameradar(w http.ResponseWriter, r *http.Request) {
	time_start_request := time.Now().UnixNano()
	// Load data in Mat.gocv
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

	// c.PrintStreams(streams)

	for _, stream := range streams {
		fmt.Println("\nCamera ip address: ", stream.Address, "\nCamera Port: ", stream.Port, "\nCamera routes: ", stream.Route(), "\nCamera username: ",
			stream.Username, "\nCamera password: ", stream.Password)
	}
	log.Println("Full request time: ", float64(time.Now().UnixNano()-time_start_request)/float64(1e9))
	io.WriteString(w, "Cameras_found: ")
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
