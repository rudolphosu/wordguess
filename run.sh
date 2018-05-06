cd chaincode-docker-devmode

open -a Terminal.app "docker_up.sh"
sleep 20
open -a Terminal.app "build_chaincode.sh"
open -a Terminal.app "use_cc.sh"
