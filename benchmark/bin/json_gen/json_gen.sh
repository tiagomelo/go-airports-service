#!/bin/bash

##
# Bash script that generates a JSON payload for airports.
#
# Author: Tiago Melo (tiagoharris@gmail.com)
##

# Validate input parameters.
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <number_of_airports> <output_file>"
    exit 1
fi

NUM_AIRPORTS=$1
OUTPUT_FILE=$2

# Ensure NUM_AIRPORTS is an integer.
if ! [[ "$NUM_AIRPORTS" =~ ^[0-9]+$ ]]; then
    echo "Error: <number_of_airports> must be a positive integer."
    exit 1
fi

# Start message.
echo "Generating ${NUM_AIRPORTS} airports in ${OUTPUT_FILE}..."

# Write opening bracket.
echo "[" > "$OUTPUT_FILE"

# Generate airports.
for ((i=1; i<=NUM_AIRPORTS; i++)); do
    if [ "$i" -lt "$NUM_AIRPORTS" ]; then
        echo "{ \"name\": \"Airport${i}\", \"city\": \"City${i}\", \"country\": \"Country${i}\", \"iata_code\": \"AP${i}\" }," >> "$OUTPUT_FILE"
    else
        echo "{ \"name\": \"Airport${i}\", \"city\": \"City${i}\", \"country\": \"Country${i}\", \"iata_code\": \"AP${i}\" }" >> "$OUTPUT_FILE"
    fi
done

# Append closing bracket.
echo "]" >> "$OUTPUT_FILE"

# End message.
echo "Generation completed. File saved as ${OUTPUT_FILE}"
