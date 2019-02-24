package main

import (
	"strconv"
	"fmt"
)

func addCommand(command string, action func(player1 *Player, userInput []string) string)  {
	cmd := Command {
		Verb: command,
		Function: action,
	}
	// cmdList = append(cmdList, cmd)
	cmdMap[command] = cmd
}

func makeCommands(){ // add these in order of significance
	addCommand("sigh", doSigh)
	addCommand("n", doNorth)
	addCommand("s", doSouth)
	addCommand("e", doEast)
	addCommand("w", doWest)
	addCommand("u", doUp)
	addCommand("d", doDown)
	addCommand("north", doNorth)
	addCommand("south", doSouth)
	addCommand("east", doEast)
	addCommand("west", doWest)
	addCommand("up", doUp)
	addCommand("down", doDown)
	addCommand("look", doLook)
	addCommand("recall", doRecall)
	addCommand("say", doSay)
	addCommand("shout", doShout)
	addCommand("where", doWhere)
}

func giveStatus(player1 *Player) string{
	text := ""
	tempDirections := map[int]string {0:"n", 1:"e", 2:"w", 3:"s", 4:"u", 5:"d"}
	// tell user where they are and what they can do
	text += (player1.Location.Name + "\n")
	text += (player1.Location.Description + "\n")
	// loop through players and see if any are in the room

	for key, _ := range playerMap {
		if playerMap[key].Location.ID == player1.Location.ID && player1.Name != key && playerMap[key].Channel != nil { // they are in the same room
			text += playerMap[key].Name + " is in the room \n" 
		}
	}
	text += "[ "
	for i, v := range player1.Location.Exits {
		if v.To != nil {
			text += (tempDirections[i] + " ")
		}
	}
	text += "]"
	return text + "\n >"
}

// functions 
func doSigh(player1 *Player, userInput []string) string  {
	notifyInRoom(player1, player1.Name + " has sighed \n > ")
	return ("you have sighed \n >")
}

func doNorth(player1 *Player, userInput []string)  string {
	text := ""
	if player1.Location.Exits[0].To != nil {
		player1.Location = player1.Location.Exits[0].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go north \n")
		text += giveStatus(player1)
	} else {
		text = ("you cannot go that way \n >")
	}
	return text
}

func doEast(player1 *Player, userInput []string) string {
	text := ""
	if player1.Location.Exits[1].To != nil {
		player1.Location = player1.Location.Exits[1].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go east \n")
		text += giveStatus(player1)
	} else {
		text += ("you cannot go that way \n >")
	}
	return text
}

func doWest(player1 *Player, userInput []string) string {
	text := ""
	if player1.Location.Exits[2].To != nil {
		player1.Location = player1.Location.Exits[2].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go west \n")
		text += giveStatus(player1)
	} else {
		text += ("you cannot go that way \n >")
	}
	return text
}

func doSouth(player1 *Player, userInput []string) string {
	text := ""
	if player1.Location.Exits[3].To != nil {
		player1.Location = player1.Location.Exits[3].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go south \n")
		text += giveStatus(player1)
	} else {
		text += ("you cannot go that way\n >")
	}
	return text
}

func doUp(player1 *Player, userInput []string)  string {
	text := ""
	if player1.Location.Exits[4].To != nil {
		player1.Location = player1.Location.Exits[4].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go up \n")
		text += giveStatus(player1)
	} else {
		text += ("you cannot go that way \n >")
	}
	return text
}

func doDown(player1 *Player, userInput []string) string {
	text := "" 
	if player1.Location.Exits[5].To != nil {
		player1.Location = player1.Location.Exits[5].To
		notifyInRoom(player1, player1.Name + " has entered the room \n")
		text += ("you go down \n")
		text += giveStatus(player1)
	} else {
		text += ("you cannot go that way \n >")
	}
	return text
}

