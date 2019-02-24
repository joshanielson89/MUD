package main

import (
	"bufio"
	"fmt"

	// "os"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"log"
	"net"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

var cmdMap map[string]Command
var playerMap map[string]*Player
var zonesMap map[int]*Zone
var roomsMap map[int]*Room

func main() {
	// initializations
	fmt.Println("Joshua Nielson - CS3410 MUD")
	// open database
	db, err := sql.Open("sqlite3", "world.db")
	if err != nil {
		fmt.Errorf("Error opening db %e", err)
	}
	defer db.Close()
	fmt.Println("Reading in world file")
	// begin tx for reading in Zones
	tx, err := db.Begin()
	if err != nil {
		fmt.Errorf("Error begining zones tx %e", err)
	}
	zonesMap, err = readZones(tx)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	// begin for reading in rooms
	tx, err = db.Begin()
	if err != nil {
		fmt.Errorf("Error begining rooms tx %e", err)
	}
	roomsMap, err = readRooms(tx, zonesMap)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	tx, err = db.Begin()
	if err != nil {
		fmt.Errorf("Error begining exits tx %e", err)
	}
	err = readExits(tx, roomsMap)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
		fmt.Printf("Read in %d zones and %d rooms \n", len(zonesMap), len(roomsMap))
	}

	playerMap = make(map[string]*Player)
	tx, err = db.Begin()
	if err != nil {
		fmt.Errorf("Error reading in players tx %e", err)
	}
	err = readPlayers(tx, playerMap)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
		fmt.Printf("Read in %d players \n", len(playerMap))
	}

	// call makeCommands to intialize all commands
	fmt.Println("Installing commands")
	cmdMap = make(map[string]Command)
	makeCommands()
	fmt.Printf("Intalled %d commands \n \n", len(cmdMap))

	from_player := make(chan fromPlayer)
	go handleConnections(from_player)

	for event := range from_player {
		// process events here from players
		if event.currentPlayer.Channel != nil {
			// handle event
			cSlice := strings.Fields(event.currentCommand)
			if cSlice[0] == "quit" || cSlice[0] == "Quit" {
				// close that channel
				if event.currentPlayer.Channel != nil {
					close(event.currentPlayer.Channel)
					event.currentPlayer.Channel = nil
				}
				continue
			}
			text := dispatch(event.currentPlayer, cSlice)
			response := toPlayer{
				Text: text,
			}
			event.currentPlayer.Channel <- response
		} else {
			// log the error and ignore event, and remove the player from playerList
			fmt.Println("player channel was nil.  Event not dispatched")
		}
	}
	return
}

func handleConnections(from_player chan fromPlayer) { // handle each new telnet connection
	fmt.Println("Now accepting incoming connections")
	// check for new connections and do the following code for each new connection
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			// handle error
		}
		go go1(conn, from_player) // Player -> MUD
	}
}

