package cli

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/square/go-jose"
)

var (
	newReg            = flag.String("new-reg", "http://192.168.10.22:4000/acme/new-reg", "New Registration URL")
	accountKeyFile    = flag.String("certKey", "account.key", "Private key file. Created if it does not exist.")
	accountTmpCrtFile = flag.String("acc-crt-file", "account-tmp.pem", "Temporary self signed certificate for challenges.")

	email  = flag.String("email", "", "Email address used for certificate retrieval.")
	domain = flag.String("domain", "", "The domain to request a certificate for")

	ecdsaCurve = flag.String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521")
	bits       = flag.Int("bits", 2048, "The size of the RSA keys in bits. Default 4096")
)

// Logger is used to log errors; if nil, the default log.Logger is used.
var Logger *log.Logger

// logger is an helper function to retrieve the available logger
func logger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return Logger
}

type Registration struct {
	Contact []string `json:"contact"`
}

type Tos struct {
	Agreement string `json:"agreement"`
}

type GetChallenges struct {
	Identifier `json:"identifier"`
}

type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ChallengesResponse struct {
	Identifier   `json:"identifier"`
	Status       string      `json:"status"`
	Expires      time.Time   `json:"expires"`
	Challenges   []Challenge `json:"challenges"`
	Combinations [][]int     `json:"combinations"`
}

type Challenge struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	URI    string `json:"uri"`
	Token  string `json:"token"`
}

type SimpleHttpsMessage struct {
	Path string `json:"path"`
}

type CsrMessage struct {
	Csr            string   `json:"csr"`
	Authorizations []string `json:"authorizations"`
}

func execute() {
	flag.Parse()
	accountKey := generateKeyPair(*accountKeyFile).(*rsa.PrivateKey)
	jsonBytes, _ := json.Marshal(Registration{Contact: []string{"mailto:" + *email}})
	logger().Printf("Posting registration to %s", *newReg)

	resp, _ := jwsPost(*newReg, jsonBytes)

	links := parseLinks(resp.Header["Link"])
	if links["next"] == "" {
		logger().Fatalln("The server did not provide enough information to proceed.")
	}

	logger().Printf("Got agreement URL: %s", links["terms-of-service"])
	logger().Printf("Got new auth URL: %s", links["next"])

	jsonBytes, _ = json.Marshal(Tos{Agreement: links["terms-of-service"]})
	logger().Printf("Posting agreement to %s", resp.Header.Get("Location"))

	resp, _ = jwsPost(resp.Header.Get("Location"), jsonBytes)
	logResponse(resp)

	jsonBytes, _ = json.Marshal(GetChallenges{Identifier{Type: "dns", Value: *domain}})
	logger().Printf("Getting challenges for type %s and domain %s", "dns", *domain)

	resp, _ = jwsPost(links["next"], jsonBytes)
	logResponse(resp)

	links = parseLinks(resp.Header["Link"])
	if links["next"] == "" {
		logger().Fatalln("The server did not provide enough information to proceed.")
	}

	logger().Printf("Got new cert URL: %s", links["next"])
	logger().Printf("Got new authorization URL: %s", resp.Header.Get("Location"))
	newCertUrl := links["next"]
	authUrl := resp.Header.Get("Location")
	body, _ := ioutil.ReadAll(resp.Body)

	var challenges ChallengesResponse
	_ = json.Unmarshal(body, &challenges)

	for _, challenge := range challenges.Challenges {
		logger().Printf("Got challenge %s", challenge.Type)
	}
	logger().Printf("Challenge combinations are: %v", challenges.Combinations)
	logger().Printf("Choosing first challenge combination and starting with %s", challenges.Challenges[challenges.Combinations[0][0]].Type)

	firstChallenge := challenges.Challenges[challenges.Combinations[0][0]]
	if firstChallenge.Type == "simpleHttps" {
		generateSelfSignedCert(accountKey)

		path := getRandomString(8) + ".txt"
		challengePath := "/.well-known/acme-challenge/" + path
		startChallengeTlsServer(challengePath, firstChallenge.Token)

		logger().Print("Waiting for domain validation...")

		jsonBytes, _ = json.Marshal(SimpleHttpsMessage{Path: path})
		logger().Printf("Sending challenge response for path %s", path)

		resp, _ = jwsPost(firstChallenge.URI, jsonBytes)
		logResponse(resp)

		// Loop until status is verified or error.
		var challengeResponse Challenge
	loop:
		for {
			decoder := json.NewDecoder(resp.Body)
			decoder.Decode(&challengeResponse)

			switch challengeResponse.Status {
			case "valid":
				logger().Print("The CA validated our credentials. Continue...")
				break loop
			case "pending":
				logger().Print("The data is still being validated. Please stand by...")
			case "invalid":
				logger().Fatalf("The CA could not validate the provided file. - %v", challengeResponse)
			default:
				logger().Fatalf("The CA returned an unexpected state. - %v", challengeResponse)
			}

			time.Sleep(1000 * time.Millisecond)
			resp, _ = http.Get(authUrl)
		}
	}

	logger().Print("Getting certificate...")
	privateSslKey := generateKeyPair("ssl-priv.key")
	csr := generateCsr(privateSslKey)
	csrString := base64.URLEncoding.EncodeToString(csr)
	jsonBytes, _ = json.Marshal(CsrMessage{Csr: csrString, Authorizations: []string{authUrl}})
	resp, _ = jwsPost(newCertUrl, jsonBytes)
	logResponse(resp)

	body, _ = ioutil.ReadAll(resp.Body)
	ioutil.WriteFile("ssl-crt.crt", body, 0644)
}

