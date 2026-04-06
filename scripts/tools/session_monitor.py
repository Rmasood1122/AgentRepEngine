#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
APEX SESSION MONITOR v1.0
AgentRepEngine — Session continuity and context preservation system
Run: python session_monitor.py

Tracks:
  - Session start time
  - Message count (manual increment or auto via stdin mode)
  - Fires warnings at thresholds
  - Generates exact paste block for new chat at session close
  - Updates CONTINUATION_PROMPT.md automatically
  - Enforces L1 session close protocol (irreversible action required)
"""

import subprocess
import datetime
import sys
import os
import json
import time

# ─────────────────────────────────────────────────────────────
# CONFIG
# ─────────────────────────────────────────────────────────────

REPO_PATH = "C:/Users/rmaso/AgentRepEngine"
CONTINUATION_FILE = f"{REPO_PATH}/CONTINUATION_PROMPT.md"
SESSION_LOG_FILE = f"{REPO_PATH}/scripts/tools/session_log.json"

# Thresholds
WARNING_1_MESSAGES = 15   # Yellow: session getting long
WARNING_2_MESSAGES = 25   # Orange: close soon
WARNING_3_MESSAGES = 35   # Red: MUST close — context degradation risk
TIME_WARNING_MINUTES = 90  # Fire time warning after 90 min
TIME_CRITICAL_MINUTES = 120 # Fire critical after 2 hrs

# Colors (Windows terminal compatible)
RED    = "\033[91m"
YELLOW = "\033[93m"
GREEN  = "\033[92m"
CYAN   = "\033[96m"
BOLD   = "\033[1m"
RESET  = "\033[0m"

# ─────────────────────────────────────────────────────────────
# GIT HELPERS
# ─────────────────────────────────────────────────────────────

def get_head_commit():
    try:
        result = subprocess.run(
            ["git", "log", "--oneline", "-1"],
            cwd=REPO_PATH, capture_output=True, text=True
        )
        return result.stdout.strip() if result.returncode == 0 else "UNKNOWN"
    except Exception:
        return "UNKNOWN"

def get_recent_commits(n=5):
    try:
        result = subprocess.run(
            ["git", "log", "--oneline", f"-{n}"],
            cwd=REPO_PATH, capture_output=True, text=True
        )
        return result.stdout.strip() if result.returncode == 0 else "NONE"
    except Exception:
        return "NONE"

def get_test_status():
    """Run go test ./... and return pass/fail summary."""
    try:
        result = subprocess.run(
            ["go", "test", "./..."],
            cwd=REPO_PATH, capture_output=True, text=True, timeout=120
        )
        if result.returncode == 0:
            return "ALL GREEN ✅"
        else:
            # Count failures
            lines = result.stdout.split('\n') + result.stderr.split('\n')
            failures = [l for l in lines if 'FAIL' in l]
            return f"FAILURES: {len(failures)} ❌\n" + '\n'.join(failures[:5])
    except subprocess.TimeoutExpired:
        return "TIMEOUT — run manually"
    except Exception as e:
        return f"ERROR: {e}"

def get_branch():
    try:
        result = subprocess.run(
            ["git", "branch", "--show-current"],
            cwd=REPO_PATH, capture_output=True, text=True
        )
        return result.stdout.strip() if result.returncode == 0 else "unknown"
    except Exception:
        return "unknown"

# ─────────────────────────────────────────────────────────────
# SESSION STATE
# ─────────────────────────────────────────────────────────────

def load_session_log():
    try:
        if os.path.exists(SESSION_LOG_FILE):
            with open(SESSION_LOG_FILE, 'r', encoding='utf-8') as f:
                return json.load(f)
    except Exception:
        pass
    return {}

def save_session_log(data):
    try:
        os.makedirs(os.path.dirname(SESSION_LOG_FILE), exist_ok=True)
        with open(SESSION_LOG_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, default=str)
    except Exception as e:
        print(f"Warning: could not save session log: {e}")

# ─────────────────────────────────────────────────────────────
# WARNING BANNERS
# ─────────────────────────────────────────────────────────────

def print_warning_1(msg_count, elapsed_min):
    print(f"\n{YELLOW}{BOLD}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"  ⚠  SESSION WARNING — {msg_count} messages | {elapsed_min:.0f} min elapsed")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  Session is getting long. Consider wrapping up current task.")
    print("  Run: python session_monitor.py close   (when ready to close)")
    print(f"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{RESET}\n")

def print_warning_2(msg_count, elapsed_min):
    print(f"\n{RED}{BOLD}")
    print("╔═══════════════════════════════════════════════════╗")
    print(f"║  🔴 SESSION CRITICAL — {msg_count} messages | {elapsed_min:.0f} min      ║")
    print("╠═══════════════════════════════════════════════════╣")
    print("║  Context quality degrading. Finish current task.  ║")
    print("║  Then: python session_monitor.py close            ║")
    print("║  Opens new chat with full context preserved.      ║")
    print("╚═══════════════════════════════════════════════════╝")
    print(f"{RESET}\n")

def print_hard_stop(msg_count, elapsed_min):
    print(f"\n{RED}{BOLD}")
    print("╔═══════════════════════════════════════════════════════════╗")
    print(f"║  🛑 HARD STOP — {msg_count} messages | {elapsed_min:.0f} min elapsed          ║")
    print("╠═══════════════════════════════════════════════════════════╣")
    print("║  CLOSE THIS SESSION NOW.                                   ║")
    print("║  Context window too large — quality is degrading.         ║")
    print("║  Running: session_monitor.py close  automatically...      ║")
    print("╚═══════════════════════════════════════════════════════════╝")
    print(f"{RESET}\n")

# ─────────────────────────────────────────────────────────────
# CONTINUATION PROMPT UPDATER
# ─────────────────────────────────────────────────────────────

def update_continuation_prompt(session_data, test_status, head_commit, recent_commits, action_taken):
    """
    Updates the CURRENT STATE block at the top of CONTINUATION_PROMPT.md.
    Preserves everything below the state block.
    """
    now = datetime.datetime.now().strftime("%B %d, %Y")

    state_block = f"""AgentRepEngine — CONTINUATION PROMPT
