#!/bin/bash

cd "$(dirname "$0")"

for enum in n c s; do
    echo "enum = $enum..."
    for k in 05 10 20; do
        echo "k = $k..."
        {
            echo "hops send recv"
            for hops in {2..64}; do
                echo "$hops $(go run message_size.go $k $hops $enum)"
            done
        } > data/message_sizes_${k}_${enum} &
    done
done
wait
