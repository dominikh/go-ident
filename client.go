// Package ident implements an RFC 1413 client
package ident

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Response struct {
	OS         string
	Charset    string
	Identifier string
}

type ResponseError struct {
	Type string
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Ident error: %s", e.Type)
}

type ProtocolError struct {
	Line string
}

func (e ProtocolError) Error() string {
	return fmt.Sprintf("Unexpected response from server: %s", e.Line)
}

func Query(ip string, portOnServer, portOnClient int) (Response, error) {
	var (
		conn   net.Conn
		err    error
		fields []string
		r      *bufio.Reader
		resp   string
	)

	conn, err = net.Dial("tcp", ip+":113")
	if err != nil {
		goto Error
	}

	_, err = conn.Write([]byte(fmt.Sprintf("%d, %d", portOnServer, portOnClient)))
	if err != nil {
		goto Error
	}

	r = bufio.NewReader(conn)
	resp, err = r.ReadString('\n')
	if err != nil {
		goto Error
	}

	fields = strings.SplitN(strings.TrimSpace(resp), " : ", 4)
	if len(fields) < 3 {
		goto ProtocolError
	}

	switch fields[1] {
	case "USERID":
		if len(fields) != 4 {
			goto ProtocolError
		}

		var os, charset string
		osAndCharset := strings.SplitN(fields[2], ",", 2)
		if len(osAndCharset) == 2 {
			os = osAndCharset[0]
			charset = osAndCharset[1]
		} else {
			os = osAndCharset[0]
			charset = "US-ASCII"
		}

		return Response{
			OS:         os,
			Charset:    charset,
			Identifier: fields[3],
		}, nil
	case "ERROR":
		if len(fields) != 3 {
			goto ProtocolError
		}

		return Response{}, ResponseError{fields[3]}
	}
ProtocolError:
	return Response{}, ProtocolError{resp}
Error:
	return Response{}, err
}
