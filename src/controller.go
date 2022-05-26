package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// RobotConnection (controller type that pairs Robot and net.Conn)
type RobotConnection struct {
	robot         Robot
	connection    net.Conn
	commandBuffer string
	hash          int
	clientHash    int
	serverHash    int
}

func initialiseController(con net.Conn) RobotConnection {
	controller := RobotConnection{}
	controller.connection = con
	controller.robot = makeRobot()
	return controller
}

func (controller *RobotConnection) setHashes(sKey, cKey int) {
	controller.hash = getHash(controller.robot.name)
	controller.serverHash = (controller.hash + sKey) % 65536
	controller.clientHash = (controller.hash + cKey) % 65536
}

// Reads whatever is being written
func (controller *RobotConnection) readConnectionBuffer(tl time.Duration) (err error) {
	err = controller.connection.SetReadDeadline(time.Now().Add(tl))
	if err != nil {
		return err
	}
	buffer := make([]byte, 1024)
	length, err := controller.connection.Read(buffer)
	if err != nil || length == 0 {
		fmt.Println(err)
		return err
	}
	controller.commandBuffer += string(buffer[:length])
	return nil
}

// Gets actual commands from messages sent through connection
func (controller *RobotConnection) getCommand(lenLimit int) (command string, err error) {
	for {
		split := abSplit(controller.commandBuffer)
		if len(split) == 2 {
			command = split[0]
			controller.commandBuffer = split[1]
			if command == abStrip(ClientRecharging) {
				err = controller.recharge()
				if err != nil {
					fmt.Println("Could not recharge")
					return
				}
				continue
			}
			return command, nil
		} else if len(controller.commandBuffer) >= lenLimit {
			return "", errors.New(ServerSyntaxError)
		}
		err := controller.readConnectionBuffer(1 * time.Second)
		if err != nil {
			fmt.Println("Error reading from buffer")
			return "", err
		}
	}
}

func (controller *RobotConnection) initRobotCoords() (err error) {
	for movesMade := 0; movesMade < 2; movesMade++ {
		_, err = controller.connection.Write([]byte(ServerMove))
		if err != nil {
			return err
		}

		command, err := controller.getCommand(MaxOkLen)
		if err != nil {
			return err
		}

		_, err = controller.setRobotCoordsAndDirection(command)
		if err != nil {
			return err
		}
	}
	return nil
}

func (controller *RobotConnection) setRobotCoordsAndDirection(command string) (coordsOld *Coordinates, err error) {
	split := strings.Split(command, " ")
	if split[0] != "OK" || len(split) != 3 {
		return nil, errors.New(ServerSyntaxError)
	}
	split = split[1:]
	var coords Coordinates
	coords[0], err = strconv.Atoi(split[0])
	if err != nil {
		return nil, errors.New(ServerSyntaxError)
	}
	coords[1], err = strconv.Atoi(split[1])
	if err != nil {
		return nil, errors.New(ServerSyntaxError)
	}

	coordsBeforeCalled := *controller.robot.coordinates
	controller.robot.coordinates[0] = coords[0]
	controller.robot.coordinates[1] = coords[1]
	controller.setRobotDirection(&coordsBeforeCalled)

	return &coordsBeforeCalled, nil
}

func (controller *RobotConnection) setRobotDirection(coordsBeforeCalled *Coordinates) {
	if coordsBeforeCalled != nil {
		if controller.robot.coordinates[0] > coordsBeforeCalled[0] {
			controller.robot.direction = RIGHT
		} else if controller.robot.coordinates[0] < coordsBeforeCalled[0] {
			controller.robot.direction = LEFT
		} else if controller.robot.coordinates[1] > coordsBeforeCalled[1] {
			controller.robot.direction = UP
		} else if controller.robot.coordinates[1] < coordsBeforeCalled[1] {
			controller.robot.direction = DOWN
		}
	}
}

func (controller *RobotConnection) recharge() (err error) {
	for {
		err = controller.readConnectionBuffer(5 * time.Second)
		if err != nil {
			return err
		}

		split := abSplit(controller.commandBuffer)
		if len(split) == 2 {
			info := split[0]
			controller.commandBuffer = split[1]
			if info != abStrip(ClientFullPower) {
				return errors.New(ServerLogicError)
			}
			return nil
		}
	}
}

func (controller *RobotConnection) getResponse(command string, lengthLimit int) (response string, err error) {
	_, err = controller.connection.Write([]byte(command))
	if err != nil {
		return "", err
	}
	return controller.getCommand(lengthLimit)
}

