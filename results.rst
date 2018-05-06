Nolan Rudolph 4/12/2018
-----------------------
Sample Program transactions - Successful game
=============================================

Install chaincode on channel
----------------------------
  transaction: peer chaincode install -p chaincodedev/chaincode/word_guess -n mycc -v 0
  response: Installed remotely response:<status:200 payload:"OK" >

Instantiate ledger + initiate first game
----------------------------
  transaction: peer chaincode instantiate -n mycc -v 0 -c '{"Args":["alice","bob","10","15"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["initGame", "alice","bob", "5","10"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"\001"

Place "positions" - secret word + hint
----------------------------
  transaction: peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","alice", "lakes","__kes"]}' -C myc
  response: status:200 payload:"Position placed - waiting on other player"

  transaction: peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","bob", "bikes","_ike_"]}' -C myc
  response: status:200 payload:"Both positions placed - make guesses"

Make guesses
----------------------------
  transaction: peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","alice", "pokes"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"Guess made - waiting on other player"

  transaction: peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","bob", "makes"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"Both guesses made - reveal secret word"

  -OR (players will tie)-

  transaction: peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","alice", "bikes"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","bob", "lakes"]}' -C myc


Players reveal secret words
----------------------------
  transaction: peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "lakes"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"Secret Word Revealed - waiting on other player"

  transaction: peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "bikes"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"Both guesses made - settle game"

Settle game
---------------------------
  transaction: peer chaincode invoke -n mycc -c '{"Args":["settleGame","1"]}' -C myc
  response: Chaincode invoke successful. result: status:200 payload:"Player 2 wins"

  -OR-

  response: Chaincode invoke successful. result: status:200 payload:"Tie"

Querying results
---------------------------
  Query: peer chaincode query -n mycc -c '{"Args":["queryGame","1"]}' -C myc
  Query Result: {"Player1_ID":"alice","Player2_ID":"bob","Player1_Bet":5,"Player2_Bet":15,"Player1_Word_Hash":"43a485720e033b469315feeedc5e9f033bbfe64a9427f7845a7f42c2eb8f6ab7","Player2_Word_Hash":"93253ae00ba9bef8a771a944c02877da35201ab6b148bbebb4a679fdaaa4dac2","Player1_Hint":"__kes","Player2_Hint":"_ike_","Player1_Guess":"pokes","Player2_Guess":"makes","Player1_Word":"lakes","Player2_Word":"bikes","State":"Player 2 wins"}

  Query: peer chaincode query -n mycc -c '{"Args":["queryPlayer","alice"]}' -C myc
  Query Result: {"Balance":5}

  Query: peer chaincode query -n mycc -c '{"Args":["queryPlayer","bob"]}' -C myc
  Query Result: {"Balance":20}

  -OR-

  Query Result: {"Player1_ID":"alice","Player2_ID":"bob","Player1_Bet":5,"Player2_Bet":15,"Player1_Word_Hash":"43a485720e033b469315feeedc5e9f033bbfe64a9427f7845a7f42c2eb8f6ab7","Player2_Word_Hash":"93253ae00ba9bef8a771a944c02877da35201ab6b148bbebb4a679fdaaa4dac2","Player1_Hint":"__kes","Player2_Hint":"_ike_","Player1_Guess":"bikes","Player2_Guess":"lakes","Player1_Word":"lakes","Player2_Word":"bikes","State":"Tie"}
  Query Result: {"Balance":10}
  Query Result: {"Balance":15}

Sample Errors:
==============
Improper balance
----------------
User tries to bet more than their current balance
  transaction: peer chaincode invoke -n mycc -c '{"Args":["initGame", "alice","bob", "15","10"]}' -C myc
  response: chaincode error (status: 500, message: Error Player 1: Insufficient Funds)

Invalid hint
----------------
  transaction: peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","bob", "bikes","_zzz_"]}' -C myc
  response: chaincode error (status: 500, message: Invalid hint - must match 3 letters of secret word in correct position)

Actions Out of Sequence
----------------
If user tries to make a guess transaction ahead of time
  transaction: peer chaincode invoke -n mycc -c '{"Args":["makeGuess", "1","alice", "pokes"]}' -C myc
  response: chaincode error (status: 500, message: Out of sequence error - not accepting guesses)

More malicious: user tries to change initial word and hint after moving on in game
  transaction: peer chaincode invoke -n mycc -c '{"Args":["placePosition", "1","alice", "lakes","__kes"]}' -C myc
  response: chaincode error (status: 500, message: Out of sequence error - not accepting positions)

Invalid word reveal
----------------
Player 2 invalid reveal:
  transaction: peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "lakes"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "pikes"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["settleGame","1"]}' -C myc
  result: Chaincode invoke successful. result: status:200 payload:"Player 1 wins"

Both invalid reveal:
  transaction: peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","alice", "sakes"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["revealSecretWord", "1","bob", "pikes"]}' -C myc
               peer chaincode invoke -n mycc -c '{"Args":["settleGame","1"]}' -C myc
  result: Chaincode invoke successful. result: status:200 payload:"Tie"
