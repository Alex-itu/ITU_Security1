package main

import (
  "crypto/tls"
  "crypto/x509"
  "io"
  "log"
  "net/http"
  "os"
  "time"
)

var client_port = ""

const (
  // by changing the the ending, can you make it posiable to reach different request handlers
  // It is just a  different endpoint.
  url = "https://localhost:8443"
)

func clientSetup() (*http.Server, error) {
  cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
  if err != nil {
    log.Fatalf("Failed to load X509 key pair: %v", err)
  }

  config := &tls.Config{
    Certificates: []tls.Certificate{cert},
  }

  router := http.NewServeMux()

  // different endpoint does different things
  router.HandleFunc("/", handleRequest)

  hospital := &http.Server{
    Addr:      client_port,
    Handler:   router,
    TLSConfig: config,
  }

  return hospital, err
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Hey from client with port:" + client_port))
}


func main() {
  go connection()
  
  client_port = ":" + "8445"


  // Keep the main function alive
  for {}
}

func connectionSetup() (*http.Client, error) {
  cert, err := os.ReadFile("ca.crt")
  if err != nil {
    return nil, err
  }

  caCertPool := x509.NewCertPool()
  caCertPool.AppendCertsFromPEM(cert)

  tlsConfig := &tls.Config{
    RootCAs: caCertPool,
  }

  tr := &http.Transport{
    TLSClientConfig: tlsConfig,
  }

  client := &http.Client{Transport: tr}

  return client, nil
}

func connection() {
  client, err := connectionSetup()
  if err != nil {
    log.Fatalf("Failed to create connection setup: %v", err)
  }


  // Get them GETs 
  for {
    time.Sleep(1 * time.Second) // Retry after a delay 
   
    resp, err := client.Get(url)
    if err != nil {
      log.Printf("Failed to get response: %v", err)
      continue
    }

    responseHandler(resp)
  }
}

func responseHandler(resp *http.Response) {
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    log.Printf("Failed to read response body: %v", err)
    return
  }

  log.Printf("Hospital Response: %s\n", body)
}