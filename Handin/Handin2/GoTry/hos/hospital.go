package main

import (
  "crypto/tls"
  "log"
  "net/http"
)

const (
  hos_port = ":8443"
  responseBody = "Thanks for joining the protocol, please send your port"
  maxClients = 3
)

var numClients = 0

func hospitalSetup() (*http.Server, error) {
  cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
  if err != nil {
    log.Fatalf("Failed to load X509 key pair: %v", err)
  }

  config := &tls.Config{
    Certificates: []tls.Certificate{cert},
  }

  router := http.NewServeMux()
  router.HandleFunc("/", connectionEstablished)
  router.HandleFunc("/j", handleRequest2)

  hospital := &http.Server{
    Addr:      hos_port,
    Handler:   router,
    TLSConfig: config,
  }

  return hospital, err
}

func main() {
  log.Printf("Listening on %s...", hos_port)

  // Does the Setup for starting a server
  hospital, err := hospitalSetup()

  // This makes sure to keep listing for requests
  err = hospital.ListenAndServeTLS("", "")
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }

}

func connectionEstablished(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(responseBody))
}

func handleRequest2(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("yoyo"))
}