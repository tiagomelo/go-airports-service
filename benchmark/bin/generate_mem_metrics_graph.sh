#!/usr/bin/env bash

##
# Bash script that generates a graphic from a CSV file (using gnuplot).
#
# Author: Tiago Melo (tiagoharris@gmail.com)
##

GNUPLOT_SCRIPT_FILE="../gnuplot/mem.gp"

# validate input parameters.
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <csv_file> <graphic_output_file>"
    exit 1
fi

CSV_FILE="$1"
GNUPLOT_GRAPHIC_FILE="$2"

function generateGraphic {
  gnuplot -e "csv_file_path='$CSV_FILE'" -e "graphic_file_name='$GNUPLOT_GRAPHIC_FILE'" $GNUPLOT_SCRIPT_FILE 
}

generateGraphic
exit