func (controller *RobotConnection) moveRobotUp() (err error) {
	dir := controller.robot.direction
	if dir == LEFT {
		err = controller.rotateOrMoveRobot(ServerTurnRight)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == DOWN {
		for i := 0; i < 2; i++ {
			err = controller.rotateOrMoveRobot(ServerTurnLeft)
			if err != nil {
				return err
			}
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == RIGHT {
		err = controller.rotateOrMoveRobot(ServerTurnLeft)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else {
		err = controller.rotateOrMoveRobot(ServerMove)
	}
	return err
}

func (controller *RobotConnection) moveRobotDown() (err error) {
	dir := controller.robot.direction
	if dir == LEFT {
		err = controller.rotateOrMoveRobot(ServerTurnLeft)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == UP {
		for i := 0; i < 2; i++ {
			err = controller.rotateOrMoveRobot(ServerTurnLeft)
			if err != nil {
				return err
			}
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == RIGHT {
		err = controller.rotateOrMoveRobot(ServerTurnRight)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else {
		err = controller.rotateOrMoveRobot(ServerMove)
	}
	return err
}

func (controller *RobotConnection) moveRobotLeft() (err error) {
	dir := controller.robot.direction
	if dir == UP {
		err = controller.rotateOrMoveRobot(ServerTurnLeft)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == RIGHT {
		for i := 0; i < 2; i++ {
			err = controller.rotateOrMoveRobot(ServerTurnLeft)
			if err != nil {
				return err
			}
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == DOWN {
		err = controller.rotateOrMoveRobot(ServerTurnRight)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else {
		err = controller.rotateOrMoveRobot(ServerMove)
	}
	return err
}

func (controller *RobotConnection) moveRobotRight() (err error) {
	dir := controller.robot.direction
	if dir == DOWN {
		err = controller.rotateOrMoveRobot(ServerTurnLeft)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == LEFT {
		for i := 0; i < 2; i++ {
			err = controller.rotateOrMoveRobot(ServerTurnLeft)
			if err != nil {
				return err
			}
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else if dir == UP {
		err = controller.rotateOrMoveRobot(ServerTurnRight)
		if err != nil {
			return err
		}
		err = controller.rotateOrMoveRobot(ServerMove)
	} else {
		err = controller.rotateOrMoveRobot(ServerMove)
	}
	return err
}

func (controller *RobotConnection) findMessage() (err error) {
	fmt.Println(controller.robot.coordinates)
	robot := controller.robot
	for robot.coordinates[0] != 0 {
		fmt.Printf("Robot named %s is moving to [%v, %v]\n", robot.name, robot.coordinates[0], robot.coordinates[1])
		c := robot.coordinates[0]
		fmt.Println(c)
		dir := robot.direction
		if c > 0 {
			err = controller.moveRobotLeft()
		} else {
			err = controller.moveRobotRight()
		}
		if c == robot.coordinates[0] {
			robot.direction = dir
			fmt.Printf("Robot named %s is bumped into an obstacle\n", robot.name)
			err := controller.handleObstacle()
			if err != nil {
				fmt.Println("Could not handle obstacle.")
				return err
			}
			return controller.findMessage()
		}
	}

	for robot.coordinates[1] != 0 {
		fmt.Printf("Robot named %s is moving to [%v, %v]\n", robot.name, robot.coordinates[0], robot.coordinates[1])
		c := robot.coordinates[1]
		if c < 0 {
			err = controller.moveRobotUp()
		} else {
			err = controller.moveRobotDown()
		}
		if c == robot.coordinates[1] {
			fmt.Printf("Robot named %s is bumped into an obstacle\n", robot.name)
			err = controller.handleObstacle()
			if err != nil {
				fmt.Println("Could not handle obstacle.")
				return err
			}
			return controller.findMessage()
		}
	}
	fmt.Printf("RESULTING COORDS: [%v, %v]\n", controller.robot.coordinates[0], controller.robot.coordinates[1])
	return nil
}

// send "SERVER_MOVE" as instruction to move the controlled robot
func (controller *RobotConnection) rotateOrMoveRobot(instruction string) (err error) {
	response, err := controller.getResponse(instruction, MaxOkLen)
	if err != nil {
		return err
	}
	_, err = controller.setRobotCoordsAndDirection(response)
	return err
}

func (controller *RobotConnection) getMessage() (result string, err error) {
	result, err = controller.getResponse(ServerPickUp, MaxMessageLen)
	return result, err
}

func (controller *RobotConnection) handleObstacle() error {
	err := controller.rotateOrMoveRobot(ServerTurnLeft)
	fmt.Printf("Robot name %s handling obstacle: moving to [%v, %v]\n", controller.robot.name, controller.robot.coordinates[0], controller.robot.coordinates[1])
	if err != nil {
		return err
	}
	err = controller.rotateOrMoveRobot(ServerMove)
	fmt.Printf("Robot name %s handling obstacle: moving to [%v, %v]\n", controller.robot.name, controller.robot.coordinates[0], controller.robot.coordinates[1])
	if err != nil {
		return err
	}

	err = controller.rotateOrMoveRobot(ServerTurnRight)
	fmt.Printf("Robot name %s handling obstacle: moving to [%v, %v]\n", controller.robot.name, controller.robot.coordinates[0], controller.robot.coordinates[1])
	if err != nil {
		return err
	}
	err = controller.rotateOrMoveRobot(ServerMove)
	fmt.Printf("Robot name %s handling obstacle: moving to [%v, %v]\n", controller.robot.name, controller.robot.coordinates[0], controller.robot.coordinates[1])
	return err
}

func (controller *RobotConnection) auth(authMessage string) (err error) {
	if len(authMessage) >= MaxUsernameLen-1 {
		return errors.New(ServerSyntaxError)
	}

	controller.robot.name = authMessage
	_, err = controller.connection.Write([]byte(ServerKeyRequest))
	if err != nil {
		return err
	}

	keyID, err := controller.getCommand(MaxKeyIdLen)
	if err != nil {
		return err
	}

	sKey, cKey, err := findKey(keyID)
	if err != nil {
		return err
	}
	controller.setHashes(sKey, cKey)
	_, err = controller.connection.Write([]byte(strconv.FormatInt(int64(controller.serverHash), 10) + "\a\b"))
	cHash, err := controller.getCommand(MaxConfirmationLen)
	if err != nil {
		return err
	}
	if len(cHash) > 5 {
		return errors.New(ServerSyntaxError)
	}
	cHashInt, err := strconv.Atoi(cHash)
	if err != nil {
		return errors.New(ServerSyntaxError)
	}
	if cHashInt != controller.clientHash {
		return errors.New(ServerLoginFailed)
	} else {
		_, err = controller.connection.Write([]byte(ServerOk))
		if err != nil {
			return err
		}
	}
	return nil
}
