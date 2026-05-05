package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"strings"
)

func SendLoopbackCommand(statePath, command string) error {
	_, err := sendLoopbackRequest(statePath, Request{Command: command}, "control command failed")
	return err
}

func SendLoopbackPayload(statePath string, payload json.RawMessage) (json.RawMessage, error) {
	resp, err := sendLoopbackRequest(statePath, Request{Payload: payload}, "payload request failed")
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func sendLoopbackRequest(statePath string, req Request, fallbackError string) (Response, error) {
	endpoint, err := ReadEndpoint(statePath)
	if err != nil {
		return Response{}, err
	}

	conn, err := net.Dial("tcp", endpoint.Address)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()

	req.Token = endpoint.Token
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return Response{}, err
	}

	var resp Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return Response{}, err
	}
	if !resp.OK {
		if strings.TrimSpace(resp.Error) == "" {
			return Response{}, errors.New(fallbackError)
		}
		return Response{}, errors.New(resp.Error)
	}
	return resp, nil
}
