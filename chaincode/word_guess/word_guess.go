package main

import (
	"fmt"
	"strconv"
	"encoding/json"
	"crypto/sha256"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type Game struct {
	Player1_ID string
	Player2_ID string
	Player1_Bet int
	Player2_Bet int
	Player1_Word_Hash string
	Player2_Word_Hash string
	Player1_Hint string
	Player2_Hint string
	Player1_Guess string
	Player2_Guess string
	Player1_Word string
	Player2_Word string
	State string
}

var game_id int //auto increment game id

type Player struct {
	Balance int
}

// Init is called during chaincode instantiation to initialize any data
func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	game_id = 1

	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 4 {
		return shim.Error("Incorrect arguments. Expecting (4): player 1, player 2 and their available funds")
	}

	// Test data for players and funds are being initialized
	test_player_1_id := args[0]
	test_player_2_id := args[1]
	test_player_1_funds,_ := strconv.Atoi(args[2])
	test_player_2_funds,_ := strconv.Atoi(args[3])
	test_player_1 := Player{Balance: test_player_1_funds}
	test_player_2 := Player{Balance: test_player_2_funds}

	//convert player structs to json
	player1AsJSON,_ := json.Marshal(test_player_1)
	player2AsJSON,_ := json.Marshal(test_player_2)

	// Add test players to the ledger by calling stub.PutState()
	err := stub.PutState(test_player_1_id, player1AsJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create Player: %s", test_player_1_id))
	}
	err = stub.PutState(test_player_2_id, player2AsJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create Player: %s", test_player_2_id))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode.
func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "queryGame" {
		result, err = queryGame(stub, args)
	} else if fn == "queryPlayer" {
		result, err = queryPlayer(stub, args)
	} else if fn == "initGame" {
		result, err = initGame(stub, args)
	} else if fn == "placePosition" {
		result, err = placePosition(stub, args)
	} else if fn == "makeGuess" {
		result, err = makeGuess(stub, args)
	} else if fn == "revealSecretWord" {
		result, err = revealSecretWord(stub, args)
	} else if fn == "settleGame" {
		result, err = settleGame(stub, args)
	} else {
		return shim.Error(fmt.Sprintf("Invalid chaincode function name"))
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Init Game will create a new game between the players specified
// A game ID will be returned, which will be used in subsequent transactions
func initGame(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("Incorrect arguments. Expecting (4): a first player ID, first player's bet, a second player ID, and second player's bet")
	}

	//get the args from the transaction proposal
	player1_id := args[0]
	player2_id := args[1]
	player1_bet,_ := strconv.Atoi(args[2])
	player2_bet,_ := strconv.Atoi(args[3])

	//check if player ids are valid
	player1AsJSON, err := stub.GetState(player1_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Player 1 ID")
	}
	player2AsJSON, err := stub.GetState(player2_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Player 2 ID")
	}
	player1 := Player{}
	json.Unmarshal(player1AsJSON, &player1)
	player2 := Player{}
	json.Unmarshal(player2AsJSON, &player2)

	//check if players have required funds for bets
	if player2_bet > player2.Balance{
		return "", fmt.Errorf("Error Player 2: Insufficient Funds")
	}	else if player1_bet > player1.Balance{
		return "", fmt.Errorf("Error Player 1: Insufficient Funds")
	} else {
		//if so, temporarily deduct bets from players balance to eliminate double spending in separate game
		player1.Balance = player1.Balance - player1_bet
		player2.Balance = player2.Balance - player2_bet
		player1AsJSON,_ := json.Marshal(player1)
		player2AsJSON,_ := json.Marshal(player2)
		stub.PutState(player1_id, player1AsJSON)
		stub.PutState(player2_id, player2AsJSON)
	}

	//Initialize new game, add to ledger
	new_game := Game{Player1_ID: player1_id, Player2_ID: player2_id, Player1_Bet: player1_bet, Player2_Bet: player2_bet, State: "Accepting positions"}
	gameAsJSON,_ := json.Marshal(new_game)
	err = stub.PutState(strconv.Itoa(game_id), gameAsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to create Game")
	}
	game_id++
	return string(game_id-1), nil
}

// Place position updates specified game with secret words and hints for a player
func placePosition(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a game id, player id, secret word, and hint")
	}

	//get the args from the transaction proposal
	game_id := args[0]
	player_id := args[1]
	secret_word := args[2]
	hint := args[3]

	returnMsg := ""

	//retrieve game from ledger
	gameAsJSON, err := stub.GetState(game_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Game ID provided")
	}
	game := Game{}
	json.Unmarshal(gameAsJSON, &game)

	//check game is in correct state
	if game.State != "Accepting positions"{
		return "", fmt.Errorf("Out of sequence error - not accepting positions")
	}

	if len(secret_word) != 5 && len(hint) != 5 {
		return "", fmt.Errorf("Invalid secret word and hint - must be of length 5")
	}

	//validate hint matches secret word
	for i, _ := range secret_word {
		if hint[i] != '_' && hint[i] != secret_word[i] {
			return "", fmt.Errorf("Invalid hint - must match 3 letters of secret word in correct position")
		}
	}

	//hash the secret word so it is not revealed in plain text on the ledger, but can be compared later
	word_hash := sha256.Sum256([]byte(secret_word))

	//validate player id, update game, adding players secret word and hint
	//if both players have placed positions, update game state to accepting guesses
	if player_id == game.Player1_ID{
		game.Player1_Word_Hash = fmt.Sprintf("%x",word_hash)
		game.Player1_Hint = hint
		if game.Player2_Hint != ""{
			game.State = "Accepting guesses"
			returnMsg = "Both positions placed - make guesses"
		} else {
			returnMsg = "Position placed - waiting on other player"
		}
	} else if player_id == game.Player2_ID{
		game.Player2_Word_Hash = fmt.Sprintf("%x",word_hash)
		game.Player2_Hint = hint
		if game.Player1_Hint != ""{
			game.State = "Accepting guesses"
			returnMsg = "Both positions placed - make guesses"
		} else {
			returnMsg = "Position placed - waiting on other player"
		}
	} else {
		return "", fmt.Errorf("Invalid Player ID provided")
	}

	//add updated game to ledger
	gameAsJSON,_ = json.Marshal(game)
	err = stub.PutState(game_id, gameAsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Game")
	}

	return returnMsg, nil
}

// Query Game returns the values for the specified game id
func makeGuess(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a game id, player id, and word guess")
	}

	//get the args from the transaction proposal
	game_id := args[0]
	player_id := args[1]
	guess := args[2]

	returnMsg := ""

	//retrieve game from ledger
	gameAsJSON, err := stub.GetState(game_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Game ID provided")
	}
	game := Game{}
	json.Unmarshal(gameAsJSON, &game)

	//check game is in correct state
	if game.State != "Accepting guesses"{
		return "", fmt.Errorf("Out of sequence error - not accepting guesses")
	}

	//validate player id, update game, adding players guess
	//if both players have made guesses, update game state to settle game
	if player_id == game.Player1_ID{
		game.Player1_Guess = guess
		if game.Player2_Guess != ""{
			game.State = "Reveal answers"
			returnMsg = "Both guesses made - reveal secret word"
		} else {
			returnMsg = "Guess made - waiting on other player"
		}
	} else if player_id == game.Player2_ID{
		game.Player2_Guess = guess
		if game.Player1_Guess != ""{
			game.State = "Reveal answers"
			returnMsg = "Both guesses made - reveal secret word"
		} else {
			returnMsg = "Guess made - waiting on other player"
		}
	} else {
		return "", fmt.Errorf("Invalid Player ID provided")
	}

	//add updated game to ledger
	gameAsJSON,_ = json.Marshal(game)
	err = stub.PutState(game_id, gameAsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Game")
	}

	return returnMsg, nil
}