// this will handle the individual player events: Player -> MUD
func go1(conn net.Conn, from_player chan fromPlayer) {
	// loop handle the event and send info back through channel using <-
	// make Player here
	to_player := make(chan toPlayer)
	go go2(conn, to_player) // MUD -> PLayer

	// ask for username and pass
	var name string
	var password string
	// find players name
	fmt.Fprintf(conn, "Please enter your name: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	userInput := string(scanner.Text())
	if userInput != "" {
		name = userInput
		// fmt.Fprintf(conn, "got username")
	} else {
		fmt.Fprintf(conn, "Player entered empty username")
	}
	fmt.Fprintf(conn, "Password: ")
	scanner.Scan()
	userInput = string(scanner.Text())
	if userInput != "" {
		password = userInput
	} else {
		fmt.Fprintf(conn, "Player entered empty password")
	}
	newSalt := makeSalt()
	newHash := makeHash(password, newSalt)

	// make a new player
	player1 := &Player{
		Name:     name,
		Location: roomsMap[3001],
		Channel:  to_player,
		Salt:     newSalt,
		Hash:     newHash,
	}

	// if player is already in the list, log him in
	if _, ok := playerMap[name]; ok {
		// check username and password
		checkHash := makeHash(password, playerMap[name].Salt)
		realHash := playerMap[name].Hash
		if subtle.ConstantTimeCompare(checkHash, realHash) != 1 {
			fmt.Println("login failed: Invalid credentials")
			close(to_player)
			playerMap[name].Channel = nil
		} else {
			fmt.Println("Login successful: " + playerMap[name].Name + " logged in.")
			player1 = playerMap[name]
			if playerMap[name].Channel != nil {
				close(playerMap[name].Channel)
				playerMap[name].Channel = nil
			}
			player1.Channel = to_player
			fmt.Fprintf(conn, "Now accepting commands: \n")
			fmt.Fprintf(conn, giveStatus(player1))
			for scanner.Scan() {
				// look for commands
				userInput := string(scanner.Text())
				if userInput != "" {
					event := fromPlayer{
						currentPlayer:  player1,
						currentCommand: userInput,
					}
					from_player <- event
				} else {
					fmt.Fprintf(conn, "commands was nil \n > ")
				}
			}
			if err := scanner.Err(); err != nil {
				// this means the connection failed
				// send a quit request to main loop
				fmt.Println("error in scanner")
			}
			event := fromPlayer{
				currentPlayer:  player1,
				currentCommand: "quit",
			}
			from_player <- event
			fmt.Println("Player was logged out")
		}

	} else { // otherwise create a new player and add them to the database
		// add new player to db
		db, err := sql.Open("sqlite3", "world.db")
		if err != nil {
			fmt.Errorf("Error opening db %e", err)
		}
		defer db.Close()
		tx, err := db.Begin()
		// make salt and hash
		salt64 := base64.StdEncoding.EncodeToString(newSalt)
		hash64 := base64.StdEncoding.EncodeToString(newHash)
		if err != nil {
			fmt.Errorf("Error adding player to db %e", err)
		}
		_, err = tx.Exec("INSERT INTO players(name, salt, hash) VALUES (?, ?, ?)", name, salt64, hash64)
		if err != nil {
			fmt.Println("hit error in Exec")
			// log.Fatal(err)
			tx.Rollback()
		} else {
			tx.Commit()
			fmt.Println("New player created")
			playerMap[name] = player1 // add new player to slice

			fmt.Fprintf(conn, "Now accepting commands: \n")
			fmt.Fprintf(conn, giveStatus(player1))
			for scanner.Scan() {
				// look for commands
				userInput := string(scanner.Text())
				if userInput != "" {
					event := fromPlayer{
						currentPlayer:  player1,
						currentCommand: userInput,
					}
					from_player <- event
				} else {
					fmt.Fprintf(conn, "commands was nil \n > ")
				}
			}
			if err := scanner.Err(); err != nil {
				// this means the connection failed
				// send a quit request to main loop
				event := fromPlayer{
					currentPlayer:  player1,
					currentCommand: "quit",
				}
				from_player <- event
			}
		}

	}
}

// this will handle events to all the players: MUD -> Players
func go2(conn net.Conn, to_player chan toPlayer) {
	// for loop processing commands from the MUD and process them here
	defer fmt.Println("connection closed: player logged out")
	defer conn.Close()
	if to_player != nil {
		for event := range to_player {
			fmt.Fprintf(conn, event.Text)
		}
	}
}

func readZones(tx *sql.Tx) (map[int]*Zone, error) {
	var zonesMap = make(map[int]*Zone)
	rows, err := tx.Query("select id, name from zones order by id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// for each row in rows
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		zone := new(Zone)
		zone.ID = id
		zone.Name = name
		zonesMap[id] = zone
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return zonesMap, err
}

func readRooms(tx *sql.Tx, zonesMap map[int]*Zone) (map[int]*Room, error) {
	// use zone_id to find corresponding zone in map of zones.
	// store that *Zone in Room object
	var roomMap = make(map[int]*Room)

	rows, err := tx.Query("select id, zone_id, name, description from rooms")
	if err != nil {
		return nil, fmt.Errorf("Error in readRooms() Query %e", err)
	}
	defer rows.Close()
	// for each row in rows
	for rows.Next() {
		var id int
		var zone_id int
		var name string
		var description string
		err = rows.Scan(&id, &zone_id, &name, &description)
		if err != nil {
			log.Fatal(err)
		}
		room := new(Room)
		room.ID = id
		room.Name = name
		room.Description = description
		room.Zone = zonesMap[zone_id]
		// for k, v := range zonesMap {	 // k = maps keys, v = value of key k
		// 	if k == id {
		// 		room.Zone = v  // finish making room
		// 		zonesMap[k].Rooms = append(zonesMap[k].Rooms, room)
		// 	}
		// }
		roomMap[id] = room

	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error in readRoom() rows.Next() %e", err)
	}
	return roomMap, err
}

func readExits(tx *sql.Tx, roomsMap map[int]*Room) error {
	directions := map[string]int{"n": 0, "e": 1, "w": 2, "s": 3, "u": 4, "d": 5}

	rows, err := tx.Query("select from_room_id, to_room_id, direction, description from Exits")
	if err != nil {
		return fmt.Errorf("Error querying exits in readExits() %e", err)
	}
	defer rows.Close()
	// for each row in rows
	for rows.Next() {
		var from_room int
		var to_room int
		var direction string
		var description string
		err = rows.Scan(&from_room, &to_room, &direction, &description)
		if err != nil {
			log.Fatal(err)
		}
		roomsMap[from_room].Exits[directions[direction]].To = roomsMap[to_room]
		roomsMap[from_room].Exits[directions[direction]].Description = description
	}
	err = rows.Err()
	if err != nil {
		fmt.Errorf("Error in rows.Next() in readExits() %e", err)
	}
	return err
}

func readPlayers(tx *sql.Tx, playerMap map[string]*Player) error {
	rows, err := tx.Query("select name, salt, hash from players")
	if err != nil {
		return fmt.Errorf("Error in readplayers() Query %e", err)
	}
	defer rows.Close()
	// for each row in rows
	for rows.Next() {
		var name string
		var salt string
		var hash string
		var salt64 []byte
		var hash64 []byte
		err = rows.Scan(&name, &salt, &hash)
		if err != nil {
			log.Fatal(err)
		}
		salt64, err = base64.StdEncoding.DecodeString(salt)
		if err != nil {
			log.Fatal(err)
		}
		hash64, err = base64.StdEncoding.DecodeString(hash)
		if err != nil {
			log.Fatal(err)
		}
		player1 := &Player{
			Name:     name,
			Location: roomsMap[3001],
			Salt:     salt64,
			Hash:     hash64,
		}
		//

		playerMap[name] = player1
	}
	return nil
}

func dispatch(player1 *Player, userInput []string) string {
	text := ""
	if _, ok := cmdMap[userInput[0]]; ok { // check if command exists
		text = cmdMap[userInput[0]].Function(player1, userInput)
		return text
	} else {
		return ("Huh? \n >")
	}
}

func makeSalt() []byte {
	salt := make([]byte, 32)
	_, err := crand.Read(salt)
	if err != nil {
		log.Fatal(err)
	}
	return salt
}

func makeHash(password string, salt []byte) []byte {
	hash := pbkdf2.Key([]byte(password), salt, 64*1024, 32, sha256.New)
	return hash
}
