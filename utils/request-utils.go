package utils

import (
	"errors"
	"net"
	"net/http"
	"path"
)

func GetIP(req *http.Request) (string, error) {
    ip, _, err := net.SplitHostPort(req.RemoteAddr)
    return ip, err
}

func GetPath(req *http.Request) string {
	return path.Dir(req.URL.Path)
}

func GetParameter(name string, req *http.Request) (string, error) {
    switch name {
    case "IP":
        return GetIP(req)
    case "PATH":
        return GetPath(req), nil
    default:
        return "", errors.New("invalid param")
    }
}