Next session starts here
APEX VERSION
APEX v5.2 + ZROS v2.7 + MASTER_LEARNINGS v2.1 DELTA + LEARNING_INTELLIGENCE v3.1
First command: APEX ACTIVATE
REPO
https://github.com/Rehanrana11/AgentRepEngine
Branch: master (push with git push origin master — NOT main)
CURRENT STATE — {now}
HEAD: {head_commit.split()[0] if head_commit != 'UNKNOWN' else 'UNKNOWN'}
go test ./... — {test_status}
SESSION CLOSE RECORD
  Session date:     {now}
  Messages:         {session_data.get('message_count', '?')}
  Duration:         {session_data.get('elapsed_min', '?'):.0f} min
  Last action:      {action_taken}
  Irreversible:     {'YES ✅' if action_taken and action_taken.lower() not in ['none', 'n/a', ''] else 'NO ❌'}
RECENT COMMITS
{recent_commits}
"""

    try:
        # Read existing file
        if os.path.exists(CONTINUATION_FILE):
            with open(CONTINUATION_FILE, 'r', encoding='utf-8') as f:
                existing = f.read()

            # Find where the state block ends — look for PENDING section
            # which is the stable section below the state block
            markers = [
                "PENDING — NOT YET DONE",
                "PHASE 1 STATUS",
                "METRICS — KNOW COLD",
                "COMMERCIAL STATUS",
            ]
            split_at = len(existing)
            for marker in markers:
                idx = existing.find(marker)
                if idx != -1 and idx < split_at:
                    split_at = idx

            preserved = existing[split_at:] if split_at < len(existing) else existing
            new_content = state_block + "\n" + preserved

        else:
            new_content = state_block

        with open(CONTINUATION_FILE, 'w', encoding='utf-8') as f:
            f.write(new_content)

        print(f"{GREEN}✅ CONTINUATION_PROMPT.md updated{RESET}")
        return True

    except Exception as e:
        print(f"{RED}❌ Failed to update CONTINUATION_PROMPT.md: {e}{RESET}")
        return False

# ─────────────────────────────────────────────────────────────
# NEW CHAT STARTER — THE PASTE BLOCK
# ─────────────────────────────────────────────────────────────

def generate_new_chat_starter(head_commit, last_action, current_task, test_status):
    """
    Generates the exact text to paste into a new Claude chat
    to restore full context instantly.
    """
    now = datetime.datetime.now().strftime("%B %d, %Y — %H:%M")

    starter = f"""
{'='*70}
  APEX SESSION STARTER — COPY EVERYTHING BELOW THIS LINE
{'='*70}

