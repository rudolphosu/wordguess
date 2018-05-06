# wordguess
Word guess game implemented with hyperledger fabric

Rules:
* The game is between two players. Before the games starts, each player places a bet on winning the game.
* Each player can bet any amount and both players do not have to bet the same amount  
* Each player challenges the other player to guess a five-letter word, revealing only three of the letters and their positions in the word.
* For example, “ _OUS_” for “HOUSE”Each player has one try to guess the other player’s wordAfter both players have made their guesses, each player must reveal the word used in the challenge and other player gets 1 point for each letter correctly guessed.
* A winner is declared based on which player has the most points.
* If there is a tie, a tie is declared. The winner receives that total amount bet by both players.
* If there is a tie, both players receive their bets back.

Chaincode is found in chaincode/word_guess/word_guess.go

To run, follow the instructions found here, but replace the commands for terminal 2 and 3 with those below:
https://hyperledger-fabric.readthedocs.io/en/release-1.1/chaincode4ade.html#testing-using-dev-mode

Terminal 2:
```
docker exec -it chaincode bash
cd word_guess
go build -o word_guess
CORE_PEER_ADDRESS=peer:7052 CORE_CHAINCODE_ID_NAME=mycc:0 ./word_guess
```

Terminal 3:
```
docker exec -it cli bash

#install and instantiate
peer chaincode install -p chaincodedev/chaincode/word_guess -n mycc -v 0

peer chaincode instantiate -n mycc -v 0 -c '{"Args":["alice","bob","10","15"]}' -C myc

#initialize game - invalid bet
peer chaincode invoke -n mycc -c '{"Args":["initGame", "alice","bob", "15","10"]}' -C myc
#valid bet
peer chaincode invoke -n mycc -c '{"Args":["initGame", "alice","bob", "5","15"]}' -C myc

#players place positions - invalid hint
peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","bob", "bikes","_zzz_"]}' -C myc
#valid
peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","alice", "lakes","__kes"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","bob", "bikes","_ike_"]}' -C myc

#players make guesses
peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","alice", "pokes"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","bob", "makes"]}' -C myc

#players reveal secret words - one invalid
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "lakes"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "pikes"]}' -C myc
#both invalid
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "sakes"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "pikes"]}' -C myc
#valid
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "lakes"]}' -C myc
peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "bikes"]}' -C myc

#settle game
peer chaincode invoke -n mycc -c '{"Args":["settleGame","1"]}' -C myc

#query game
peer chaincode query -n mycc -c '{"Args":["queryGame","1"]}' -C myc

#query Player
peer chaincode query -n mycc -c '{"Args":["queryPlayer","alice"]}' -C myc

```
