package main

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

func abStrip(s string) string {
	return strings.Replace(s, "\a\b", "", 1)
}

func abSplit(s string) []string {
	return strings.SplitN(s, "\a\b", 2)
}

func getHash(name string) int {
	sum := 0
	for _, ch := range name {
		sum += int(ch)
	}
	return sum * 1000 % 65536
}

func findKey(id string) (sKey, cKey int, err error) {
	intId, err := strconv.Atoi(id)
	outOfRange := !(0 <= intId && intId <= 4)
	if err != nil || outOfRange {
		var errRet error
		if outOfRange {
			errRet = errors.New(ServerKeyOutOfRangeError)
		} else {
			errRet = errors.New(ServerSyntaxError)
		}
		return math.MaxInt, math.MaxInt, errRet
	}
	sKey = keyPairs[intId]['s']
	cKey = keyPairs[intId]['c']
	return sKey, cKey, nil
}

func handle(con net.Conn) {
	defer func(con net.Conn) {
		err := con.Close()
		if err != nil {
			fmt.Println("Could not close connection properly")
		}
	}(con)

	fmt.Printf("Connected: [%s] \n", con.RemoteAddr().String())
	controller := initialiseController(con)

	authM, err := controller.getCommand(MaxUsernameLen)
	if err != nil {
		fmt.Println("Something went wrong while trying to get command")
		_, err := controller.connection.Write([]byte(err.Error()))
		if err != nil {
			fmt.Printf("Could not send error through connection (%s)\n", []byte(err.Error()))
		}
		return
	}

	err = controller.auth(authM)
	if err != nil {
		fmt.Println("Something went wrong during authentication")
		_, err := controller.connection.Write([]byte(err.Error()))
		if err != nil {
			fmt.Printf("Could not send error through connection (%s)\n", []byte(err.Error()))
		}
		return
	}

	err = controller.initRobotCoords()
	if err != nil {
		fmt.Println("Something went wrong when trying to initialise robot coords")
		_, err := controller.connection.Write([]byte(err.Error()))
		if err != nil {
			fmt.Printf("Could not send error through connection (%s)\n", []byte(err.Error()))
		}
		return
	}

	err = controller.findMessage()
	if err != nil {
		fmt.Println("Something went wrong when searching for hidden message")
		_, err := controller.connection.Write([]byte(err.Error()))
		if err != nil {
			fmt.Printf("Could not send error through connection (%s)\n", []byte(err.Error()))
		}
		return
	}

	message, err := controller.getMessage()
	if err != nil {
		fmt.Println("Something went wrong when getting the hidden message")
		_, err := controller.connection.Write([]byte(err.Error()))
		if err != nil {
			fmt.Printf("Could not send error through connection (%s)\n", []byte(err.Error()))
		}
		return
	}

	fmt.Printf("Success! Got the message: %s\n. Isn't it great?", message)
	_, err = controller.connection.Write([]byte(ServerLogout))
	if err != nil {
		fmt.Printf("Could not send error through connection (%s,: %s)\n", ServerLogout, err.Error())
	}
}

func main() {
	dstream, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(dstream net.Listener) {
		err := dstream.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(dstream)

	for {
		con, err := dstream.Accept()
		if err != nil {
			fmt.Println(err)
			err := con.Close()
			if err != nil {
				fmt.Println("Could not close connection")
			}
			continue
		}
		go handle(con)
	}
}
