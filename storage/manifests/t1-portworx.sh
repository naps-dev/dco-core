#!/bin/bash

# Array containing the ordered list of YAML files
yaml_files=("01-px-operator.yaml" "02-px-stc.yaml" "03-px-sc.yaml")

# Loop through the array and execute each file
for file in "${yaml_files[@]}"
do
    # Apply the YAML configuration using kubectl
    echo "Applying $file ..."
    kubectl apply -f $file

    # Check if command was successful
    if [ $? -eq 0 ]
    then
        echo "$file applied successfully."
    else
        echo "Failed to apply $file. Exiting."
        exit 1
    fi

    # Wait for 1 minute before applying the next file
    echo "Waiting for 1 minute before applying the next file..."
    sleep 60
done

echo "All files applied successfully."

