#!/bin/bash

# Check if an argument was passed
if [ $# -eq 0 ]
then
  echo "Please provide a directory path as an argument."
  exit 1
fi

# Set the directory path to the first argument passed
directory=$1

# Find all the folders in the directory
folders=$(find $directory -type d)

# Loop through each folder
for folder in $folders
do
  echo "checking $folder for main.go file"
  # Check if main.go file exists in the folder
  if [ -e "$folder/main.go" ]
  then
    # Change to the folder
    cd $folder

    # Run the go command
    go run main.go

    # Return to the original directory
    cd $directory
  else
    echo "main.go not found in $folder"
  fi
done
