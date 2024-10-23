package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type OtherClientPorts struct {
	Ports []string
}

type clientInfo struct {
	Port string
}

type clientShare struct {
	Share int
}

var mutex sync.Mutex

var client_port = ""

var maxClients = 3

var sharesFromClients []int

var sec = 500
var capForCom int
var N int
var ownMadeShares []int

var data int
var dataMax int

var client *http.Client

const (
	// by changing the the ending, can you make it possible to reach different request handlers
	// It is just a  different endpoint.
	urlHos = "https://localhost:8443"
)

func main() {
	port := os.Args[1:]

	if cap(port) != 1 {
		client_port = "8445"
	} else {
		client_port = port[0]
	}
	// println("port: " + port[0])

	// capForCom = sec / 3
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// N = r.Intn(capForCom)

	dataMax = 500 / 3

	rand.Seed(time.Now().UnixNano())
	data = rand.Intn(dataMax)

	// Does the Setup for starting a server
	clientServer, err := clientSetup()
	if err != nil {
		panic("ClientServer failed")
	}
	log.Printf("ClientServer: " + client_port + " is up and running")

	// This makes sure to keep listing for requests
	go listenAndServe(clientServer)

	// Starts a connection to the hospital to see if can response
	clientCon := connection()

	//println(status)

	// with the hospital alive and running send client's port to it
	postPortToHospital(clientCon)

	// Keep the main function alive
	for {
	}
}

// func makeShares (secret, N, fieldSize int) ([]int) {
//   var shares []int
// 	var totalShares int

// 	for i := 0; i < fieldSize-1; i++ {
// 		share := rand.Intn(secret-1) + 1
// 		shares = append(shares, share)
// 		totalShares += share
// 	}

// 	shares = append(shares, N-totalShares)
//   log.Println("make share check")
// 	return shares
// }

func makeShares(p int, data int, amount int) []int {
	log.Println("makeShares check")
	var shares []int
	var totalShares int

	for i := 0; i < amount-1; i++ {
		share := rand.Intn(p-1) + 1
		log.Println("share nr " + strconv.Itoa(i) + ": " + strconv.Itoa(share))
		shares = append(shares, share)
		totalShares += share
	}

	shares = append(shares, data-totalShares)

	return shares
}




// ----------------------------------- Post Request handlers -----------------------------------

// As the name says, gives the client's port to hospital
func postPortToHospital(client *http.Client) {
	clientIn := clientInfo{
		Port: client_port,
	}
	//log.Printf("port: %v", clientIn)

	// Sadly, this is needed, you cant post without a body...
	bodyBytes, err := json.Marshal(&clientIn)
	if err != nil {
		log.Fatal(err)
	}
	bodyReader := bytes.NewReader(bodyBytes)

	// This post request gives the current client's port to the hospital
	resp, err := client.Post((urlHos + "/ClientPortPost"), "string", bodyReader)
	responseHandler(resp)
}

func postAggSharesToHos() {
	log.Println("postAggSharesToHos check")
	log.Println("OwmMAdeShares len : " + strconv.Itoa(len(ownMadeShares)))
	log.Println("OwmMAdeShares : " + strconv.Itoa(ownMadeShares[len(ownMadeShares)-1]))

	sharesFromClients = append(sharesFromClients, ownMadeShares[len(ownMadeShares)-1])

	var aggregateShare int

	for _, share := range sharesFromClients {
		aggregateShare = aggregateShare + share
	}

	log.Println("aggregate share is " + strconv.Itoa(aggregateShare))

	aggregate := clientShare{
		Share: aggregateShare,
	}

	b, err := json.Marshal(aggregate)
	if err != nil {
		panic("marshal wrong")
	}

	// url := fmt.Sprintf("https://localhost:%d/shares", hospitalPort)
	resp, err := client.Post(urlHos+"/GetShares", "string", bytes.NewReader(b))
	if err != nil {
		panic("response wrong")
	}
	log.Println("Posted Agg shares to hos, code : " + resp.Status)
}