// Reveal SecretWord requires a user to reenter their secret word to ensure it has not been tampered with when settling game
func revealSecretWord(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a game id, player id, and original secret word")
	}

	//get the args from the transaction proposal
	game_id := args[0]
	player_id := args[1]
	secret_word := args[2]

	returnMsg := ""

	//retrieve game from ledger
	gameAsJSON, err := stub.GetState(game_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Game ID provided")
	}
	game := Game{}
	json.Unmarshal(gameAsJSON, &game)

	//check game is in correct state
	if game.State != "Reveal answers"{
		return "", fmt.Errorf("Out of sequence error - not revealing answers")
	}

	//validate player id, update game, adding players guess
	//if both players have made guesses, update game state to settle game
	if player_id == game.Player1_ID{
		game.Player1_Word = secret_word
		if game.Player2_Word != ""{
			game.State = "Settle game"
			returnMsg = "Both guesses made - settle game"
		} else {
			returnMsg = "Secret Word Revealed - waiting on other player"
		}
	} else if player_id == game.Player2_ID{
		game.Player2_Word = secret_word
		if game.Player1_Guess != ""{
			game.State = "Settle game"
			returnMsg = "Both guesses made - settle game"
		} else {
			returnMsg = "Secret Word Revealed- waiting on other player"
		}
	} else {
		return "", fmt.Errorf("Invalid Player ID provided")
	}

	//add updated game to ledger
	gameAsJSON,_ = json.Marshal(game)
	err = stub.PutState(game_id, gameAsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Game")
	}

	return returnMsg,nil
}

