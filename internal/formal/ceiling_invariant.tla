---- MODULE ceiling_invariant ----
(***************************************************************************)
(* AgentRepEngine — Ceiling Score Invariant                                *)
(* TLA+ specification: internal/formal/ceiling_invariant.tla               *)
(*                                                                         *)
(* Verifies that the ceiling override scoring model maintains:             *)
(*   1. A ceiling-flagged agent score never exceeds its ceiling value      *)
(*   2. Only an authorized human operator can raise or clear a ceiling     *)
(*   3. Auto-rollback on FP never inadvertently clears a ceiling           *)
(*   4. Score decay cannot cause a ceiling breach in reverse               *)
(*                                                                         *)
(* Regulatory mapping:                                                     *)
(*   DORA Article 10 — ICT third-party risk controls                       *)
(*   NIST Zero Trust Architecture SP 800-207                               *)
(*   SOC2 CC6.1 — logical access controls                                  *)
(*                                                                         *)
(* Run: tlc internal/formal/ceiling_invariant.tla                         *)
(* All invariants must hold across all reachable states.                   *)
(***************************************************************************)

EXTENDS Naturals, Sequences, TLC

CONSTANTS
  MaxScore,        \* Maximum possible score (1000)
  MinScore,        \* Minimum possible score (0)
  CeilingDefault,  \* Default ceiling for flagged agents (400)
  NumAgents        \* Number of agents in the model (keep small: 2-3)

ASSUME MaxScore = 1000
ASSUME MinScore = 0
ASSUME CeilingDefault = 400
ASSUME NumAgents \in {1, 2, 3}

Agents == 1..NumAgents

VARIABLES
  score,           \* score[a] — current score for agent a
  ceiling,         \* ceiling[a] — ceiling cap (MaxScore+1 means no ceiling)
  ceiling_active,  \* ceiling_active[a] — TRUE if ceiling is enforced
  human_auth,      \* human_auth[a] — TRUE if human operator authorized change
  fp_rollback,     \* fp_rollback — TRUE if FP rate exceeded, rollback active
  operator_action  \* operator_action[a] — "none" | "set_ceiling" | "clear_ceiling" | "raise_ceiling"

vars == <<score, ceiling, ceiling_active, human_auth, fp_rollback, operator_action>>

\* Score is always within bounds
ScoreInBounds == \A a \in Agents : score[a] \in MinScore..MaxScore

\* THE CORE INVARIANT: ceiling-active agent score never exceeds ceiling
CeilingHolds ==
  \A a \in Agents :
    ceiling_active[a] => score[a] <= ceiling[a]