func postShareToClient(ports []string) {
	// log.Println("post share check")
	// log.Println("ownShare len : " + strconv.Itoa(len(ownShare)) )
	// log.Println("ports len : " + strconv.Itoa(len(ports)))

	for index, share := range ownMadeShares {
		if index == maxClients-1 {
			break
		}
		ownShare := clientShare{
			Share: share,
		}

		b, err := json.Marshal(ownShare)
		if err != nil {
			panic("marshel wrong")
		}
		urlToClients := "https://localhost:"
		// log.Println("posting to " + urlToClients + ports[index] + "/SendShares check")
		// log.Println("ownShare share : " + strconv.Itoa(ownShare.share))
		// log.Println("index : " + strconv.Itoa(index))
		log.Println("share : " + strconv.Itoa(share))
		resp, err := client.Post(urlToClients+ports[index]+"/SendShares", "string", bytes.NewReader(b))
		if err != nil {
			panic("response To " + ports[index] + " went wrong")
		}
		log.Println("post share " + strconv.Itoa(ownShare.Share) + " to: " + ports[index] + " code: " + resp.Status)
	}
}






// ----------------------------------- Read Request handlers -----------------------------------

func readClientPorts(body io.ReadCloser) []string {
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

func readSharePost(body io.ReadCloser) int {
	clientShare := &clientShare{}
	// log.Println("read share check")

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
	}

	// Takes the content of the request and puts it in clientIn
	err = json.Unmarshal(bodyBytes, clientShare)
	if err != nil {
		log.Printf("Failed to Unmarshal: %v", err)
	}

	// just to make it easier, I add it to global var
	log.Println("share from client : " + strconv.Itoa(clientShare.Share))
	sharesFromClients = append(sharesFromClients, clientShare.Share)

	return clientShare.Share
}






// ----------------------------------- Server Request handlers -----------------------------------

func connectionEstablished(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hey from client with port:" + client_port))
}

func hospitalPostsAllPort(w http.ResponseWriter, r *http.Request) {
	ports := readClientPorts(r.Body)
	log.Printf("Got Ports: %v", ports)
	// time.Sleep(time.Second * 1)

	// mutex.Lock()
	ownMadeShares = makeShares(dataMax, data, maxClients)
	// mutex.Unlock()
	// ownMadeShares = makeShares(capForCom, N, maxClients)

	// time.Sleep(time.Second * 1)
	postShareToClient(ports)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("" + client_port + " : Got all the ports"))
}

func GetShareFromClients(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	share := readSharePost(r.Body)
	mutex.Unlock()

	log.Printf("Got share: %v", share)

	log.Println("sharesFromClients len : " + strconv.Itoa(len(sharesFromClients)))

	if len(ownMadeShares) != 0 && (maxClients-1) == len(sharesFromClients) {
		postAggSharesToHos()
	}

	w.WriteHeader(http.StatusOK)
}







// ----------------------------------- Server Setup -----------------------------------

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
  router.HandleFunc("/SendShares", GetShareFromClients)

  clientServer := &http.Server{
    Addr:      ":" + client_port,
    Handler:   router,
    TLSConfig: config,
  }

  return clientServer, err
}

func listenAndServe(clientServer *http.Server) {
  err := clientServer.ListenAndServeTLS("", "")
  if err != nil {
    log.Fatalf("Failed to start server: %v", err)
  }
}






// ----------------------------------- Client connection Setup -----------------------------------

func connection() (*http.Client) {
  // setup meaning, what the client is using in as configs in http request
  client = connectionSetup()

  // Simple http get request, just to make sure that the hospital is running
  resp, err := client.Get(urlHos)
  if err != nil {
    log.Printf("Failed to get response: %v", err)

  }

  responseHandler(resp)
  return client
}

func getRequestToHospital(client *http.Client, url string) (*http.Response, error) {
  var resp *http.Response
    var err error

    // Loop indefinitely until a successful response is received
    for {
        resp, err = client.Get(url)
        if err != nil {
            log.Println("Error making request:", err)
            time.Sleep(1 * time.Second)
            continue
        }

        if resp.StatusCode == http.StatusOK {
            return resp, nil
        } else {
            log.Printf("Received response: %d\n", resp.StatusCode)
            time.Sleep(1 * time.Second)
        }
    }
}


func responseHandler(resp *http.Response) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return
	}

	log.Printf("Hospital: %s\n", body)
}



func connectionSetup() *http.Client {
	// Pretty much just checks if this file exist
	cert, err := os.ReadFile("ca.crt")
	if err != nil {
		return nil
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

	return client
}