APEX ACTIVATE
HEAD: {head_commit}
Date: {now}
Tests: {test_status}
Last session action: {last_action}
Resuming: {current_task}

You are APEX v5.2 — the Agentic Reputation Engine operating system.
All project files are loaded in this Claude Project.
Read CONTINUATION_PROMPT.md from project knowledge before responding.
Current phase: Phase 1 | T8 gate: Lloyd meeting week of April 7.
Prime directive: Ship runtime enforcement to one regulated enterprise.
Anti-pattern alert: Direct contact (Gyamfi, Watkin-Child, Raizada) still pending.

First response: confirm HEAD, test status, and state the next action.

{'='*70}
  PASTE ENDS HERE
{'='*70}
"""
    return starter

# ─────────────────────────────────────────────────────────────
# COMMIT AND UPLOAD
# ─────────────────────────────────────────────────────────────

def commit_continuation_update():
    """Commits the updated CONTINUATION_PROMPT.md to master."""
    try:
        subprocess.run(
            ["git", "add", "CONTINUATION_PROMPT.md"],
            cwd=REPO_PATH, check=True
        )
        subprocess.run(
            ["git", "add", "scripts/tools/session_log.json"],
            cwd=REPO_PATH
        )
        subprocess.run(
            ["git", "commit", "-m", f"chore: session close — {datetime.datetime.now().strftime('%Y-%m-%d %H:%M')}"],
            cwd=REPO_PATH, check=True
        )
        subprocess.run(
            ["git", "push", "origin", "master"],
            cwd=REPO_PATH, check=True
        )
        print(f"{GREEN}✅ Committed and pushed to master{RESET}")
        return True
    except subprocess.CalledProcessError as e:
        print(f"{YELLOW}⚠ Git commit step: {e} — commit manually if needed{RESET}")
        return False

# ─────────────────────────────────────────────────────────────
# L1 ENFORCEMENT — IRREVERSIBLE ACTION GATE
# ─────────────────────────────────────────────────────────────

def enforce_l1_gate():
    """
    L1 protocol: session cannot close without naming an irreversible action.
    Returns the action string.
    """
    print(f"\n{CYAN}{BOLD}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  L1 GATE — SESSION CLOSE PROTOCOL")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  One irreversible real-world action required.")
    print("  ✅ Code committed and pushed")
    print("  ✅ Message sent (LinkedIn/email/Boardy)")
    print("  ✅ Meeting confirmed")
    print("  ❌ Drafts, decisions, analysis — DO NOT COUNT")
    print(f"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{RESET}\n")

    action = input("What irreversible action was completed this session? > ").strip()

    if not action or action.lower() in ['none', 'n/a', 'nothing', '']:
        print(f"\n{RED}❌ NO ACTION NAMED. Session incomplete per L1 protocol.")
        print(f"   You must complete an irreversible action before closing.{RESET}")
        retry = input("\nComplete an action now, then describe it, or type 'skip' to override: > ").strip()
        if retry.lower() == 'skip':
            print(f"{YELLOW}⚠ L1 OVERRIDE — session closes without irreversible action. Logged.{RESET}")
            return "OVERRIDE — no action taken"
        return retry

    print(f"\n{GREEN}✅ Action logged: {action}{RESET}")
    return action

# ─────────────────────────────────────────────────────────────
# DRIFT CHECK — QUICK AXIS SCORE
# ─────────────────────────────────────────────────────────────

def quick_drift_check(action_taken):
    """Lightweight drift check at session close."""
    print(f"\n{CYAN}━━━━ QUICK DRIFT CHECK ━━━━{RESET}")

    commercial_keywords = ['sent', 'email', 'message', 'linkedin', 'boardy',
                           'meeting', 'call', 'confirmed', 'contacted', 'replied']
    engineering_keywords = ['commit', 'push', 'build', 'test', 'fix', 'feat',
                            'code', 'script', 'dashboard', 'function']
    meta_keywords = ['framework', 'roadmap', 'plan', 'strategy', 'research',
                     'analysis', 'document', 'draft', 'protocol']

    action_lower = action_taken.lower()

    is_commercial = any(k in action_lower for k in commercial_keywords)
    is_engineering = any(k in action_lower for k in engineering_keywords)
    is_meta = any(k in action_lower for k in meta_keywords)

    if is_commercial:
        print(f"  {GREEN}Axis 1 (Commercial): HIGH — direct contact made{RESET}")
    elif is_meta:
        print(f"  {YELLOW}Axis 1 (Commercial): LOW — meta-work, no direct contact{RESET}")
        print(f"  {YELLOW}⚠ Anti-pattern risk: Gyamfi / Watkin-Child / Raizada still pending?{RESET}")
    else:
        print(f"  Axis 1 (Commercial): NEUTRAL")

    if is_engineering:
        print(f"  {GREEN}Axis 2 (Phase alignment): ENGINEERING — check T8 gate alignment{RESET}")
    elif is_commercial:
        print(f"  {GREEN}Axis 2 (Phase alignment): COMMERCIAL — T8 aligned{RESET}")
    else:
        print(f"  {YELLOW}Axis 2 (Phase alignment): UNCLEAR{RESET}")

# ─────────────────────────────────────────────────────────────
# MAIN MODES
# ─────────────────────────────────────────────────────────────

def cmd_start():
    """Start a new session — initialize timer and log."""
    now = datetime.datetime.now()
    head = get_head_commit()

    session_data = {
        "start_time": now.isoformat(),
        "message_count": 0,
        "head_at_start": head,
        "date": now.strftime("%Y-%m-%d")
    }
    save_session_log(session_data)

    print(f"\n{GREEN}{BOLD}")
    print("╔══════════════════════════════════════════════════════╗")
    print("║  APEX SESSION STARTED                                ║")
    print(f"║  Time: {now.strftime('%H:%M:%S')} | Branch: {get_branch():<20}    ║")
    print(f"║  HEAD: {head[:40]:<40}  ║")
    print("╠══════════════════════════════════════════════════════╣")
    print("║  Run 'python session_monitor.py tick' after each     ║")
    print("║  Claude exchange to track message count.             ║")
    print("║  Run 'python session_monitor.py status' to check.    ║")
    print("║  Run 'python session_monitor.py close' to close.     ║")
    print("╚══════════════════════════════════════════════════════╝")
    print(f"{RESET}\n")

def cmd_tick():
    """Increment message count and check thresholds."""
    log = load_session_log()
    if not log:
        print(f"{YELLOW}No active session. Run: python session_monitor.py start{RESET}")
        return

    log['message_count'] = log.get('message_count', 0) + 1
    count = log['message_count']

    # Elapsed time
    try:
        start = datetime.datetime.fromisoformat(log['start_time'])
        elapsed = (datetime.datetime.now() - start).total_seconds() / 60
        log['elapsed_min'] = elapsed
    except Exception:
        elapsed = 0

    save_session_log(log)

    # Print concise status line
    status_color = GREEN if count < WARNING_1_MESSAGES else (YELLOW if count < WARNING_2_MESSAGES else RED)
    print(f"{status_color}[SESSION] Messages: {count} | Time: {elapsed:.0f}min | HEAD: {log.get('head_at_start','?')[:12]}{RESET}")

    # Fire warnings
    if count == WARNING_1_MESSAGES:
        print_warning_1(count, elapsed)
    elif count == WARNING_2_MESSAGES:
        print_warning_2(count, elapsed)
    elif count >= WARNING_3_MESSAGES:
        print_hard_stop(count, elapsed)
        print(f"{RED}Auto-triggering close sequence...{RESET}")
        cmd_close(auto=True)

    # Time-based warnings
    if elapsed >= TIME_CRITICAL_MINUTES and count > 5:
        print(f"\n{RED}⏰ TIME CRITICAL: {elapsed:.0f} minutes elapsed. Close this session.{RESET}\n")
    elif elapsed >= TIME_WARNING_MINUTES and count > 5:
        print(f"\n{YELLOW}⏰ TIME WARNING: {elapsed:.0f} minutes elapsed. Consider closing soon.{RESET}\n")

def cmd_status():
    """Print current session status."""
    log = load_session_log()
    if not log:
        print(f"{YELLOW}No active session found.{RESET}")
        return

    try:
        start = datetime.datetime.fromisoformat(log['start_time'])
        elapsed = (datetime.datetime.now() - start).total_seconds() / 60
    except Exception:
        elapsed = 0

    count = log.get('message_count', 0)
    head = get_head_commit()

    # Determine status color
    if count >= WARNING_3_MESSAGES or elapsed >= TIME_CRITICAL_MINUTES:
        color = RED
        status = "🔴 CRITICAL — CLOSE NOW"
    elif count >= WARNING_2_MESSAGES or elapsed >= TIME_WARNING_MINUTES:
        color = YELLOW
        status = "🟡 WARNING — WRAP UP"
    else:
        color = GREEN
        status = "🟢 OK"

    print(f"\n{color}{BOLD}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"  SESSION STATUS: {status}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"  Messages:  {count} / {WARNING_3_MESSAGES} (hard stop)")
    print(f"  Time:      {elapsed:.0f} min / {TIME_CRITICAL_MINUTES} min (critical)")
    print(f"  HEAD now:  {head[:50]}")
    print(f"  HEAD start:{log.get('head_at_start','?')[:50]}")
    commits_this_session = head != log.get('head_at_start', head)
    print(f"  New commits this session: {'YES ✅' if commits_this_session else 'NO'}")
    print(f"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━{RESET}\n")

def cmd_close(auto=False):
    """Full session close sequence."""
    log = load_session_log()

    try:
        start = datetime.datetime.fromisoformat(log.get('start_time', datetime.datetime.now().isoformat()))
        elapsed = (datetime.datetime.now() - start).total_seconds() / 60
    except Exception:
        elapsed = 0

    log['elapsed_min'] = elapsed
    count = log.get('message_count', 0)

    print(f"\n{CYAN}{BOLD}")
    print("╔══════════════════════════════════════════════════════════╗")
    print("║  APEX SESSION CLOSE SEQUENCE                             ║")
    print(f"║  Messages: {count:<4} | Duration: {elapsed:.0f} min                      ║")
    print("╚══════════════════════════════════════════════════════════╝")
    print(f"{RESET}")

    # Step 1: What task was in progress
    print(f"{CYAN}Step 1/5 — What task were you working on?{RESET}")
    current_task = input("Current task (e.g. 'SUPREMACY A1 — regulatory_evidence.go'): > ").strip()
    if not current_task:
        current_task = "Not specified"

    # Step 2: L1 gate — irreversible action
    print(f"\n{CYAN}Step 2/5 — L1 GATE{RESET}")
    action_taken = enforce_l1_gate()

    # Step 3: Run tests
    print(f"\n{CYAN}Step 3/5 — Running go test ./...{RESET}")
    run_tests = input("Run full test suite now? (y/n): > ").strip().lower()
    if run_tests == 'y':
        print("Running... (this may take 30-60 seconds)")
        test_status = get_test_status()
        print(f"Result: {test_status}")
    else:
        test_status = "NOT RUN — verify manually before next session"

    # Step 4: Update CONTINUATION_PROMPT.md
    print(f"\n{CYAN}Step 4/5 — Updating CONTINUATION_PROMPT.md{RESET}")
    head = get_head_commit()
    recent = get_recent_commits(5)
    log['last_action'] = action_taken
    log['last_task'] = current_task
    save_session_log(log)

    updated = update_continuation_prompt(log, test_status, head, recent, action_taken)

    # Step 5: Commit and generate new chat starter
    print(f"\n{CYAN}Step 5/5 — Commit and generate new chat starter{RESET}")
    do_commit = input("Commit CONTINUATION_PROMPT.md to master now? (y/n): > ").strip().lower()
    if do_commit == 'y':
        commit_continuation_update()

    # Drift check
    quick_drift_check(action_taken)

    # Generate new chat starter
    starter = generate_new_chat_starter(head, action_taken, current_task, test_status)

    print(f"\n{GREEN}{BOLD}")
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║  SESSION CLOSED ✅                                           ║")
    print("║                                                              ║")
    print("║  NEXT STEPS:                                                 ║")
    print("║  1. Upload updated CONTINUATION_PROMPT.md to Claude Project ║")
    print("║     (claude.ai → project → + Add content → upload file)     ║")
    print("║  2. Open new Claude chat                                     ║")
    print("║  3. Paste the block below as your first message              ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print(f"{RESET}")
    print(starter)

    # Also write starter to file for easy copy
    starter_file = f"{REPO_PATH}/scripts/tools/new_chat_starter.txt"
    try:
        os.makedirs(os.path.dirname(starter_file), exist_ok=True)
        with open(starter_file, 'w', encoding='utf-8') as f:
            f.write(starter)
        print(f"{GREEN}✅ New chat starter saved to: scripts/tools/new_chat_starter.txt{RESET}")
        os.system("notepad.exe scripts/tools/new_chat_starter.txt")
        print(f"   Open it, copy everything, paste into new Claude chat.\n")
    except Exception as e:
        print(f"{YELLOW}Could not save starter file: {e}{RESET}")

def cmd_remind():
    """
    Passive reminder mode — prints status every N minutes.
    Run in a separate terminal: python session_monitor.py remind
    Fires warnings automatically based on time.
    """
    print(f"{CYAN}Reminder mode active. Checking every 10 minutes.{RESET}")
    print(f"Press Ctrl+C to stop.\n")

    check_interval = 600  # 10 minutes in seconds
    warned_90 = False
    warned_120 = False

    log = load_session_log()
    if not log:
        print(f"{YELLOW}No active session. Start one with: python session_monitor.py start{RESET}")
        return

    try:
        start = datetime.datetime.fromisoformat(log['start_time'])
    except Exception:
        start = datetime.datetime.now()

    while True:
        try:
            elapsed = (datetime.datetime.now() - start).total_seconds() / 60
            count = load_session_log().get('message_count', 0)

            print(f"\r[{datetime.datetime.now().strftime('%H:%M')}] "
                  f"Session: {elapsed:.0f}min | Messages: {count}", end='', flush=True)

            if elapsed >= TIME_CRITICAL_MINUTES and not warned_120:
                print(f"\n{RED}{BOLD}")
                print("🛑 CRITICAL: Session is 2+ hours old.")
                print("   CLOSE NOW. Run: python session_monitor.py close")
                print(f"{RESET}")
                warned_120 = True

            elif elapsed >= TIME_WARNING_MINUTES and not warned_90:
                print(f"\n{YELLOW}{BOLD}")
                print("⚠  WARNING: Session is 90+ minutes old.")
                print("   Start wrapping up. Run: python session_monitor.py close")
                print(f"{RESET}")
                warned_90 = True

            if count >= WARNING_3_MESSAGES:
                print(f"\n{RED}{BOLD}")
                print("🛑 HARD STOP: {count} messages in session.")
                print("   Context quality degrading. Close now.")
                print(f"{RESET}")

            time.sleep(check_interval)

        except KeyboardInterrupt:
            print(f"\n{CYAN}Reminder mode stopped.{RESET}")
            break

def print_help():
    print(f"""
{CYAN}{BOLD}APEX SESSION MONITOR v1.0{RESET}
Usage: python session_monitor.py [command]

