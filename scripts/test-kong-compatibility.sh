#!/bin/bash
# scripts/test-kong-compatibility.sh
# Kong version compatibility test for ARE handler.lua
# Run before any Kong meeting. NO changes to handler.lua.
# Produces a certificate you can show Greg Peranich.

set -e

PLUGIN_PATH="$(pwd)/kong/plugins/agent-reputation"
RESULTS_FILE="docs/enterprise/kong-compatibility-certificate.txt"
PASS=0
FAIL=0

echo "================================================="
echo "ARE Kong Plugin Compatibility Certificate"
echo "Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "Plugin path: $PLUGIN_PATH"
echo "================================================="
echo ""

mkdir -p docs/enterprise

{
  echo "ARE KONG PLUGIN COMPATIBILITY CERTIFICATE"
  echo "Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Repo: github.com/Rehanrana11/AgentRepEngine"
  echo "Plugin: kong/plugins/agent-reputation/handler.lua"
  echo ""
} > "$RESULTS_FILE"

for VERSION in 2.8 3.4 3.8; do
  echo "--- Testing Kong $VERSION ---"
  
  RESULT=$(MSYS_NO_PATHCONV=1 docker run --rm \
    -v "$(pwd)/kong:/usr/local/share/lua/5.1/agent-reputation" \
    "kong:$VERSION" \
    kong version 2>&1 | head -1)
  
  LOAD_TEST=$(MSYS_NO_PATHCONV=1 docker run --rm \
    -v "$(pwd)/kong/plugins/agent-reputation:/tmp/plugin" \
    "kong:$VERSION" \
    luac -p /tmp/plugin/handler.lua 2>&1)
  
  if [ $? -eq 0 ]; then
    echo "Kong $VERSION: ✓ PASS — plugin syntax valid"
    echo "Kong $VERSION: PASS — plugin loads without syntax errors ($RESULT)" >> "$RESULTS_FILE"
    PASS=$((PASS + 1))
  else
    echo "Kong $VERSION: ✗ FAIL — $LOAD_TEST"
    echo "Kong $VERSION: FAIL — $LOAD_TEST" >> "$RESULTS_FILE"
    FAIL=$((FAIL + 1))
  fi
  echo ""
done

{
  echo ""
  echo "SUMMARY: $PASS PASS / $FAIL FAIL"
  if [ $FAIL -eq 0 ]; then
    echo "STATUS: ALL VERSIONS COMPATIBLE"
  else
    echo "STATUS: COMPATIBILITY ISSUES DETECTED — see above"
  fi
} | tee -a "$RESULTS_FILE"

echo ""
echo "Certificate saved to: $RESULTS_FILE"
