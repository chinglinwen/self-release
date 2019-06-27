go build
pkill self-release
# ./self-release -w http://wechat-notify.devops.haodai.net -agentid 1000003 -secret G5h7CTEqkBw-Fe3luf2JM8UNNJAcYTpbXvpveY7M3lg &> s.log &
./self-release &> s.log &
tail -f s.log