// Settle game tallies the points earned by each player, determines a winner or tie, and distributes funds accordingly
func settleGame(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a game id")
	}

	game_id := args[0]

	//retrieve game from ledger
	gameAsJSON, err := stub.GetState(game_id)
	if err != nil {
		return "", fmt.Errorf("Invalid Game ID provided")
	}
	game := Game{}
	json.Unmarshal(gameAsJSON, &game)

	//check game is in correct state
	if game.State != "Settle game"{
		return "", fmt.Errorf("Out of sequence error - not settling game")
	}

	//retrieve players from ledger
	player1AsJSON, err := stub.GetState(game.Player1_ID)
	if err != nil {
		return "", fmt.Errorf("Failed to get Player")
	}
	player2AsJSON, err := stub.GetState(game.Player2_ID)
	if err != nil {
		return "", fmt.Errorf("Failed to get Player")
	}
	player1 := Player{}
	json.Unmarshal(player1AsJSON, &player1)
	player2 := Player{}
	json.Unmarshal(player2AsJSON, &player2)

	//compare secret word hashes and determine if either player tampered with results
	//declare forfeit and reallocate funds as necessary
	//otherwise, determine winner by points earned
	player1_valid := game.Player1_Word_Hash == fmt.Sprintf("%x",sha256.Sum256([]byte(game.Player1_Word)))
	player2_valid := game.Player2_Word_Hash == fmt.Sprintf("%x",sha256.Sum256([]byte(game.Player2_Word)))
	if !player1_valid && !player2_valid{
		game.State = "Tie"
		player1.Balance = player1.Balance + game.Player1_Bet
		player2.Balance = player2.Balance + game.Player2_Bet
	} else if !player1_valid{
		game.State = "Player 2 wins"
		player2.Balance = player2.Balance + game.Player1_Bet + game.Player2_Bet
	} else if !player2_valid{
		game.State = "Player 1 wins"
		player1.Balance = player1.Balance + game.Player1_Bet + game.Player2_Bet
	} else {
		//tally player 1's points
		p1_points := 0
		for i,_ := range game.Player1_Guess{
			if game.Player1_Guess[i] == game.Player2_Word[i]{
				p1_points++
			}
		}

		//tally player 2's points
		p2_points := 0
		for i,_ := range game.Player2_Guess{
			if game.Player2_Guess[i] == game.Player1_Word[i]{
				p2_points++
			}
		}

		//compare points and determine outcome, reallocating funds as necessary
		if p1_points > p2_points{
			game.State = "Player 1 wins"
			player1.Balance = player1.Balance + game.Player1_Bet + game.Player2_Bet
		} else if p2_points > p1_points{
			game.State = "Player 2 wins"
			player2.Balance = player2.Balance + game.Player1_Bet + game.Player2_Bet
		} else {
			game.State = "Tie"
			player1.Balance = player1.Balance + game.Player1_Bet
			player2.Balance = player2.Balance + game.Player2_Bet
		}
	}

	//commit game result and new player balances to ledger
	player1AsJSON,_ = json.Marshal(player1)
	player2AsJSON,_ = json.Marshal(player2)
	err = stub.PutState(game.Player1_ID, player1AsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Player")
	}
	err = stub.PutState(game.Player2_ID, player2AsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Player")
	}
	gameAsJSON,_ = json.Marshal(game)
	err = stub.PutState(game_id, gameAsJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to update Game")
	}
	return game.State, nil
}

// Query Game returns the values for the specified game id
func queryGame(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a game id")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get game: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Game not found: %s", args[0])
	}
	return string(value), nil
}

// Query Player returns the values for the specified player id
func queryPlayer(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a player id")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get player: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Player not found: %s", args[0])
	}
	return string(value), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s", err)
	}
}
