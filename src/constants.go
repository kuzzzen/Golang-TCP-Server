package main

const (
	ServerMove               = "102 MOVE\a\b"
	ServerTurnLeft           = "103 TURN LEFT\a\b"
	ServerTurnRight          = "104 TURN RIGHT\a\b"
	ServerPickUp             = "105 GET MESSAGE\a\b"
	ServerLogout             = "106 LOGOUT\a\b"
	ServerOk                 = "200 OK\a\b"
	ServerLoginFailed        = "300 LOGIN FAILED\a\b"
	ServerSyntaxError        = "301 SYNTAX ERROR\a\b"
	ServerLogicError         = "302 LOGIC ERROR\a\b"
	ClientRecharging         = "RECHARGING\a\b"
	ClientFullPower          = "FULL POWER\a\b"
	ServerKeyRequest         = "107 KEY REQUEST\a\b"
	ServerKeyOutOfRangeError = "303 KEY OUT OF RANGE\a\b"
	MaxUsernameLen           = 20
	MaxKeyIdLen              = 5
	MaxConfirmationLen       = 7
	MaxOkLen                 = 12
	MaxMessageLen            = 100
)

type keyMap = map[byte]int

var keyPairs = [5]keyMap{
	{'s': 23019, 'c': 32037},
	{'s': 32037, 'c': 29295},
	{'s': 18789, 'c': 13603},
	{'s': 16443, 'c': 29533},
	{'s': 18189, 'c': 21952},
}

type Direction int

const (
	LEFT Direction = iota
	UP
	RIGHT
	DOWN
	UNDEFINED
)
