#!/bin/bash
# Demo: Multi-Claude Orchestration
# Shows how one orchestrator could control multiple Claude sessions

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║          Multi-Claude Orchestration Demo                    ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "This demo shows how an orchestrator Claude could:"
echo "  1. Create multiple Claude Code sessions in tmux"
echo "  2. Assign different tasks to each"
echo "  3. Monitor their progress"
echo "  4. Coordinate their work"
echo ""
echo -e "${YELLOW}Note: This is a demonstration. Real orchestration would be done${NC}"
echo -e "${YELLOW}by a Claude session using Desktop Commander's execute_command.${NC}"
echo ""
read -p "Press Enter to begin demo..."

# Clean up any existing demo sessions
echo -e "\n${BLUE}Cleaning up any old demo sessions...${NC}"
tmux kill-session -t orchestrator-demo 2>/dev/null || true
tmux kill-session -t worker-frontend 2>/dev/null || true
tmux kill-session -t worker-backend 2>/dev/null || true
tmux kill-session -t worker-tests 2>/dev/null || true

# Create orchestrator session
echo -e "\n${BLUE}Step 1: Creating orchestrator session...${NC}"
tmux new-session -d -s orchestrator-demo -x 200 -y 50
tmux send-keys -t orchestrator-demo "cd /home/matt/projects/TFE" C-m
tmux send-keys -t orchestrator-demo "clear" C-m
tmux send-keys -t orchestrator-demo "echo 'Orchestrator Claude Session'" C-m
tmux send-keys -t orchestrator-demo "echo '========================'" C-m
tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo 'I can control and monitor worker Claude sessions.'" C-m
sleep 1

# Create worker sessions
echo -e "${BLUE}Step 2: Creating worker Claude sessions...${NC}"

echo -e "  ${GREEN}→${NC} Creating frontend-dev session"
tmux new-session -d -s worker-frontend -x 200 -y 50
tmux send-keys -t worker-frontend "cd /home/matt/projects/TFE" C-m
tmux send-keys -t worker-frontend "clear" C-m
tmux send-keys -t worker-frontend "echo '🎨 Frontend Development Claude'" C-m
tmux send-keys -t worker-frontend "echo 'Ready for frontend tasks...'" C-m

echo -e "  ${GREEN}→${NC} Creating backend-dev session"
tmux new-session -d -s worker-backend -x 200 -y 50
tmux send-keys -t worker-backend "cd /home/matt/projects/TFE" C-m
tmux send-keys -t worker-backend "clear" C-m
tmux send-keys -t worker-backend "echo '⚙️  Backend Development Claude'" C-m
tmux send-keys -t worker-backend "echo 'Ready for backend tasks...'" C-m

echo -e "  ${GREEN}→${NC} Creating test-dev session"
tmux new-session -d -s worker-tests -x 200 -y 50
tmux send-keys -t worker-tests "cd /home/matt/projects/TFE" C-m
tmux send-keys -t worker-tests "clear" C-m
tmux send-keys -t worker-tests "echo '🧪 Testing Claude'" C-m
tmux send-keys -t worker-tests "echo 'Ready for testing tasks...'" C-m

sleep 2

# Orchestrator assigns tasks
echo -e "\n${BLUE}Step 3: Orchestrator assigns tasks to workers...${NC}"

echo -e "  ${CYAN}Sending task to frontend worker...${NC}"
tmux send-keys -t worker-frontend "echo ''" C-m
tmux send-keys -t worker-frontend "echo '📋 Task received from orchestrator:'" C-m
tmux send-keys -t worker-frontend "echo '   Implement user profile UI component'" C-m
tmux send-keys -t worker-frontend "echo '   Location: ui/profile.go'" C-m
tmux send-keys -t worker-frontend "echo '   Status: In progress...'" C-m

echo -e "  ${CYAN}Sending task to backend worker...${NC}"
tmux send-keys -t worker-backend "echo ''" C-m
tmux send-keys -t worker-backend "echo '📋 Task received from orchestrator:'" C-m
tmux send-keys -t worker-backend "echo '   Implement user profile API endpoint'" C-m
tmux send-keys -t worker-backend "echo '   Location: api/profile.go'" C-m
tmux send-keys -t worker-backend "echo '   Status: In progress...'" C-m

echo -e "  ${CYAN}Sending task to test worker...${NC}"
tmux send-keys -t worker-tests "echo ''" C-m
tmux send-keys -t worker-tests "echo '📋 Task received from orchestrator:'" C-m
tmux send-keys -t worker-tests "echo '   Waiting for frontend and backend to complete...'" C-m
tmux send-keys -t worker-tests "echo '   Will test integration when ready'" C-m

sleep 2

# Orchestrator monitors progress
echo -e "\n${BLUE}Step 4: Orchestrator monitors worker progress...${NC}"

echo -e "  ${YELLOW}Capturing frontend output...${NC}"
FRONTEND_OUTPUT=$(tmux capture-pane -t worker-frontend -p -S -10)
echo "$FRONTEND_OUTPUT" | head -5

