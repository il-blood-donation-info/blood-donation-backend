#!/bin/sh

# Define variables
total_duration=30  # Total duration in seconds
interval=5  # Interval in seconds
end_time=$(($(date +%s) + total_duration))

# Run the test in a loop
while [ $(date +%s) -lt $end_time ]; do
    result=$(wget -q -O- --no-check-certificate https://blood-info:8443/stations | jq '. | length')
    if [ -n "$result" ] && [ "$result" -ne 0 ]; then
        echo "Test passed: There are $result stations reported."
        exit 0
    fi

    sleep $interval  # Wait for the specified interval before the next test
done

echo "Test failed: Result is empty or zero."
exit 1  # Test failed
