# kill all children once script exits
# https://stackoverflow.com/questions/360201/how-do-i-kill-background-processes-jobs-when-my-shell-script-exits
trap "exit" INT TERM
trap "kill 0" EXIT

./cloud -id 'node1' -name 'Node 1' &
sleep 2
./cloud -id 'node2' -name 'Node 2' -network ':9000' -port '9001' &

while true : 
do 
    echo "working"
    sleep 100
done