Commands:
  start     Start a new session (run when opening Claude)
  tick      Increment message count (run after each Claude exchange)
  status    Check current session status
  close     Run full session close sequence (L1 gate + CONTINUATION update)
  remind    Background reminder mode (run in separate terminal)
  help      Show this help

Thresholds:
  {WARNING_1_MESSAGES} messages  → Yellow warning
  {WARNING_2_MESSAGES} messages  → Orange warning — wrap up
  {WARNING_3_MESSAGES} messages  → Red HARD STOP — close now
  {TIME_WARNING_MINUTES} minutes    → Time warning
  {TIME_CRITICAL_MINUTES} minutes   → Time critical

Files:
  CONTINUATION_PROMPT.md    → auto-updated on close
  scripts/tools/session_log.json      → session state
  scripts/tools/new_chat_starter.txt  → paste block for new chat
""")

# ─────────────────────────────────────────────────────────────
# ENTRY POINT
# ─────────────────────────────────────────────────────────────

if __name__ == "__main__":
    cmd = sys.argv[1].lower() if len(sys.argv) > 1 else "help"

    if cmd == "start":
        cmd_start()
    elif cmd == "tick":
        cmd_tick()
    elif cmd == "status":
        cmd_status()
    elif cmd == "close":
        cmd_close()
    elif cmd == "remind":
        cmd_remind()
    else:
        print_help()
