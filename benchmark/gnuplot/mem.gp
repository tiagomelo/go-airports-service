##
# gnuplot script to generate a graphic for total process memory usage with smooth lines.
#
# It expects two parameters:
#
# csv_file_path       - path to the CSV file containing the data
# graphic_file_name   - the output PNG file name for the generated graph
#
# Expected CSV file columns:
#   1: second (elapsed time in seconds)
#   2: total_real_mem_mb (Total Real Memory used)
#
# Author: Tiago Melo (tiagoharris@gmail.com)
##

# Set the output as a PNG image of size 800x600 pixels
set terminal png size 800,600

# Set the file name for the output graphic
set output graphic_file_name

# Enable grid lines on the plot
set grid

# Specify that the CSV file is comma-separated
set datafile separator ","

# Remove time-based x-axis settings since we use a simple numeric counter
unset xdata
unset timefmt

# Set the title of the graph
set title "Total Memory Usage (MB) Over Time"

# Label the axes
set xlabel "Time (seconds)"
set ylabel "Memory Usage (MB)"

# Draw a box around the legend and place it in the upper right corner
set key box
set key right

# Define the style for the total real memory line in red
set style line 1 lc rgb '#FF0000' lt 1 lw 2 pt 7 pi -1 ps 1.5

# Force Y-axis range to avoid flat graphs
stats csv_file_path using 2 nooutput
y_min = STATS_min - 2   # Add padding below minimum
y_max = STATS_max + 2   # Add padding above maximum
set yrange [y_min:y_max]

# Use 'smooth csplines' to create smoother curves
plot csv_file_path using 1:2 title 'Total Real Memory (MB)' with lines ls 1 smooth csplines
