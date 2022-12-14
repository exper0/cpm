#! /bin/sh

function foo() {
    echo "XX!"
}

trap foo TERM

while true
do
  sleep 60 & wait
done
