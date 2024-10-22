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
	// "strconv"
	//"time"
)

type OtherClientPorts struct {
	Ports []string
}

type clientInfo struct {
	Port string
}
 var client_port = ""

const (
  // by changing the the ending, can you make it possible to reach different request handlers
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

  // these are all "listning" for request
  // different endpoint does different things
  router.HandleFunc("/", connectionEstablished)
  router.HandleFunc("/GetClientsPorts", hospitalPostsAllPort)
  router.HandleFunc("/SendShares", connectionEstablished)
  

  clientServer := &http.Server{
    Addr:      ":" + client_port,
    Handler:   router,
    TLSConfig: config,
  }

  return clientServer, err
}

func connectionEstablished(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Hey from client with port:" + client_port))
}

func hospitalPostsAllPort(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  ports := readClientPorts(r.Body)
  log.Printf("Got Ports: %v", ports)
}

func readClientPorts(body io.ReadCloser) ([]string) {
  // not sure why, but this has to &clientInfo else I non-pointer error in unmarshaling
  ports := &OtherClientPorts{}

  bodyBytes, err := io.ReadAll(body)
  if err != nil {
    log.Printf("Failed to read response body: %v", err)
  }

  // Takes the content of the request and puts it in clientIn
  err = json.Unmarshal(bodyBytes, ports)
  if err != nil {
    log.Printf("Failed to Unmarshal: %v", err)
  }
  return ports.Ports
}



func main() {
  port := os.Args[1:]

  if cap(port) != 1 {
    client_port = "8445"
  } else {
    client_port = port[0]
  }
  println("port: " + port[0])
  
  // Does the Setup for starting a server
  clientServer, err := clientSetup()
  if err != nil {
    panic("ClientServer failed")
  }
  println("ClientServer up and running")

  // This makes sure to keep listing for requests
  go listenAndServe(clientServer)
  

  
  
  // Starts a connection to the hospital to see if can response 
  clientCon, status := connection()
  if status != 200 {
    panic("Connection went wrong")
  }
  println(status)

  // with the hospital alive and running send client's port to it
  postPortToHospital(clientCon)
  


  // Keep the main function alive
  for {}
}

func listenAndServe(clientServer *http.Server) {
  err := clientServer.ListenAndServeTLS("", "")
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }
}

func connectionSetup() (*http.Client, error) {
  // Pretty much just checks if this file exist
  cert, err := os.ReadFile("ca.crt")
  if err != nil {
    return nil, err
  }

  // Add certs to your "key chain"
  caCertPool := x509.NewCertPool()
  caCertPool.AppendCertsFromPEM(cert)

  // TLS Configuration
  // Simple adds cert into tls (Not quiet right, just for understanding)
  tlsConfig := &tls.Config{
    RootCAs: caCertPool,
  }

  // now with the cert in the tls, we simply say that in a http request, use this tls config 
  tr := &http.Transport{
    TLSClientConfig: tlsConfig,
  }

  // A client now has the given tls config to use in request
  client := &http.Client{Transport: tr}

  return client, nil
}

func connection() (*http.Client, int) {
  // setup meaning, what the client is using in as configs in http request  
  client, err := connectionSetup()
  if err != nil {
    log.Fatalf("Failed to create connection setup: %v", err)
  }

  // Simple http get request, just to make sure that the hospital is running
  resp, err := client.Get(url)
  if err != nil {
    log.Printf("Failed to get response: %v", err)
  }

  responseHandler(resp)
  return client, resp.StatusCode
}

// As the name says, gives the client's port to hospital
func postPortToHospital(client *http.Client) {
  clientIn := clientInfo {
    Port: client_port,
  }
  log.Printf("port: %v", clientIn)

  // Sadly, this is needed, you cant post without a body...
  bodyBytes, err := json.Marshal(&clientIn)
  if err != nil {
    log.Fatal(err)
  }
  bodyReader := bytes.NewReader(bodyBytes)

  // This post request gives the current client's port to the hospital
  resp, err := client.Post((url + "/ClientPortPost"), "string", bodyReader)
  
  responseHandler(resp)
}

// maybe it will be a fits all function... hopefully 
func responseHandler(resp *http.Response) {
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    log.Printf("Failed to read response body: %v", err)
    return
  }

  log.Printf("Response: %s\n", body)
}