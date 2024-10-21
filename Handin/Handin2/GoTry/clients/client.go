package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)


type clientInfo struct {
	Port int
}
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
    Addr:      ":" + client_port,
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
  
  client_port = "8445"
  
  // Does the Setup for starting a server
  client, err := clientSetup()
  
  
  // Starts a connection to the hospital to see if can response 
  connection()


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


  
  for {
    time.Sleep(1 * time.Second) // Retry after a delay 
   
    resp, err := client.Get(url)
    if err != nil {
      log.Printf("Failed to get response: %v", err)
      continue
    }

    responseHandler(resp)

    postPortToHospital(client)
   
  }
}

func postPortToHospital(client *http.Client) {
  i, err := strconv.Atoi(client_port)
  if err != nil {
      // ... handle error
      panic(err)
  }

  clientin := clientInfo {
    Port: i,
  }

  bodyBytes, err := json.Marshal(&clientin)
  if err != nil {
    log.Fatal(err)
  }

  bodyReader := bytes.NewReader(bodyBytes)
  resp1, err := client.Post((url + "/j"), "string", bodyReader)
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