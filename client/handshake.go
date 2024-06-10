package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type handshakeCallbackHandler struct {
	challengeChannel  chan ChallengeInitiateResponse
	capabilityChannel chan string
}

func (hs *handshakeCallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	var req ChallengeCompleteRequest
	if err1 := UnmarshalBody(r.Body, &req); err1 != nil {
		http.Error(w, err1.Error(), http.StatusBadRequest)
		return
	}

	challenge := <-hs.challengeChannel
	log.Printf("Got id token %v; challenge secret %v", challenge.IdSecret, challenge.ChallengeSecret)

	hs.capabilityChannel <- req.Capability

	if err := json.NewEncoder(w).Encode(ChallengeCompleteResponse{ChallengeSecret: challenge.ChallengeSecret}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (hs *handshakeCallbackHandler) exchangeChallenge(challenge ChallengeInitiateResponse) string {
	hs.challengeChannel <- challenge
	return <-hs.capabilityChannel
}

func (gm *GlideinManagerClient) DoHandshake() (string, error) {
	handshakeHandler := handshakeCallbackHandler{
		challengeChannel:  make(chan ChallengeInitiateResponse),
		capabilityChannel: make(chan string),
	}
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%v", gm.Port)}

	// Setup a handler for the response sent by the gm-object-server
	http.HandleFunc("/challenge/response", handshakeHandler.handleCallback)

	// Start a server in a separate goroutine, defer its shutdown to the end of this call
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Unexpected Server Error: %v\n", err)
		}
		log.Println("Server shut down as expected")
	}()
	defer server.Shutdown(context.TODO())

	challenge, err := gm.initiateHandshake()
	if err != nil {
		return "", err
	}
	capability := handshakeHandler.exchangeChallenge(challenge)

	return capability, nil
}

func (gm *GlideinManagerClient) initiateHandshake() (ChallengeInitiateResponse, error) {
	initiateReq := ChallengeInitiateRequest{
		ClientName:      gm.HostName,
		CallbackAddress: fmt.Sprintf("%v:%v/challenge/response", gm.HostName, gm.Port),
	}
	initiateResp := ChallengeInitiateResponse{}

	body, err1 := json.Marshal(initiateReq)
	respBody, err2 := http.Post(fmt.Sprintf("%v/api/public/challenge/initiate", gm.ManagerUrl), "application/json", bytes.NewBuffer(body))
	if err := errors.Join(err1, err2); err != nil {
		return initiateResp, err
	}

	return initiateResp, UnmarshalBody(respBody.Body, &initiateResp)
}
