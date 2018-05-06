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
