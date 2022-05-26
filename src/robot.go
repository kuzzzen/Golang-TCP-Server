package main

type Coordinates [2]int

type Robot struct {
	name              string
	coordinates       *Coordinates
	direction         Direction
	coordsInitialised bool
}

func makeRobot() Robot {
	robot := Robot{}
	robot.coordinates = new(Coordinates)
	robot.direction = UNDEFINED
	return robot
}
