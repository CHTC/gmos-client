package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Perform a mutual authentication handshake with the Glidein Manager.
// The authentication occurs across a set of 3 request/response pairs:
//
//  1. The client initiates a request to the Glidein Manager containing its hostname.
//     The Glidein Manager checks that the client's name is in its allow-list, then responds
//     with a challenge value.
//
//  2. The Glidein Manager submits a request to the the client's hostname and listener port
//     containing an access token. The client responds with the challenge value given by the
//     Glidein Manager in the first request. If the returned challenge value is incorrect, the
//     Glidein Manager discards the access token.
//
//  3. The client submits the access token back to the Glidein Manager. The Glidein Manager then
//     activates the access token, which can be used in subsequent authenticated requests.
//
// To support handling the callback from the Glidein Manager in step (2), the client must start
// a temporary webserver that listens on the given exposed listenerPort.
func (gm *GlideinManagerClient) DoHandshake(listenerPort int) error {

	hs := HandshakeCallbackHandler{}
	// 0. Start a temporary server to listen for a callback from the Glidein Manager
	hs.startCallbackListener(fmt.Sprintf("0.0.0.0:%v", listenerPort))
	defer hs.stopCallbackListener(context.TODO())

	// 1. Submit a request to initiate the handshake
	challenge, err := gm.initiateHandshake(listenerPort)
	if err != nil {
		return err
	}

	// 2. Supply the challenge value from the initial response to the callback handler,
	//    Retrieve the capability given in the callback
	capability, err := hs.exchangeChallenge(challenge, 5*time.Second)
	if err != nil {
		return err
	}

	// 3. Validate the capability against the Glidein Manager to ensure the exchange
	//    was successful
	if err := gm.validateCapability(capability); err != nil {
		return err
	}

	// 4. Store the credentials for subsequent authenticated requests
	gm.Credentials = GlideinManagerCredentials{Capability: capability}
	return nil
}

// Initiate the handshake via a web request as described in step (1).
func (gm *GlideinManagerClient) initiateHandshake(listenerPort int) (ChallengeInitiateResponse, error) {
	initiateReq := ChallengeInitiateRequest{
		ClientName:      gm.HostName,
		CallbackAddress: fmt.Sprintf("http://%v:%v/challenge/response", gm.HostName, listenerPort),
	}
	initiateResp := ChallengeInitiateResponse{}

	body, err := json.Marshal(initiateReq)
	if err != nil {
		return initiateResp, errors.Wrap(err, "failed to construct challenge/initiate request body")
	}
	respBody, err := http.Post(gm.RouteFor("/api/public/challenge/initiate"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return initiateResp, errors.Wrap(err, "failed to submit challenge/initiate request")
	}

	return initiateResp, UnmarshalBody(respBody.Body, &initiateResp)
}

// Submit a basic-auth authenticated request using the given capability as described in (3)
func (gm *GlideinManagerClient) validateCapability(capability string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", gm.RouteFor("/api/private/verify-auth"), nil)
	if err != nil {
		return errors.Wrap(err, "failed to construct verify-auth request")
	}

	req.SetBasicAuth(gm.HostName, capability)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to submit verify-auth request")
	}

	if resp.StatusCode != 200 {
		return errors.New("verify-auth credential rejected")
	}

	return nil
}

type CapabilityResult struct {
	capability string
	err        error
}

// Helper struct that creates a temporary server that listens for the Glidein Manager's callback
// in step (2)
type HandshakeCallbackHandler struct {
	challengeChannel  chan ChallengeInitiateResponse
	capabilityChannel chan CapabilityResult
	server            *http.Server
}

// helper function to write an error message to both the http response writer and
// capability output channel
func (hs *HandshakeCallbackHandler) callbackError(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
	hs.capabilityChannel <- CapabilityResult{capability: "", err: err}
}

// Web handler for the temporary server. Receives a POST request from the Glidein Manager,
// then waits on a channel. Reads the challenge received in step (1) from an input channel, then
// writes the access token supplied in the callback to an output channel
func (hs *HandshakeCallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	var req ChallengeCompleteRequest
	if err := UnmarshalBody(r.Body, &req); err != nil {
		hs.callbackError(w, errors.Wrap(err, "failed to unmarshal callback request body"), http.StatusBadRequest)
		return
	}

	challenge := ChallengeInitiateResponse{}
	select {
	case challenge = <-hs.challengeChannel:
		log.Printf("Got id token %v; challenge secret %v", challenge.IdSecret, challenge.ChallengeSecret)
	case <-time.After(5 * time.Second): // require a challenge to be supplied within a timeout
		err := errors.New("didn't receive a challenge secret within expected duration")
		hs.callbackError(w, err, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ChallengeCompleteResponse{ChallengeSecret: challenge.ChallengeSecret}); err != nil {
		hs.callbackError(w, errors.Wrap(err, "failed to construct callback response body"), http.StatusInternalServerError)
		return
	}
	// TODO this is hacky, need to ensure response processing completes on Glidein Manager side
	// before informing caller that exchange was successful
	go func() {
		time.Sleep(100 * time.Millisecond)
		hs.capabilityChannel <- CapabilityResult{capability: req.Capability, err: nil}
	}()
}

// Write a value to the callback handler's input channel, then read a value from its output channel
func (hs *HandshakeCallbackHandler) exchangeChallenge(challenge ChallengeInitiateResponse, timeout time.Duration) (string, error) {
	select {
	case hs.challengeChannel <- challenge:
		// no-op
	case <-time.After(timeout):
		return "", errors.New("timeout writing to response handler goroutine")
	}
	select {
	case cap := <-hs.capabilityChannel:
		return cap.capability, cap.err
	case <-time.After(timeout):
		return "", errors.New("timeout waiting for callback from server")
	}
}

// Start a temporary webserver server in a background goroutine to listen to the callback from the
// Glidein Manager
func (hs *HandshakeCallbackHandler) startCallbackListener(addr string) {
	hs.challengeChannel = make(chan ChallengeInitiateResponse)
	hs.capabilityChannel = make(chan CapabilityResult)
	hs.server = &http.Server{Addr: addr}

	// Setup a handler for the response sent by the gm-object-server
	http.HandleFunc("/challenge/response", hs.handleCallback)

	// Start a server in a separate goroutine, defer its shutdown to the end of this call
	go func() {
		if err := hs.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Unexpected Server Error: %v\n", err)
		}
		log.Println("Server shut down as expected")
	}()
}

// Stop the temporary webserver
func (hs *HandshakeCallbackHandler) stopCallbackListener(ctx context.Context) {
	hs.server.Shutdown(ctx)
}