func doLook(player1 *Player, userInput []string) string {
	// if userInput is just "look"
	// listPlayers()
	text := ""
	if len(userInput) == 1 {
		text +=  giveStatus(player1)
	} else if len(userInput) == 2 {
	// if user input is "look north", "look south", etc..
		if userInput[1] == "north" || userInput[1] == "n" {
			if player1.Location.Exits[0].Description != "" {
				text += (player1.Location.Exits[0].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way")
			}
		} else if userInput[1] == "east" || userInput[1] == "e" {
			if player1.Location.Exits[1].Description != "" {
				text += (player1.Location.Exits[1].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way")
			}
		} else if userInput[1] == "west" || userInput[1] == "w" {
			if player1.Location.Exits[2].Description != "" {
				text += (player1.Location.Exits[2].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way")
			}
		} else if userInput[1] == "south" || userInput[1] == "s" {
			if player1.Location.Exits[3].Description != "" {
				text += (player1.Location.Exits[3].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way")
			}
		} else if userInput[1] == "up" || userInput[1] == "u" {
			if player1.Location.Exits[4].Description != "" {
				text += (player1.Location.Exits[4].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way")
			}
		} else if userInput[1] == "down" || userInput[1] == "d" {
			if player1.Location.Exits[5].Description != "" {
				text += (player1.Location.Exits[5].Description + "\n >")
			} else{
				text += ("There's nothing interesting that way" + "\n >")
			} 
		} else {
			text += ("That's not a direction" + "\n >")
		} 
	} else{
		text +=  ("Enter a valid direction. i.e. 'look north' or 'look n' " + "\n >")
	}
	return text
}

func doRecall(player1 *Player, userInput []string) string {
	player1.Location = roomsMap[3001]
	return giveStatus(player1)
}

func doSay(player1 *Player, userInput []string) string {
	message := ""
	for i := 1; i < len(userInput); i++ {
		message += (userInput[i] + " ")
	}
	message += "\n"
	for key, _ := range playerMap {
		if playerMap[key].Location.ID == player1.Location.ID && player1.Name != playerMap[key].Name && playerMap[key].Channel != nil{ // they are in the same room
			message1 := (player1.Name + " says: " + message) 
			to_player := toPlayer {
				Text: message1,
			}
			playerMap[key].Channel <- to_player
		}
	}
	return "You said: " + message + "\n > "
}

func doShout(player1 *Player, userInput []string) string {
	message := ""
	for i := 1; i < len(userInput); i++ {
		message += (userInput[i] + " ")
	}
	message += "\n"
	for key, _ := range playerMap {
		if playerMap[key].Location.Zone.ID == player1.Location.Zone.ID && player1.Name != playerMap[key].Name && playerMap[key].Channel != nil { // they are in the same room
			message1 := (player1.Name + " says: " + message) 
			to_player := toPlayer {
				Text: message1,
			}
			playerMap[key].Channel <- to_player
		}
	}
	return "You said: " + message + "\n > "
}

func doWhere(player1 *Player, userInput []string) string {
	// loop over players and find the ones in the same zone
	message := ""
	for key, _ := range playerMap {
		if playerMap[key].Location.Zone.ID == player1.Location.Zone.ID && player1.Name != playerMap[key].Name && playerMap[key].Channel != nil{ // they are in the same room
			message += (playerMap[key].Name +  " is in Room " + strconv.Itoa(playerMap[key].Location.ID) + ": " + playerMap[key].Location.Name )
			
		}
	}
	return message + "\n >"
}

// helper
func notifyInRoom(player1 *Player, message string) { // see if there are any people in the room you just arrived in
	for key, _ := range playerMap {
		if player1.Location.ID == playerMap[key].Location.ID && player1.Name != key && playerMap[key].Channel != nil {
			to_player := toPlayer {
				Text: message,
			}
			playerMap[key].Channel <- to_player
		}
	}
}

func listPlayers(){
	for key, _ := range playerMap {
		fmt.Println(playerMap[key])
	}
}