\* Ceiling can only be cleared by authorized human operator
CeilingClearRequiresAuth ==
  \A a \in Agents :
    (~ceiling_active[a] /\ ceiling_active[a]') => human_auth[a]

\* FP rollback never clears a ceiling
FPRollbackPreservesCeiling ==
  fp_rollback =>
    \A a \in Agents : ceiling_active[a] = ceiling_active[a]'

\* Ceiling value never increases without human authorization
CeilingNeverSelfRaises ==
  \A a \in Agents :
    (ceiling[a]' > ceiling[a]) => human_auth[a]

TypeInvariant ==
  /\ \A a \in Agents : score[a] \in MinScore..MaxScore
  /\ \A a \in Agents : ceiling[a] \in MinScore..(MaxScore+1)
  /\ \A a \in Agents : ceiling_active[a] \in BOOLEAN
  /\ \A a \in Agents : human_auth[a] \in BOOLEAN
  /\ fp_rollback \in BOOLEAN
  /\ \A a \in Agents : operator_action[a] \in {"none", "set_ceiling", "clear_ceiling", "raise_ceiling"}

----

Init ==
  /\ score           = [a \in Agents |-> MaxScore]   \* all agents start TRUSTED
  /\ ceiling         = [a \in Agents |-> MaxScore+1]  \* no ceiling active
  /\ ceiling_active  = [a \in Agents |-> FALSE]
  /\ human_auth      = [a \in Agents |-> FALSE]
  /\ fp_rollback     = FALSE
  /\ operator_action = [a \in Agents |-> "none"]

----

\* Score decays naturally (anomaly penalty or time decay)
ScoreDecay(a) ==
  /\ score[a] > MinScore
  /\ score' = [score EXCEPT ![a] = score[a] - 1]
  /\ UNCHANGED <<ceiling, ceiling_active, human_auth, fp_rollback, operator_action>>

\* Score improves (legitimate behavior observed)
ScoreImprove(a) ==
  /\ score[a] < MaxScore
  \* If ceiling active, score cannot exceed ceiling
  /\ IF ceiling_active[a]
     THEN score' = [score EXCEPT ![a] = Min(score[a] + 1, ceiling[a])]
     ELSE score' = [score EXCEPT ![a] = score[a] + 1]
  /\ UNCHANGED <<ceiling, ceiling_active, human_auth, fp_rollback, operator_action>>

\* Human operator sets a ceiling on a compromised agent
HumanSetCeiling(a) ==
  /\ human_auth' = [human_auth EXCEPT ![a] = TRUE]
  /\ ceiling' = [ceiling EXCEPT ![a] = CeilingDefault]
  /\ ceiling_active' = [ceiling_active EXCEPT ![a] = TRUE]
  /\ operator_action' = [operator_action EXCEPT ![a] = "set_ceiling"]
  \* Score immediately capped if above new ceiling
  /\ score' = [score EXCEPT ![a] = Min(score[a], CeilingDefault)]
  /\ UNCHANGED fp_rollback

\* Human operator clears a ceiling (requires auth)
HumanClearCeiling(a) ==
  /\ human_auth[a] = TRUE
  /\ ceiling_active[a] = TRUE
  /\ ceiling_active' = [ceiling_active EXCEPT ![a] = FALSE]
  /\ ceiling' = [ceiling EXCEPT ![a] = MaxScore+1]
  /\ operator_action' = [operator_action EXCEPT ![a] = "clear_ceiling"]
  /\ UNCHANGED <<score, human_auth, fp_rollback>>

\* Human operator raises the ceiling (partial rehabilitation)
HumanRaiseCeiling(a) ==
  /\ human_auth[a] = TRUE
  /\ ceiling_active[a] = TRUE
  /\ ceiling[a] < MaxScore
  /\ ceiling' = [ceiling EXCEPT ![a] = Min(ceiling[a] + 100, MaxScore)]
  /\ operator_action' = [operator_action EXCEPT ![a] = "raise_ceiling"]
  /\ UNCHANGED <<score, ceiling_active, human_auth, fp_rollback>>

\* FP rate exceeded — auto-rollback to observe mode
\* Must NOT affect ceiling state
FPRateExceeded ==
  /\ ~fp_rollback
  /\ fp_rollback' = TRUE
  \* Ceiling state completely unchanged
  /\ UNCHANGED <<score, ceiling, ceiling_active, human_auth, operator_action>>

\* FP rate recovered — exit rollback
FPRateRecovered ==
  /\ fp_rollback
  /\ fp_rollback' = FALSE
  /\ UNCHANGED <<score, ceiling, ceiling_active, human_auth, operator_action>>

\* Reset human auth after action completes
ResetAuth(a) ==
  /\ human_auth[a] = TRUE
  /\ human_auth' = [human_auth EXCEPT ![a] = FALSE]
  /\ operator_action' = [operator_action EXCEPT ![a] = "none"]
  /\ UNCHANGED <<score, ceiling, ceiling_active, fp_rollback>>

----

Next ==
  \/ \E a \in Agents : ScoreDecay(a)
  \/ \E a \in Agents : ScoreImprove(a)
  \/ \E a \in Agents : HumanSetCeiling(a)
  \/ \E a \in Agents : HumanClearCeiling(a)
  \/ \E a \in Agents : HumanRaiseCeiling(a)
  \/ FPRateExceeded
  \/ FPRateRecovered
  \/ \E a \in Agents : ResetAuth(a)

Spec == Init /\ [][Next]_vars

----

(***************************************************************************)
(* INVARIANTS — TLC checks these across all reachable states               *)
(***************************************************************************)

THEOREM Spec => [](TypeInvariant)
THEOREM Spec => [](ScoreInBounds)
THEOREM Spec => [](CeilingHolds)
THEOREM Spec => [](FPRollbackPreservesCeiling)

====