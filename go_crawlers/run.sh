#!/bin/bash

# Get the directory path from the user
read -p "Enter the directory path: " directory

# Find all the folders in the directory
folders=$(find $directory -type d)

# Loop through each folder
for folder in $folders
do
  echo "cheking $folder for main.go file"
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

