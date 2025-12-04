#!/bin/bash

# Load testing script using Vegeta
# Install vegeta first: go install github.com/tsenart/vegeta@latest

set -e

# Configuration
TARGET_URL="${TARGET_URL:-http://localhost:3000/graphql}"
RATE="${RATE:-50}"           # Requests per second
DURATION="${DURATION:-1m}"   # Test duration
OUTPUT_FILE="${OUTPUT_FILE:-loadtest_results.bin}"

echo "Starting load test..."
echo "Target: $TARGET_URL"
echo "Rate: $RATE requests/second"
echo "Duration: $DURATION"
echo ""

# Create a GraphQL query payload
PAYLOAD='{"query":"{ artists { id name currentPrice } }"}'

# Run the attack
echo "POST $TARGET_URL" | \
  vegeta attack \
    -rate=$RATE \
    -duration=$DURATION \
    -body="$PAYLOAD" \
    -header="Content-Type: application/json" \
    -output=$OUTPUT_FILE

echo ""
echo "Attack completed. Generating report..."
echo ""

# Generate and display report
vegeta report $OUTPUT_FILE

echo ""
echo "Detailed metrics:"
vegeta report -type=json $OUTPUT_FILE | jq '.'

echo ""
echo "P95 latency:"
vegeta report $OUTPUT_FILE | grep "latency" | awk '{print $5}'

# Check if P95 is under 200ms
P95=$(vegeta report $OUTPUT_FILE | grep "latency" | awk '{print $5}' | sed 's/ms//')
if (( $(echo "$P95 < 200" | bc -l) )); then
  echo "✅ P95 latency ($P95 ms) is under 200ms - PASS"
  exit 0
else
  echo "❌ P95 latency ($P95 ms) exceeds 200ms - FAIL"
  exit 1
fi