func startChallengeTlsServer(path string, token string) {

	cert, err := tls.LoadX509KeyPair(*accountTmpCrtFile, *accountKeyFile)
	tlsConf := new(tls.Config)
	tlsConf.Certificates = []tls.Certificate{cert}

	tlsListener, err := tls.Listen("tcp", ":443", tlsConf)
	if err != nil {
		logger().Fatalf("Could not start TLS listener on %s for challenge handling! - %v", ":443", err)
	}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Host == *domain && r.Method == "GET" {
			w.Write([]byte(token))
			tlsListener.Close()
		}
	})

	srv := http.Server{Addr: ":443", Handler: nil}
	go func() {
		srv.Serve(tlsListener)
		logger().Print("TLS Server exited.")
	}()
}

func getRandomString(length int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func logResponse(resp *http.Response) {
	logger().Println(resp.Status)
	for k, v := range resp.Header {
		logger().Printf("-- %s: %s", k, v)
	}
}

func jwsPost(url string, content []byte) (*http.Response, error) {
	url = strings.Replace(url, "localhost", "192.168.10.22", -1)

	privKeyBytes, err := ioutil.ReadFile(*accountKeyFile)
	key, err := jose.LoadPrivateKey(privKeyBytes)
	if err != nil {
		panic(err)
	}

	signer, err := jose.NewSigner(jose.RS256, key)
	if err != nil {
		panic(err)
	}
	signed, err := signer.Sign(content)
	if err != nil {
		panic(err)
	}
	signedContent := signed.FullSerialize()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(signedContent)))
	if err != nil {
		logger().Fatalf("Error posting content: %s", err)
	}

	return resp, err
}

func generateCsr(privateKey interface{}) []byte {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: *domain,
		},
		EmailAddresses: []string{*email},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)

	csrOut, err := os.Create("csr.pem")
	if err != nil {
		logger().Fatalf("Could not create certificate request file: %s", err)
	}

	pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return csrBytes
}

func generateKeyPair(fileName string) interface{} {
	logger().Println("Generating key pair ...")

	var privateKey interface{}
	var err error
	switch *ecdsaCurve {
	case "":
		privateKey, err = rsa.GenerateKey(rand.Reader, *bits)
	case "P224":
		privateKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		privateKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	}

	if err != nil {
		logger().Fatalf("Failed to generate private key: %s", err)
	}

	var pemKey pem.Block
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		pemKey = pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	case *ecdsa.PrivateKey:
		privateBytes, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			logger().Fatalf("Could not marshal ECDSA private key: %v", err)
		}

		pemKey = pem.Block{Type: "EC PRIVATE KEY", Bytes: privateBytes}
	}

	certOut, err := os.Create(fileName)
	if err != nil {
		logger().Fatalf("Could not create private key file: %s", err)
	}

	pem.Encode(certOut, &pemKey)
	certOut.Close()

	return privateKey
}

func generateSelfSignedCert(privKey *rsa.PrivateKey) {

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   *email,
			Organization: []string{*domain},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{*domain},
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create(*accountTmpCrtFile)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
}
