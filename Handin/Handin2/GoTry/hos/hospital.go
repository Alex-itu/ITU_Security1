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
  "sync"
	"strconv"
)



const (
  hos_port = "8443"
  maxClients = 3
)

type clientInfo struct {
	Port string
}

type OtherClientPorts struct {
	Ports []string
}

type clientShare struct {
  Share int
}

var clientinfo = 0
var clientPorts = []string{}
var numClients = 0
var countShares = 0

var mutex sync.Mutex


func main() {
  log.Printf("Listening on %s...", hos_port)

  // Does the Setup for starting a server
  hospital, err := hospitalSetup()
  if err != nil {
    panic("HospitalServer failed")
  }
  // This makes sure to keep listing for requests
  go listenAndServe(hospital)

  for {}
}





// ----------------------------------- Post Request handlers -----------------------------------

func sendPortsToAllClients() {
  // Pretty much just checks if this file exist
  cert, err := os.ReadFile("ca.crt")
  if err != nil {
    panic("No ca.crt")
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

  url := "https://localhost:"
  
  for index, elem := range clientPorts {
    ports := &OtherClientPorts{}

    
    innerIndex := index
    for i := 0; i < (maxClients - 1); i++{
      innerIndex = (innerIndex + 1) % maxClients
      ports.Ports = append(ports.Ports, clientPorts[innerIndex])
    }

    
    bodyBytes, err := json.Marshal(&ports)
    if err != nil {
      log.Printf("Error marshaling ports for client %s: %v", elem, err)
      continue // Skip to the next client
    }

    bodyReader := bytes.NewReader(bodyBytes)
    
    log.Printf("Posting to " + url + elem + "/GetClientsPorts")
    resp, err := client.Post((url + elem + "/GetClientsPorts"), "string", bodyReader)
    if err != nil || resp.StatusCode != 200{
      log.Printf("Failed to get response: %v", err)
      continue // Skip to the next client
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
      log.Printf("Unexpected status code from client %s: %d", elem, resp.StatusCode)
      continue // Skip to the next client
    }

    log.Printf("Successfully posted to client %s", elem)
  }
}







// ----------------------------------- Read Request handlers -----------------------------------

func readClientPort(body io.ReadCloser) (string) {
  // not sure why, but this has to &clientInfo else I non-pointer error in unmarshaling
  clientIn := &clientInfo{}

  bodyBytes, err := io.ReadAll(body)
  if err != nil {
    log.Printf("Failed to read response body: %v", err)
  }

  // Takes the content of the request and puts it in clientIn
  err = json.Unmarshal(bodyBytes, clientIn)
  if err != nil {
    log.Printf("Failed to Unmarshal: %v", err)
  }
  return clientIn.Port
}

func getSharesFromClients(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  body, err := io.ReadAll(r.Body)
		if err != nil {
			panic("failed to readAll")
		}

		share := &clientShare{}
		err = json.Unmarshal(body, share)
		if err != nil {
			panic("failed to Unmarshal")
		}

    mutex.Lock()
		clientinfo = clientinfo + share.Share
		countShares++
    mutex.Unlock()
    
		log.Println("got share : " + strconv.Itoa(share.Share))

		if countShares == maxClients {
			log.Println("Got all shares, Value is : " + strconv.Itoa(clientinfo))
		}
		
}








// ----------------------------------- Server Request handlers -----------------------------------

func connectionEstablished(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Thanks for joining the protocol, please send your port"))
}


// TODO: Take client's port and save it
func handleClientPortPost(w http.ResponseWriter, r *http.Request) {
  // gets the client port from the post request
  clientPort := readClientPort(r.Body)
  
  mutex.Lock()
  clientPorts = append(clientPorts, clientPort)
  mutex.Unlock()

  numClients++
  
  log.Printf("received port from client: %s", clientPort)
  log.Printf("Current known ports: %s", clientPorts)

  
  s1 := strconv.Itoa(maxClients)
  s2 := strconv.Itoa(numClients)
  
  log.Printf("max clients is set to " + s1 + ". Current number of clients is " + s2)
  if numClients >= maxClients {
    log.Printf("Reached max clients. Sending all ports to clients")
    sendPortsToAllClients()
  }
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Port: " + clientPort + " has been added"))
}







// ----------------------------------- Server Setup -----------------------------------

func hospitalSetup() (*http.Server, error) {
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
  router.HandleFunc("/ClientPortPost", handleClientPortPost)
  router.HandleFunc("/GetShares", getSharesFromClients)

  hospital := &http.Server{
    Addr:      ":" + hos_port,
    Handler:   router,
    TLSConfig: config,
  }

  return hospital, err
}

func listenAndServe(hospital *http.Server) {
  err := hospital.ListenAndServeTLS("", "")
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }
}

