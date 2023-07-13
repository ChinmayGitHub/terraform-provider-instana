package testutils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	rand "math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const contentTypeHeaderName = "Content-Type"
const healthPath = "/health"

// EchoHandlerFunc is a handler function for the TestHTTPServer which echos the request
// with a http status code 200
func EchoHandlerFunc(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		r.Write(bytes.NewBufferString("Failed to get request"))
	} else {
		w.Header().Set(contentTypeHeaderName, r.Header.Get(contentTypeHeaderName))
		w.WriteHeader(http.StatusOK)
		w.Write(requestBody)
	}
}

// healthFunc is a handler function which is registered on path /health to check if server
// is running
func healthFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	r.Write(bytes.NewBufferString("OK"))
}

// NewTestHTTPServer create and starts a new TestHTTPServer on random port
func NewTestHTTPServer() TestHTTPServer {
	router := mux.NewRouter()
	router.HandleFunc(healthPath, healthFunc)
	return &testHTTPServerImpl{
		router:      router,
		callCounter: make(map[string]int),
	}
}

// MinPortNumber the minimum port number used by the http test server
const MinPortNumber = 50000

// MaxPortNumber the maximum port number used by the http test server
const MaxPortNumber = 59000

// TestHTTPServer simple helper to mock an http server for testing.
type TestHTTPServer interface {
	GetPort() int
	GetCallCount(method string, path string) int
	AddRoute(method string, path string, handlerFunc http.HandlerFunc)
	Start()
	Close()
	WriteInternalServerError(w http.ResponseWriter, err error)
	WriteJSONResponse(w http.ResponseWriter, jsonData []byte)
}

type testHTTPServerImpl struct {
	router      *mux.Router
	port        *int
	httpServer  *http.Server
	callCounter map[string]int
}

// GetPort returns the dynamic server port
func (server *testHTTPServerImpl) GetPort() int {
	if server.port == nil {
		port := server.randomFreePort()
		server.port = &port
	}
	return *server.port
}

// GetCallCount returns the call counter for the given method and path
func (server *testHTTPServerImpl) GetCallCount(method string, path string) int {
	key := method + "_" + path
	val, ok := server.callCounter[key]
	if !ok {
		return 0
	}
	return val
}

// AddRoute adds a new route. Routes can only be added before the server was started
func (server *testHTTPServerImpl) AddRoute(method string, path string, handlerFunc http.HandlerFunc) {
	server.router.HandleFunc(path, server.wrapHandlerFunc(handlerFunc)).Methods(method)
}

func (server *testHTTPServerImpl) wrapHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + "_" + r.URL.Path
		val, ok := server.callCounter[key]
		if !ok {
			val = 0
		}
		server.callCounter[key] = val + 1
		handlerFunc(w, r)
	}
}

// Start starts the http service with the configured routes
func (server *testHTTPServerImpl) Start() {
	binding := fmt.Sprintf(":%d", server.GetPort())
	srv := &http.Server{
		Addr:    binding,
		Handler: server.router,
	}
	go func() {
		rootFolder, err := GetRootFolder()
		if err != nil {
			log.Fatalf("Failed to get root folder of project: %s", err)
			return
		}
		certFile := fmt.Sprintf("%s/testutils/test-server.pem", rootFolder)
		keyFile := fmt.Sprintf("%s/testutils/test-server.key", rootFolder)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
			log.Fatalf("Failed to start http server using binding %s: %s", binding, err)
		}

	}()
	server.httpServer = srv

	server.waitForServerAlive()
}

// RandomPort creates a random port between 50000 and 59000
func (server *testHTTPServerImpl) randomFreePort() int {
	maxAttempts := 5
	attempt := 0
	randomPort := server.randomPort()
	for attempt < maxAttempts && server.isPortInUse(randomPort) {
		attempt++
		randomPort = server.randomPort()
	}
	return randomPort
}

func (server *testHTTPServerImpl) randomPort() int {
	source := cryptoSource{}
	random := rand.New(source)
	return random.Intn(MaxPortNumber-MinPortNumber) + MinPortNumber
}

func (server *testHTTPServerImpl) isPortInUse(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("failed to bind port %d; %s", port, err)
		return false
	}
	err = l.Close()
	if err != nil {
		log.Fatalf("Failed to close listener for port  %d; %s", port, err)
		return false
	}
	return true
}

func (server *testHTTPServerImpl) waitForServerAlive() {
	url := fmt.Sprintf("https://localhost:%d/health", server.GetPort())

	for i := 0; i < 5; i++ {
		if resp, err := http.Get(url); err == nil && resp.StatusCode == 200 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Close stops the http listener
func (server *testHTTPServerImpl) Close() {
	if server.httpServer != nil {
		server.httpServer.Close()
	}
}

// WriteInternalServerError Writes the provided error message as a response message and sets status code 501 - Internal Server Error with content type text/plain
func (server *testHTTPServerImpl) WriteInternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set(contentTypeHeaderName, "text/plain; charset=utf-8")
	w.Write([]byte(err.Error()))
}

// WriteJSONResponse Writes the provided data with content type application/json and status code 200 OK to the ResponseWriter
func (server *testHTTPServerImpl) WriteJSONResponse(w http.ResponseWriter, jsonData []byte) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(contentTypeHeaderName, "application/json; charset=utf-8")
	w.Write(jsonData)
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