echo -e "\n  ${YELLOW}Capturing backend output...${NC}"
BACKEND_OUTPUT=$(tmux capture-pane -t worker-backend -p -S -10)
echo "$BACKEND_OUTPUT" | head -5

echo -e "\n  ${YELLOW}Capturing test output...${NC}"
TEST_OUTPUT=$(tmux capture-pane -t worker-tests -p -S -10)
echo "$TEST_OUTPUT" | head -5

sleep 2

# Simulate progress
echo -e "\n${BLUE}Step 5: Workers make progress...${NC}"

tmux send-keys -t worker-frontend "echo '   ✅ Profile component created'" C-m
tmux send-keys -t worker-frontend "echo '   ✅ Styling applied'" C-m
tmux send-keys -t worker-frontend "echo '   Status: Complete!'" C-m

tmux send-keys -t worker-backend "echo '   ✅ GET /api/profile endpoint created'" C-m
tmux send-keys -t worker-backend "echo '   ✅ PUT /api/profile endpoint created'" C-m
tmux send-keys -t worker-backend "echo '   Status: Complete!'" C-m

sleep 2

# Orchestrator coordinates next phase
echo -e "\n${BLUE}Step 6: Orchestrator coordinates integration...${NC}"

echo -e "  ${CYAN}Instructing test worker to begin...${NC}"
tmux send-keys -t worker-tests "echo ''" C-m
tmux send-keys -t worker-tests "echo '📋 Update from orchestrator:'" C-m
tmux send-keys -t worker-tests "echo '   Frontend complete: ui/profile.go'" C-m
tmux send-keys -t worker-tests "echo '   Backend complete: api/profile.go'" C-m
tmux send-keys -t worker-tests "echo '   Beginning integration tests...'" C-m
tmux send-keys -t worker-tests "echo '   ✅ Test 1: Profile UI renders'" C-m
tmux send-keys -t worker-tests "echo '   ✅ Test 2: API endpoint responds'" C-m
tmux send-keys -t worker-tests "echo '   ✅ Test 3: UI integrates with API'" C-m
tmux send-keys -t worker-tests "echo '   Status: All tests passed!'" C-m

sleep 2

# Orchestrator reports final status
echo -e "\n${BLUE}Step 7: Orchestrator synthesizes final report...${NC}"

tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo '═══════════════════════════════════════════'" C-m
tmux send-keys -t orchestrator-demo "echo '🎯 Final Status Report'" C-m
tmux send-keys -t orchestrator-demo "echo '═══════════════════════════════════════════'" C-m
tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo 'Frontend Claude: ✅ Complete'" C-m
tmux send-keys -t orchestrator-demo "echo '  - ui/profile.go created'" C-m
tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo 'Backend Claude: ✅ Complete'" C-m
tmux send-keys -t orchestrator-demo "echo '  - api/profile.go created'" C-m
tmux send-keys -t orchestrator-demo "echo '  - GET and PUT endpoints'" C-m
tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo 'Testing Claude: ✅ Complete'" C-m
tmux send-keys -t orchestrator-demo "echo '  - All integration tests passed'" C-m
tmux send-keys -t orchestrator-demo "echo ''" C-m
tmux send-keys -t orchestrator-demo "echo 'Overall: ✅ Feature complete and tested!'" C-m

sleep 2

# Show all sessions
echo -e "\n${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                   Demo Complete!                             ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "Created tmux sessions:"
echo ""
tmux list-sessions | grep -E "(orchestrator-demo|worker-)" || echo "  No demo sessions found"
echo ""
echo "You can view any session with:"
echo -e "  ${CYAN}tmux attach -t orchestrator-demo${NC}  (main orchestrator)"
echo -e "  ${CYAN}tmux attach -t worker-frontend${NC}    (frontend worker)"
echo -e "  ${CYAN}tmux attach -t worker-backend${NC}     (backend worker)"
echo -e "  ${CYAN}tmux attach -t worker-tests${NC}       (testing worker)"
echo ""
echo "Detach from tmux with: Ctrl+B then D"
echo ""
echo "Clean up demo sessions with:"
echo -e "  ${CYAN}tmux kill-session -t orchestrator-demo${NC}"
echo -e "  ${CYAN}tmux kill-session -t worker-frontend${NC}"
echo -e "  ${CYAN}tmux kill-session -t worker-backend${NC}"
echo -e "  ${CYAN}tmux kill-session -t worker-tests${NC}"
echo ""
echo -e "${YELLOW}In a real scenario:${NC}"
echo -e "${YELLOW}- The orchestrator would be a Claude Code session${NC}"
echo -e "${YELLOW}- It would use Desktop Commander's execute_command${NC}"
echo -e "${YELLOW}- Worker sessions would be actual Claude Code instances${NC}"
echo -e "${YELLOW}- Real code would be written and coordinated${NC}"
