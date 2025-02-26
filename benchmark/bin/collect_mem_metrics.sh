#!/bin/bash

##
# Bash script that collects memory usage.
#
# Author: Tiago Melo (tiagoharris@gmail.com)
##

# validate input parameters.
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <output_file>"
    exit 1
fi

output_file="$1"

# write CSV header.
echo "second,total_mem_mb" > "$output_file"

counter=0

while true; do
  # get PID of the process listening on port 4444.
  pid=$(lsof -ti tcp:4444 | head -n 1)

  if [[ -z "$pid" ]]; then
    echo "$counter,0" >> "$output_file"  # Log zero if process not found
    sleep 1
    ((counter++))
    continue
  fi

  # get full `top` output.
  top_output=$(top -l 1 -pid "$pid")

  # dynamically find the MEM column position.
  mem_col=$(echo "$top_output" | awk 'NR==1 {for (i=1; i<=NF; i++) if ($i == "MEM") print i}')

  # if column not found, assume 8th column as fallback.
  if [[ -z "$mem_col" ]]; then
    mem_col=8
  fi

  # extract memory usage from `top`.
  mem_value=$(echo "$top_output" | awk -v col="$mem_col" -v pid="$pid" '$1 == pid {print $col}')

  # convert MEM to MB.
  case "$mem_value" in
    *G) total_mem_mb=$(echo "$mem_value" | sed 's/G//' | awk '{printf "%.0f", $1 * 1024}') ;;
    *M) total_mem_mb=$(echo "$mem_value" | sed 's/M//') ;;
    *K) total_mem_mb=$(echo "$mem_value" | sed 's/K//' | awk '{printf "%.0f", $1 / 1024}') ;;
    *) total_mem_mb=0 ;;
  esac

  # ensure valid number.
  if ! [[ "$total_mem_mb" =~ ^[0-9]+$ ]]; then
    total_mem_mb=0
  fi

  # append to CSV.
  echo "$counter,$total_mem_mb" >> "$output_file"

  # increment counter.
  ((counter++))

  # wait 1 sec before next reading.
  sleep 1
done
