package main

type Player struct { 
    Name string
    Channel chan toPlayer
    Location *Room
    Salt []byte
    Hash []byte
}

type toPlayer struct {
    Text string
}

type fromPlayer struct {
    currentPlayer *Player 
    currentCommand string 
}

type Command struct {
	Verb string
	Function func(player1 *Player, userInput []string) string
}

type Zone struct {
    ID    int 
    Name  string
    Rooms []*Room
}

type Room struct {
    ID          int 
    Zone        *Zone
    Name        string
    Description string
    Exits       [6]Exit
}

type Exit struct {
    To          *Room
    Description string
}