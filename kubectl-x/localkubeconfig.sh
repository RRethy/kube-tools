#!/bin/bash

# localkubeconfig.sh - Example script showing how to use kubectl-x kubeconfig copy
# This script demonstrates creating an isolated kubeconfig for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== kubectl-x kubeconfig copy Example ===${NC}"
echo ""
echo "This script demonstrates how to create an isolated kubeconfig file"
echo "that can be used with the KUBECONFIG environment variable."
echo ""

# Step 1: Copy the current kubeconfig to a local file
echo -e "${YELLOW}Step 1: Copying current kubeconfig to XDG_DATA_HOME...${NC}"
KUBECONFIG_PATH=$(kubectl x kubeconfig copy)

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Kubeconfig copied to: $KUBECONFIG_PATH${NC}"
else
    echo -e "${RED}✗ Failed to copy kubeconfig${NC}"
    exit 1
fi

# Step 2: Show how to use the copied kubeconfig
echo ""
echo -e "${YELLOW}Step 2: Using the isolated kubeconfig...${NC}"
echo ""
echo "You can now use this kubeconfig in several ways:"
echo ""
echo "1. Export it for the current shell session:"
echo -e "   ${GREEN}export KUBECONFIG=\"$KUBECONFIG_PATH\"${NC}"
echo ""
echo "2. Use it for a single command:"
echo -e "   ${GREEN}KUBECONFIG=\"$KUBECONFIG_PATH\" kubectl get pods${NC}"
echo ""
echo "3. Create an alias for isolated kubectl:"
echo -e "   ${GREEN}alias kubectl-local=\"KUBECONFIG='$KUBECONFIG_PATH' kubectl\"${NC}"
echo ""

# Step 3: Demonstrate usage
echo -e "${YELLOW}Step 3: Testing the isolated kubeconfig...${NC}"
echo ""

# Check current context
echo "Current context in isolated kubeconfig:"
KUBECONFIG="$KUBECONFIG_PATH" kubectl config current-context

echo ""
echo "Available contexts:"
KUBECONFIG="$KUBECONFIG_PATH" kubectl config get-contexts --no-headers -o name

# Step 4: Show how to make changes without affecting the main kubeconfig
echo ""
echo -e "${YELLOW}Step 4: Making changes to the isolated kubeconfig...${NC}"
echo ""
echo "You can now safely switch contexts or namespaces without affecting your main kubeconfig:"
echo ""
echo "Examples:"
echo -e "   ${GREEN}KUBECONFIG=\"$KUBECONFIG_PATH\" kubectl x ctx${NC}  # Switch context"
echo -e "   ${GREEN}KUBECONFIG=\"$KUBECONFIG_PATH\" kubectl x ns${NC}   # Switch namespace"
echo ""

# Optional: Create a convenience function
echo -e "${YELLOW}Optional: Add this function to your shell profile for easy access:${NC}"
echo ""
cat << 'EOF'
kubectl-local() {
    # Get a fresh copy of kubeconfig for isolated work
    local ISOLATED_KUBECONFIG=$(kubectl x kubeconfig copy 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "Using isolated kubeconfig: $ISOLATED_KUBECONFIG" >&2
        export KUBECONFIG="$ISOLATED_KUBECONFIG"
        echo "KUBECONFIG set. Use 'unset KUBECONFIG' to return to default." >&2
    else
        echo "Failed to create isolated kubeconfig" >&2
        return 1
    fi
}
EOF

echo ""
echo -e "${GREEN}=== Setup Complete ===${NC}"
echo ""
echo "Your isolated kubeconfig is ready at:"
echo -e "${GREEN}$KUBECONFIG_PATH${NC}"
echo ""
echo "This file is independent of your main kubeconfig (~/.kube/config)"
echo "and can be modified without affecting your default configuration."