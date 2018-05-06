docker exec -it chaincode bash

cd word_guess

go build -o word_guess

CORE_PEER_ADDRESS=peer:7052 CORE_CHAINCODE_ID_NAME=mycc:0 ./word_guess
